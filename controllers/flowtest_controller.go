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
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	flowv1beta1 "github.com/banzaicloud/logging-operator/pkg/sdk/api/v1beta1"
	loggingpipelineplumberv1alpha1 "github.com/mrsupiri/logging-pipeline-plumber/pkg/sdk/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// FlowTestReconciler reconciles a FlowTest object
type FlowTestReconciler struct {
	AggregatorNamespace string
	PodSimulatorImage   Image
	LogOutputImage      Image
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=logging.banzaicloud.io,resources=flows;clusterflows;outputs;clusteroutputs,verbs=get;watch;list;create;delete
//+kubebuilder:rbac:groups=loggingpipelineplumber.isala.me,resources=flowtests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=loggingpipelineplumber.isala.me,resources=flowtests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=loggingpipelineplumber.isala.me,resources=flowtests/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods;services;configmaps;namespaces,verbs=get;watch;list;create;delete
//+kubebuilder:rbac:groups="",resources=pods/log,verbs=get;list
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

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

	var flowTest loggingpipelineplumberv1alpha1.FlowTest
	if err := r.Get(ctx, req.NamespacedName, &flowTest); err != nil {
		// all the resources are already deleted
		if apierrors.IsNotFound(err) {
			// Remove if log aggregator is still running
			if err := r.cleanUpOutputResources(ctx); client.IgnoreNotFound(err) != nil {
				return ctrl.Result{Requeue: true}, err
			}
			return ctrl.Result{Requeue: false}, nil
		}
		logger.Error(err, "failed to get the flowtest")
		return ctrl.Result{Requeue: false}, err
	}

	ctx = context.WithValue(ctx, "flowTest", flowTest)

	// name of our custom finalizer
	finalizerName := "flowtests.loggingpipelineplumber.isala.me/finalizer"
	// examine DeletionTimestamp to determine if object is under deletion
	if !flowTest.ObjectMeta.DeletionTimestamp.IsZero() {
		if err := r.deleteResources(ctx, finalizerName); err != nil {
			// if fail to delete the external dependency here, return with error
			// so that it can be retried
			return ctrl.Result{Requeue: true}, err
		}
		// Stop reconciliation as the item is being deleted
		return ctrl.Result{Requeue: false}, nil
	}

	if flowTest.ObjectMeta.Name == "" {
		logger.V(-1).Info("flowtest without a name queued")
		return ctrl.Result{Requeue: false}, nil
	}

	// Reconcile depending on status
	switch flowTest.Status.Status {

	// This will run only at first iteration
	case "":
		// Set the finalizer
		controllerutil.AddFinalizer(&flowTest, finalizerName)
		if err := r.Update(ctx, &flowTest); err != nil {
			logger.Error(err, "failed to add finalizer")
			return ctrl.Result{Requeue: true}, err
		}
		// Set the status
		flowTest.Status.Status = loggingpipelineplumberv1alpha1.Created
		if err := r.Status().Update(ctx, &flowTest); err != nil {
			logger.Error(err, "failed to set status as created")
			return ctrl.Result{Requeue: true}, err
		}
		r.Recorder.Event(&flowTest, v1.EventTypeNormal, EventReasonProvision, "moved to created state")
		return ctrl.Result{Requeue: true}, nil

	case loggingpipelineplumberv1alpha1.Created:
		if err := r.provisionResource(ctx); err != nil {
			r.Recorder.Event(&flowTest, v1.EventTypeWarning, EventReasonProvision, fmt.Sprintf("error while provision flow resources: %s", err.Error()))
			return ctrl.Result{Requeue: true}, r.setErrorStatus(ctx, err)
		}
		r.Recorder.Event(&flowTest, v1.EventTypeNormal, EventReasonProvision, "all the need resources were scheduled")
		// Give 1 minute to resource to provisioned
		return ctrl.Result{RequeueAfter: time.Minute}, nil

	case loggingpipelineplumberv1alpha1.Running:
		fiveMinuteAfterCreation := flowTest.CreationTimestamp.Add(5 * time.Minute)
		//        Timeout                            or    all test are passing
		if time.Now().After(fiveMinuteAfterCreation) || allTestPassing(flowTest.Status) {
			flowTest.Status.Status = loggingpipelineplumberv1alpha1.Completed
			if err := r.Status().Update(ctx, &flowTest); err != nil {
				logger.Error(err, "failed to set status as completed")
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
		}

		logger.V(1).Info("checking log indexes")
		err := r.checkForPassingFlowTest(ctx)
		if err != nil {
			r.Recorder.Event(&flowTest, v1.EventTypeWarning, EventReasonReconcile, fmt.Sprintf("error while checking log indexes: %s", err.Error()))
		}
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err

	case loggingpipelineplumberv1alpha1.Completed:
		if err := r.deleteResources(ctx, finalizerName); err != nil {
			return ctrl.Result{Requeue: true}, err
		}
		r.Recorder.Event(&flowTest, v1.EventTypeNormal, EventReasonCleanup, "all the provisioned resources were scheduled to be deleted")
		return ctrl.Result{}, nil

	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FlowTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&loggingpipelineplumberv1alpha1.FlowTest{}).
		WithEventFilter(eventFilter()).
		Complete(r)
}

func eventFilter() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CDR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
	}
}

func (r *FlowTestReconciler) checkForPassingFlowTest(ctx context.Context) error {
	logger := log.FromContext(ctx)
	flowTest := ctx.Value("flowTest").(loggingpipelineplumberv1alpha1.FlowTest)

	if flowTest.Spec.ReferenceFlow.Kind == "ClusterFlow" {
		var flows flowv1beta1.ClusterFlowList

		var referenceFlow flowv1beta1.ClusterFlow
		if err := r.Get(ctx, types.NamespacedName{
			Namespace: flowTest.Spec.ReferenceFlow.Namespace,
			Name:      flowTest.Spec.ReferenceFlow.Name,
		}, &referenceFlow); err != nil {
			return err
		}

		if err := r.List(ctx, &flows, &client.MatchingLabels{"loggingpipelineplumber.isala.me/flowtest": flowTest.ObjectMeta.Name}); client.IgnoreNotFound(err) != nil {
			logger.Error(err, fmt.Sprintf("failed to get provisioned %s", flows.Kind))
			return err
		}

		for _, flow := range flows.Items {
			passing, err := r.checkIndex(ctx, flow.ObjectMeta.Name)
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
		if err := r.List(ctx, &flows, &client.MatchingLabels{"loggingpipelineplumber.isala.me/flowtest": flowTest.ObjectMeta.Name}); client.IgnoreNotFound(err) != nil {
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

			passing, err := r.checkIndex(ctx, flow.ObjectMeta.Name)
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

func setPassingFilter(passingFilters []flowv1beta1.Filter, filters []flowv1beta1.Filter, flowTest *loggingpipelineplumberv1alpha1.FlowTest) {
	for _, passingFilter := range passingFilters {
		for i, filter := range filters {
			if reflect.DeepEqual(passingFilter, filter) {
				flowTest.Status.FilterStatus[i] = true
			}
		}
	}
}

func setPassingMatches(passingMatches []flowv1beta1.Match, matches []flowv1beta1.Match, flowTest *loggingpipelineplumberv1alpha1.FlowTest) {
	for _, passingFilter := range passingMatches {
		for i, filter := range matches {
			if reflect.DeepEqual(passingFilter, filter) {
				flowTest.Status.MatchStatus[i] = true
			}
		}
	}
}

func setPassingClusterMatches(passingMatches []flowv1beta1.ClusterMatch, matches []flowv1beta1.ClusterMatch, flowTest *loggingpipelineplumberv1alpha1.FlowTest) {
	for _, passingFilter := range passingMatches {
		for i, filter := range matches {
			if reflect.DeepEqual(passingFilter, filter) {
				flowTest.Status.MatchStatus[i] = true
			}
		}
	}
}
