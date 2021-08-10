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
	"context"
	"fmt"
	"reflect"
	"time"

	flowv1beta1 "github.com/banzaicloud/logging-operator/pkg/sdk/api/v1beta1"
	loggingplumberv1alpha1 "github.com/mrsupiri/logging-pipeline-plumber/pkg/sdk/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// FlowTestReconciler reconciles a FlowTest object
type FlowTestReconciler struct {
	PodSimulatorImage Image
	LogOutputImage    Image
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=logging.banzaicloud.io,resources=flows;clusterflows;outputs;clusteroutputs,verbs=get;watch;list;create;delete
//+kubebuilder:rbac:groups=loggingplumber.isala.me,resources=flowtests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=loggingplumber.isala.me,resources=flowtests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=loggingplumber.isala.me,resources=flowtests/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods;services;configmaps;namespaces,verbs=get;watch;list;create;delete
//+kubebuilder:rbac:groups="",resources=pods/log,verbs=get;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
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
		if apierrors.IsNotFound(err) {
			if err := r.cleanUpResources(ctx, req.Name); client.IgnoreNotFound(err) != nil {
				return ctrl.Result{}, err
			}
		} else {
			logger.Error(err, "failed to get the flowtest")
		}
		return ctrl.Result{Requeue: false}, client.IgnoreNotFound(err)
	}

	ctx = context.WithValue(ctx, "flowTest", flowTest)

	if flowTest.Status.Status == "" {
		flowTest.Status.Status = loggingplumberv1alpha1.Created
		if err := r.Status().Update(ctx, &flowTest); err != nil {
			logger.Error(err, "failed to update flowtest status")
			return ctrl.Result{}, r.setErrorStatus(ctx, client.IgnoreNotFound(err))
		}
		return ctrl.Result{}, nil
	}

	if flowTest.Status.Status == loggingplumberv1alpha1.Created {
		if err := r.provisionResource(ctx); err != nil {
			return ctrl.Result{}, r.setErrorStatus(ctx, err)
		}
	}

	if flowTest.Status.Status == loggingplumberv1alpha1.Completed {
		if err := r.cleanUpResources(ctx, req.Name); client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, r.setErrorStatus(ctx, err)
		}
		if err := r.cleanUpOutputResources(ctx); client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, r.setErrorStatus(ctx, err)
		}
		return ctrl.Result{Requeue: false}, nil
	}

	if flowTest.Status.Status == loggingplumberv1alpha1.Running {
		oneMinuteAfterCreation := flowTest.CreationTimestamp.Add(1 * time.Minute)
		fiveMinuteAfterCreation := flowTest.CreationTimestamp.Add(5 * time.Minute)
		// Timeout
		if time.Now().After(fiveMinuteAfterCreation) {
			flowTest.Status.Status = loggingplumberv1alpha1.Completed
			if err := r.Status().Update(ctx, &flowTest); err != nil {
				logger.Error(err, "failed to update flowtest status")
				return ctrl.Result{}, r.setErrorStatus(ctx, client.IgnoreNotFound(err))
			}
			return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
		} else if time.Now().After(oneMinuteAfterCreation) { // Give 1 minute to resource to provisioned
			logger.V(1).Info("checking log indexes")
			if err := r.checkForPassingFlowTest(ctx); err != nil {
				return ctrl.Result{RequeueAfter: 30 * time.Second}, err
			}
		} else {
			return ctrl.Result{RequeueAfter: oneMinuteAfterCreation.Sub(time.Now())}, nil
		}
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FlowTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&loggingplumberv1alpha1.FlowTest{}).
		Complete(r)
}

func (r *FlowTestReconciler) checkForPassingFlowTest(ctx context.Context) error {
	logger := log.FromContext(ctx)
	flowTest := ctx.Value("flowTest").(loggingplumberv1alpha1.FlowTest)

	if flowTest.Spec.ReferenceFlow.Kind == "ClusterFlow" {
		var flows flowv1beta1.ClusterFlowList

		var referenceFlow flowv1beta1.ClusterFlow
		if err := r.Get(ctx, types.NamespacedName{
			Namespace: flowTest.Spec.ReferenceFlow.Namespace,
			Name:      flowTest.Spec.ReferenceFlow.Name,
		}, &referenceFlow); err != nil {
			return err
		}

		if err := r.List(ctx, &flows, &client.MatchingLabels{"loggingplumber.isala.me/flowtest": flowTest.ObjectMeta.Name}); client.IgnoreNotFound(err) != nil {
			logger.Error(err, fmt.Sprintf("failed to get provisioned %s", flows.Kind))
			return err
		}

		for _, flow := range flows.Items {
			passing, err := CheckIndex(ctx, flow.ObjectMeta.Name)
			if err != nil {
				return err
			}
			if passing {
				logger.V(1).Info(fmt.Sprintf("flow %s is passing", flow.ObjectMeta.Name))
				if err := r.Delete(ctx, &flow); err != nil {
					logger.Error(err, "failed to delete flow status")
					return err
				}
				setPassingFilter(flow.Spec.Filters, referenceFlow.Spec.Filters, &flowTest)
				setPassingClusterMatches(flow.Spec.Match, referenceFlow.Spec.Match, &flowTest)
			}
		}

	} else {
		var flows flowv1beta1.FlowList
		if err := r.List(ctx, &flows, &client.MatchingLabels{"loggingplumber.isala.me/flowtest": flowTest.ObjectMeta.Name}); client.IgnoreNotFound(err) != nil {
			logger.Error(err, fmt.Sprintf("failed to get provisioned %s", flows.Kind))
			return err
		}

		for _, flow := range flows.Items {
			var referenceFlow flowv1beta1.Flow
			if err := r.Get(ctx, types.NamespacedName{
				Namespace: flowTest.Spec.ReferenceFlow.Namespace,
				Name:      flowTest.Spec.ReferenceFlow.Name,
			}, &referenceFlow); err != nil {
				return err
			}

			passing, err := CheckIndex(ctx, flow.ObjectMeta.Name)
			if err != nil {
				return err
			}
			if passing {
				logger.V(1).Info(fmt.Sprintf("flow %s is passing", flow.ObjectMeta.Name))
				if err := r.Delete(ctx, &flow); err != nil {
					logger.Error(err, "failed to delete flow status")
					return err
				}

				setPassingFilter(flow.Spec.Filters, referenceFlow.Spec.Filters, &flowTest)
				setPassingMatches(flow.Spec.Match, referenceFlow.Spec.Match, &flowTest)
			}
		}
	}
	return r.Status().Update(ctx, &flowTest)
}

func setPassingFilter(passingFilters []flowv1beta1.Filter, filters []flowv1beta1.Filter, flowTest *loggingplumberv1alpha1.FlowTest) {
	for _, passingFilter := range passingFilters {
		for i, filter := range filters {
			if reflect.DeepEqual(passingFilter, filter) {
				flowTest.Status.FilterStatus[i] = true
			}
		}
	}
}

func setPassingMatches(passingMatches []flowv1beta1.Match, matches []flowv1beta1.Match, flowTest *loggingplumberv1alpha1.FlowTest) {
	for _, passingFilter := range passingMatches {
		for i, filter := range matches {
			if reflect.DeepEqual(passingFilter, filter) {
				flowTest.Status.MatchStatus[i] = true
			}
		}
	}
}

func setPassingClusterMatches(passingMatches []flowv1beta1.ClusterMatch, matches []flowv1beta1.ClusterMatch, flowTest *loggingplumberv1alpha1.FlowTest) {
	for _, passingFilter := range passingMatches {
		for i, filter := range matches {
			if reflect.DeepEqual(passingFilter, filter) {
				flowTest.Status.MatchStatus[i] = true
			}
		}
	}
}
