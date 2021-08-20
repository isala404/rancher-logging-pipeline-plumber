package controllers

import (
	"bytes"
	"context"
	"fmt"

	flowv1beta1 "github.com/banzaicloud/logging-operator/pkg/sdk/api/v1beta1"
	loggingplumberv1alpha1 "github.com/mrsupiri/logging-pipeline-plumber/pkg/sdk/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
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
				Name:            referencePod.Spec.Containers[0].Name,
				Image:           fmt.Sprintf("%s:%s", r.PodSimulatorImage.Repository, r.PodSimulatorImage.Tag),
				ImagePullPolicy: v1.PullPolicy(r.PodSimulatorImage.PullPolicy),
				Command:         []string{"pod-simulator"},
				Args:            []string{"-log_file", "/simulation.log"},
				VolumeMounts:    []v1.VolumeMount{{Name: "config-volume", MountPath: "/simulation.log", SubPath: "simulation.log"}},
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

	if referencePod.ObjectMeta.Labels != nil {
		simulationPod.ObjectMeta.Labels = referencePod.ObjectMeta.Labels
	} else {
		simulationPod.ObjectMeta.Labels = make(map[string]string)
	}

	for k, v := range extraLabels {
		simulationPod.ObjectMeta.Labels[k] = v
	}

	if err := r.Create(ctx, &simulationPod); err != nil {
		logger.Error(err, "failed to create the simulation pod")
		return err
	}

	logger.V(1).Info("deployed simulation pod", "pod-uuid", simulationPod.UID)

	if err := r.provisionOutputResource(ctx); err != nil {
		return err
	}

	if err := r.deploySlicedFlows(ctx, extraLabels, &flowTest); err != nil {
		return err
	}

	flowTest.Status.Status = loggingplumberv1alpha1.Running

	if err := r.Status().Update(ctx, &flowTest); err != nil {
		logger.Error(err, "failed to update flowtest status")
		return err
	}

	return nil
}

func (r *FlowTestReconciler) deploySlicedFlows(ctx context.Context, extraLabels map[string]string, flowTest *loggingplumberv1alpha1.FlowTest) (err error) {
	logger := log.FromContext(ctx)

	// TODO: handle this sane way
	if flowTest.Spec.ReferenceFlow.Kind == "ClusterFlow" {
		var referenceFlow flowv1beta1.ClusterFlow
		if err = r.Get(ctx, types.NamespacedName{
			Namespace: flowTest.Spec.ReferenceFlow.Namespace,
			Name:      flowTest.Spec.ReferenceFlow.Name,
		}, &referenceFlow); err != nil {
			return
		}

		flowTest.Status.MatchStatus = make([]bool, len(referenceFlow.Spec.Match))
		flowTest.Status.FilterStatus = make([]bool, len(referenceFlow.Spec.Filters))

		i := 0
		flowTemplate, outTemplate := r.clusterFlowTemplates(referenceFlow, *flowTest)
		for x := 0; x <= len(referenceFlow.Spec.Match)-1; x++ {
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

			targetFlow.Spec.Match = []flowv1beta1.ClusterMatch{referenceFlow.Spec.Match[x]}

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

			// ensure logs are only coming from our simulation pod
			targetFlow.Spec.Match = []flowv1beta1.ClusterMatch{{
				ClusterSelect: &flowv1beta1.ClusterSelect{Labels: extraLabels},
			}}

			targetFlow.Spec.Filters = append(targetFlow.Spec.Filters, referenceFlow.Spec.Filters[:x]...)

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

		flowTest.Status.MatchStatus = make([]bool, len(referenceFlow.Spec.Match))
		flowTest.Status.FilterStatus = make([]bool, len(referenceFlow.Spec.Filters))

		i := 0
		flowTemplate, outTemplate := r.flowTemplates(referenceFlow, *flowTest)

		for x := 0; x <= len(referenceFlow.Spec.Match)-1; x++ {
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

			targetFlow.Spec.Match = []flowv1beta1.Match{referenceFlow.Spec.Match[x]}

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

			// ensure logs are only coming from our simulation pod
			targetFlow.Spec.Match = []flowv1beta1.Match{{
				Select: &flowv1beta1.Select{Labels: extraLabels},
			}}

			targetFlow.Spec.Filters = append(targetFlow.Spec.Filters, referenceFlow.Spec.Filters[:x]...)

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

func (r *FlowTestReconciler) provisionOutputResource(ctx context.Context) error {
	logger := log.FromContext(ctx)
	flowTest := ctx.Value("flowTest").(loggingplumberv1alpha1.FlowTest)

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
					Namespace: r.AggregatorNamespace,
					Labels: GetLabels("logging-plumber-log-aggregator", nil,
						map[string]string{"loggingplumber.isala.me/component": "log-aggregator"}),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Name:            "log-output",
						Image:           fmt.Sprintf("%s:%s", r.LogOutputImage.Repository, r.LogOutputImage.Tag),
						ImagePullPolicy: v1.PullPolicy(r.LogOutputImage.PullPolicy),
						Ports: []v1.ContainerPort{{
							Name:          "http",
							ContainerPort: 80,
							Protocol:      "TCP",
						}},
					}},
				},
			}
			if err := r.Create(ctx, &outputPod); err != nil {
				if apierrors.IsAlreadyExists(err) {
					logger.V(1).Info("found a already deployed log output pod")
				} else {
					logger.Error(err, "failed to create the output pod pod")
					return err
				}
			}
			logger.V(1).Info("deployed log output pod", "pod-uuid", outputPod.UID)

			outputPodSVC := v1.Service{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "V1",
					Kind:       "Service",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "logging-plumber-log-aggregator",
					Namespace: r.AggregatorNamespace,
					Labels: GetLabels("logging-plumber-log-aggregator", nil,
						map[string]string{"loggingplumber.isala.me/component": "log-aggregator"}),
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{{
						Name:       "http",
						Protocol:   "TCP",
						Port:       80,
						TargetPort: intstr.IntOrString{Type: intstr.String, StrVal: "http"},
					}},
					Selector: outputPod.Labels,
				},
			}

			if err := r.Create(ctx, &outputPodSVC); err != nil {
				if apierrors.IsAlreadyExists(err) {
					logger.V(1).Info("found a already deployed log output service")
				} else {
					logger.Error(err, "failed to create the output pod service")
					return err
				}
			}

			logger.V(1).Info("deployed output pod service", "service-uuid", outputPodSVC.UID)
		}
	} else {
		logger.V(1).Info("found a already deployed log output pod", "pod-uuid", outputPod.UID)
	}

	return nil
}
