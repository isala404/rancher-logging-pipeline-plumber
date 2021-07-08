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
	flowv1beta1 "github.com/banzaicloud/logging-operator/pkg/sdk/api/v1beta1"
	loggingplumberv1alpha1 "github.com/mrsupiri/rancher-logging-explorer/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// FlowTestReconciler reconciles a FlowTest object
type FlowTestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=logging.banzaicloud.io,resources=flow,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=logging.banzaicloud.io,resources=output,verbs=get;list;watch;create;update;patch;delete
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
		return ctrl.Result{Requeue: false}, nil
	}

	if flowTest.Status.Status == loggingplumberv1alpha1.Running {
		// Timeout
		twoMinuteAfterCreation := flowTest.CreationTimestamp.Add(2 * time.Minute)
		if time.Now().After(twoMinuteAfterCreation) {
			flowTest.Status.Status = loggingplumberv1alpha1.Completed
			if err := r.Status().Update(ctx, &flowTest); err != nil {
				logger.Error(err, "failed to update flowtest status")
				return ctrl.Result{}, r.setErrorStatus(ctx, client.IgnoreNotFound(err))
			}
			return ctrl.Result{}, nil
		} else {
			logger.V(1).Info("checking log indexes")
			if err := r.checkForPassingFlowTest(ctx); err != nil {
				return ctrl.Result{RequeueAfter: 30 * time.Second}, err
			}
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

	var flows flowv1beta1.FlowList
	if err := r.List(ctx, &flows, &client.MatchingLabels{"loggingplumber.isala.me/flowtest": flowTest.ObjectMeta.Name}); client.IgnoreNotFound(err) != nil {
		logger.Error(err, fmt.Sprintf("failed to get provisioned %s", flows.Kind))
		return err
	}

	for _, flow := range flows.Items {
		if flow.Status.Active != nil && *flow.Status.Active == false {
			continue
		}
		passing, err := CheckIndex(ctx, flow.ObjectMeta.Name)
		if err != nil {
			return err
		}
		if passing {
			active := false
			flow.Status.Active = &active
			logger.V(1).Info(fmt.Sprintf("flow %s is passing", flow.ObjectMeta.Name))
			if err := r.Status().Update(ctx, &flow); err != nil {
				logger.Error(err, "failed to update flow status")
				return err
			}
			for _, match := range flow.Spec.Match {
				flowTest.Status.PassedMatches = appendMatchIfMissing(flowTest.Status.PassedMatches, &match)
			}
			for _, filter := range flow.Spec.Filters {
				flowTest.Status.PassedFilters = appendFilterIfMissing(flowTest.Status.PassedFilters, &filter)
			}
			if err := r.Status().Update(ctx, &flowTest); err != nil {
				logger.Error(err, "failed to update flow status")
				return err
			}
		}
	}

	return nil
}

func appendMatchIfMissing(matches []*flowv1beta1.Match, match *flowv1beta1.Match) []*flowv1beta1.Match {
	for _, ele := range matches {
		if reflect.DeepEqual(&ele, &match) {
			return matches
		}
	}
	return append(matches, match)
}

func appendFilterIfMissing(filters []*flowv1beta1.Filter, filter *flowv1beta1.Filter) []*flowv1beta1.Filter {
	for _, ele := range filters {
		if reflect.DeepEqual(&ele, &filter) {
			return filters
		}
	}
	return append(filters, filter)
}
