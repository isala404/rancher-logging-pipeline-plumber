package controllers

import (
	"bytes"
	"context"
	"fmt"
	flowv1beta1 "github.com/banzaicloud/logging-operator/pkg/sdk/api/v1beta1"
	loggingplumberv1alpha1 "github.com/mrsupiri/rancher-logging-explorer/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *FlowTestReconciler) provisionResource(ctx context.Context) error {
	logger := log.FromContext(ctx)
	flowTest := ctx.Value("flowTest").(loggingplumberv1alpha1.FlowTest)

	logOutput := new(bytes.Buffer)
	for _, line := range flowTest.Spec.SentMessages {
		_, _ = logOutput.WriteString(fmt.Sprintf("%s\n", line))
	}

	Immutable := true
	configMap := v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "V1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-configmap", flowTest.ObjectMeta.UID),
			Namespace: flowTest.Spec.ReferencePod.Namespace,
			Labels:    GetLabels("pod-simulation", &flowTest),
		},
		Immutable:  &Immutable,
		BinaryData: map[string][]byte{"simulation.log": logOutput.Bytes()},
	}

	if err := r.Create(ctx, &configMap); err != nil {
		logger.Error(err, "failed to create ConfigMap with simulation.log")
		return err
	}

	logger.V(1).Info("deployed config map with simulation.log", "uuid", configMap.ObjectMeta.UID)

	var referencePod v1.Pod
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: flowTest.Spec.ReferencePod.Namespace,
		Name:      flowTest.Spec.ReferencePod.Name,
	}, &referencePod); err != nil {
		return err
	}

	simulationPod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "V1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-simulation", flowTest.UID),
			Namespace: flowTest.Spec.ReferencePod.Namespace,
			Labels:    map[string]string{},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				// TODO: Handle more than or less than 1 Container (#12)
				Name:         referencePod.Spec.Containers[0].Name,
				Image:        "k3d-rancher-logging-explorer-registry:5000/rancher-logging-explorer/pod-simulator:latest",
				VolumeMounts: []v1.VolumeMount{{Name: "config-volume", MountPath: "/var/logs"}},
			}},
			Volumes: []v1.Volume{
				{
					Name: "config-volume",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{Name: fmt.Sprintf("%s-configmap", flowTest.ObjectMeta.UID)},
						},
					},
				},
			},
			NodeSelector: referencePod.Spec.NodeSelector,
		},
	}

	extraLabels := GetLabels("pod-simulation", &flowTest)

	simulationPod.ObjectMeta.Labels = referencePod.ObjectMeta.Labels

	for k, v := range extraLabels {
		simulationPod.ObjectMeta.Labels[k] = v
	}

	if err := r.Create(ctx, &simulationPod); err != nil {
		logger.Error(err, "failed to create the simulation pod")
		return err
	}

	logger.V(1).Info("deployed simulation pod", "pod-uuid", simulationPod.UID)

	var outputPod v1.Pod
	if err := r.Get(ctx, client.ObjectKey{Name: "logging-plumber-log-aggregator", Namespace: flowTest.ObjectMeta.Namespace}, &outputPod); err != nil {
		if apierrors.IsNotFound(err) {
			outputPod := v1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "V1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "logging-plumber-log-aggregator",
					Namespace: flowTest.ObjectMeta.Namespace,
					Labels: GetLabels("logging-plumber-log-aggregator", nil,
						map[string]string{"loggingplumber.isala.me/component": "log-aggregator"}),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Name:  "log-output",
						Image: "paynejacob/log-output:latest",
						Ports: []v1.ContainerPort{{
							Name:          "http",
							ContainerPort: 80,
							Protocol:      "TCP",
						}},
					}},
				},
			}
			if err := r.Create(ctx, &outputPod); err != nil {
				logger.Error(err, "failed to create the log output pod")
				return err
			}
			logger.V(1).Info("deployed log output pod", "pod-uuid", outputPod.UID)
		}
	} else {
		logger.V(1).Info("found a already deployed log output pod", "pod-uuid", outputPod.UID)
	}

	matches, clusterMatches, filters, err := r.deploySlicedFlows(ctx, extraLabels)

	if err != nil {
		return err
	}

	flowTest.Status.Status = loggingplumberv1alpha1.Running

	flowTest.Status.FailedFilters = filters
	flowTest.Status.FailedMatches = matches
	flowTest.Status.FailedClusterMatches = clusterMatches

	if err := r.Status().Update(ctx, &flowTest); err != nil {
		logger.Error(err, "failed to update flowtest status")
		return err
	}

	return nil
}

func (r *FlowTestReconciler) deploySlicedFlows(ctx context.Context, extraLabels map[string]string) (matches []flowv1beta1.Match, clusterMatches []flowv1beta1.ClusterMatch, filters []flowv1beta1.Filter, err error) {
	logger := log.FromContext(ctx)
	flowTest := ctx.Value("flowTest").(loggingplumberv1alpha1.FlowTest)

	// TODO: handle this sane way
	if flowTest.Spec.ReferenceFlow.Kind == "ClusterFlow" {
		var referenceFlow flowv1beta1.ClusterFlow
		if err = r.Get(ctx, types.NamespacedName{
			Namespace: flowTest.Spec.ReferenceFlow.Namespace,
			Name:      flowTest.Spec.ReferenceFlow.Name,
		}, &referenceFlow); err != nil {
			return
		}

		clusterMatches = referenceFlow.Spec.Match
		filters = referenceFlow.Spec.Filters

		i := 0
		flowTemplate, outTemplate := clusterFlowTemplates(referenceFlow, flowTest, extraLabels)

		for x := 1; x <= len(referenceFlow.Spec.Match); x++ {
			targetFlow := *flowTemplate.DeepCopy()
			targetOutput := *outTemplate.DeepCopy()

			targetFlow.ObjectMeta.Name = fmt.Sprintf("%s-%d-match", flowTest.ObjectMeta.UID, i)
			targetFlow.ObjectMeta.Labels["loggingplumber.isala.me/test-id"] = fmt.Sprintf("%d", i)
			targetFlow.ObjectMeta.Labels["loggingplumber.isala.me/test-type"] = "match"

			targetOutput.ObjectMeta.Name = fmt.Sprintf("%s-%d-match", flowTest.ObjectMeta.UID, i)
			targetOutput.ObjectMeta.Labels["loggingplumber.isala.me/test-id"] = fmt.Sprintf("%d", i)
			targetOutput.ObjectMeta.Labels["loggingplumber.isala.me/test-type"] = "match"
			targetOutput.Spec.HTTPOutput.Endpoint = fmt.Sprintf("%s/%s/", targetOutput.Spec.HTTPOutput.Endpoint, targetFlow.ObjectMeta.Name)

			targetFlow.Spec.GlobalOutputRefs = []string{targetOutput.ObjectMeta.Name}

			targetFlow.Spec.Match = append(targetFlow.Spec.Match, referenceFlow.Spec.Match[:x]...)

			if err = r.Create(ctx, &targetOutput); err != nil {
				logger.Error(err, fmt.Sprintf("failed to deploy Flow #%d for %s", i, referenceFlow.ObjectMeta.Name))
				return
			}

			if err = r.Create(ctx, &targetFlow); err != nil {
				logger.Error(err, fmt.Sprintf("failed to deploy Flow #%d for %s", i, referenceFlow.ObjectMeta.Name))
				return
			}

			logger.V(1).Info("deployed match slice", "test-id", i)
			i++
		}

		for x := 1; x <= len(referenceFlow.Spec.Filters); x++ {
			targetFlow := *flowTemplate.DeepCopy()
			targetOutput := *outTemplate.DeepCopy()

			targetFlow.ObjectMeta.Name = fmt.Sprintf("%s-%d-filture", flowTest.ObjectMeta.UID, i)
			targetFlow.ObjectMeta.Labels["loggingplumber.isala.me/test-type"] = "filter"
			targetFlow.ObjectMeta.Labels["loggingplumber.isala.me/test-id"] = fmt.Sprintf("%d", i)

			targetOutput.ObjectMeta.Name = fmt.Sprintf("%s-%d-filture", flowTest.ObjectMeta.UID, i)
			targetOutput.ObjectMeta.Labels["loggingplumber.isala.me/test-type"] = "filter"
			targetOutput.ObjectMeta.Labels["loggingplumber.isala.me/test-id"] = fmt.Sprintf("%d", i)
			targetOutput.Spec.HTTPOutput.Endpoint = fmt.Sprintf("%s/%s/", targetOutput.Spec.HTTPOutput.Endpoint, targetFlow.ObjectMeta.Name)

			targetFlow.Spec.GlobalOutputRefs = []string{targetOutput.ObjectMeta.Name}

			targetFlow.Spec.Match = nil

			targetFlow.Spec.Filters = referenceFlow.Spec.Filters[:x]

			if err = r.Create(ctx, &targetOutput); err != nil {
				logger.Error(err, fmt.Sprintf("failed to deploy Flow #%d for %s", i, referenceFlow.ObjectMeta.Name))
				return
			}

			if err = r.Create(ctx, &targetFlow); err != nil {
				logger.Error(err, fmt.Sprintf("failed to deploy Flow #%d for %s", i, referenceFlow.ObjectMeta.Name))
				return
			}
			logger.V(1).Info("deployed filter slice", "test-id", i)
			i++
		}

	} else {
		var referenceFlow flowv1beta1.Flow
		if err = r.Get(ctx, types.NamespacedName{
			Namespace: flowTest.Spec.ReferenceFlow.Namespace,
			Name:      flowTest.Spec.ReferenceFlow.Name,
		}, &referenceFlow); err != nil {
			return
		}

		matches = referenceFlow.Spec.Match
		filters = referenceFlow.Spec.Filters

		i := 0
		flowTemplate, outTemplate := flowTemplates(referenceFlow, flowTest, extraLabels)

		for x := 1; x <= len(referenceFlow.Spec.Match); x++ {
			targetFlow := *flowTemplate.DeepCopy()
			targetOutput := *outTemplate.DeepCopy()

			targetFlow.ObjectMeta.Name = fmt.Sprintf("%s-%d-match", flowTest.ObjectMeta.UID, i)
			targetFlow.ObjectMeta.Labels["loggingplumber.isala.me/test-id"] = fmt.Sprintf("%d", i)
			targetFlow.ObjectMeta.Labels["loggingplumber.isala.me/test-type"] = "match"

			targetOutput.ObjectMeta.Name = fmt.Sprintf("%s-%d-match", flowTest.ObjectMeta.UID, i)
			targetOutput.ObjectMeta.Labels["loggingplumber.isala.me/test-id"] = fmt.Sprintf("%d", i)
			targetOutput.ObjectMeta.Labels["loggingplumber.isala.me/test-type"] = "match"
			targetOutput.Spec.HTTPOutput.Endpoint = fmt.Sprintf("%s/%s/", targetOutput.Spec.HTTPOutput.Endpoint, targetFlow.ObjectMeta.Name)

			targetFlow.Spec.LocalOutputRefs = []string{targetOutput.ObjectMeta.Name}

			targetFlow.Spec.Match = append(targetFlow.Spec.Match, referenceFlow.Spec.Match[:x]...)

			if err = r.Create(ctx, &targetOutput); err != nil {
				logger.Error(err, fmt.Sprintf("failed to deploy Flow #%d for %s", i, referenceFlow.ObjectMeta.Name))
				return
			}

			if err = r.Create(ctx, &targetFlow); err != nil {
				logger.Error(err, fmt.Sprintf("failed to deploy Flow #%d for %s", i, referenceFlow.ObjectMeta.Name))
				return
			}

			logger.V(1).Info("deployed match slice", "test-id", i)
			i++
		}

		for x := 1; x <= len(referenceFlow.Spec.Filters); x++ {
			targetFlow := *flowTemplate.DeepCopy()
			targetOutput := *outTemplate.DeepCopy()

			targetFlow.ObjectMeta.Name = fmt.Sprintf("%s-%d-filture", flowTest.ObjectMeta.UID, i)
			targetFlow.ObjectMeta.Labels["loggingplumber.isala.me/test-type"] = "filter"
			targetFlow.ObjectMeta.Labels["loggingplumber.isala.me/test-id"] = fmt.Sprintf("%d", i)

			targetOutput.ObjectMeta.Name = fmt.Sprintf("%s-%d-filture", flowTest.ObjectMeta.UID, i)
			targetOutput.ObjectMeta.Labels["loggingplumber.isala.me/test-type"] = "filter"
			targetOutput.ObjectMeta.Labels["loggingplumber.isala.me/test-id"] = fmt.Sprintf("%d", i)
			targetOutput.Spec.HTTPOutput.Endpoint = fmt.Sprintf("%s/%s/", targetOutput.Spec.HTTPOutput.Endpoint, targetFlow.ObjectMeta.Name)

			targetFlow.Spec.LocalOutputRefs = []string{targetOutput.ObjectMeta.Name}

			targetFlow.Spec.Match = nil

			targetFlow.Spec.Filters = referenceFlow.Spec.Filters[:x]

			if err = r.Create(ctx, &targetOutput); err != nil {
				logger.Error(err, fmt.Sprintf("failed to deploy Flow #%d for %s", i, referenceFlow.ObjectMeta.Name))
				return
			}

			if err = r.Create(ctx, &targetFlow); err != nil {
				logger.Error(err, fmt.Sprintf("failed to deploy Flow #%d for %s", i, referenceFlow.ObjectMeta.Name))
				return
			}
			logger.V(1).Info("deployed filter slice", "test-id", i)
			i++
		}
	}

	return
}
