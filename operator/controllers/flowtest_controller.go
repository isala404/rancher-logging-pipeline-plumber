/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"bytes"
	"context"
	"fmt"
	loggingplumberv1alpha1 "github.com/mrsupiri/rancher-logging-explorer/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// FlowTestReconciler reconciles a FlowTest object
type FlowTestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=loggingplumber.isala.me,resources=flowtests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=loggingplumber.isala.me,resources=flowtests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=loggingplumber.isala.me,resources=flowtests/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the FlowTest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *FlowTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Reconciling")

	var flowTest loggingplumberv1alpha1.FlowTest
	if err := r.Get(ctx, req.NamespacedName, &flowTest); err != nil {
		return ctrl.Result{Requeue: false}, client.IgnoreNotFound(err)
	}

	if flowTest.Status.Status == "" {
		flowTest.Status.Status = loggingplumberv1alpha1.Created
		if err := r.Status().Update(ctx, &flowTest); err != nil {
			logger.Error(err, "failed to update flowtest status")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		return ctrl.Result{}, nil
	}

	if flowTest.Status.Status == loggingplumberv1alpha1.Created {
		if err := r.provisionResource(ctx, flowTest); err != nil {
			return ctrl.Result{Requeue: false}, client.IgnoreNotFound(err)
		}
	}

	//var referenceFlow flowv1beta1.Flow
	//if err := r.Get(ctx, types.NamespacedName{
	//	Namespace: flowTest.Spec.ReferenceFlow.Namespace,
	//	Name:      flowTest.Spec.ReferenceFlow.Name,
	//}, &referenceFlow); err != nil {
	//	return ctrl.Result{Requeuefalse}, err
	//}

	return ctrl.Result{Requeue: false}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FlowTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&loggingplumberv1alpha1.FlowTest{}).
		Complete(r)
}

func (r *FlowTestReconciler) provisionResource(ctx context.Context, flowTest loggingplumberv1alpha1.FlowTest) error {
	logger := log.FromContext(ctx)

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
			Name:      fmt.Sprintf("%s-configmap", flowTest.Spec.ReferencePod.Name),
			Namespace: flowTest.Spec.ReferencePod.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":                "pod-simulation",
				"app.kubernetes.io/managed-by":          "rancher-logging-explorer",
				"app.kubernetes.io/created-by":          "logging-plumber",
				"loggingplumber.isala.me/flowtest-uuid": string(flowTest.ObjectMeta.UID),
				"loggingplumber.isala.me/flowtest":      flowTest.ObjectMeta.Name,
			},
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
			Name:      fmt.Sprintf("%s-simulation", referencePod.ObjectMeta.Name),
			Namespace: flowTest.Spec.ReferencePod.Namespace,
			Labels:    map[string]string{},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				// TODO: Handle more than or less than 1 Container
				Name:         referencePod.Spec.Containers[0].Name,
				Image:        "k3d-rancher-logging-explorer-registry:5000/rancher-logging-explorer/pod-simulator:latest",
				VolumeMounts: []v1.VolumeMount{{Name: "config-volume", MountPath: "/var/logs"}},
			}},
			Volumes: []v1.Volume{
				{
					Name: "config-volume",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{Name: fmt.Sprintf("%s-configmap", flowTest.Spec.ReferencePod.Name)},
						},
					},
				},
			},
		},
	}

	simulationPod.ObjectMeta.Labels = referencePod.ObjectMeta.Labels
	simulationPod.ObjectMeta.Labels["app.kubernetes.io/name"] = "pod-simulation"
	simulationPod.ObjectMeta.Labels["app.kubernetes.io/managed-by"] = "rancher-logging-explorer"
	simulationPod.ObjectMeta.Labels["app.kubernetes.io/created-by"] = "logging-plumber"
	simulationPod.ObjectMeta.Labels["loggingplumber.isala.me/flowtest-uuid"] = string(flowTest.ObjectMeta.UID)
	simulationPod.ObjectMeta.Labels["loggingplumber.isala.me/flowtest"] = flowTest.ObjectMeta.Name

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
					Labels: map[string]string{
						"app.kubernetes.io/name":       "pod-simulation",
						"app.kubernetes.io/managed-by": "rancher-logging-explorer",
						"app.kubernetes.io/created-by": "logging-plumber",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Name:  "log-output",
						Image: "paynejacob/log-output:latest",
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

	flowTest.Status.Status = loggingplumberv1alpha1.Running

	if err := r.Status().Update(ctx, &flowTest); err != nil {
		logger.Error(err, "failed to update flowtest status")
		return err
	}

	return nil
}

