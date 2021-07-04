package controllers

import (
	"context"
	"fmt"
	flowv1beta1 "github.com/banzaicloud/logging-operator/pkg/sdk/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *FlowTestReconciler) cleanUpResources(ctx context.Context, flowTestName string) error {
	logger := log.FromContext(ctx)

	matchingLabels := &client.MatchingLabels{"loggingplumber.isala.me/flowtest": flowTestName}

	var podList v1.PodList
	if err := r.List(ctx, &podList, matchingLabels); client.IgnoreNotFound(err) != nil {
		logger.Error(err, fmt.Sprintf("failed to get provisioned %s", podList.Kind))
		return err
	}

	for _, resource := range podList.Items {
		if err := r.Delete(ctx, &resource); client.IgnoreNotFound(err) != nil {
			logger.Error(err, fmt.Sprintf("failed to delete a provisioned %s", resource.Kind), "uuid", resource.GetUID(), "name", resource.GetName())
			return err
		}
		logger.V(1).Info(fmt.Sprintf("%s deleted", resource.Kind), "uuid", resource.GetUID(), "name", resource.GetName())
	}

	var configMapList v1.ConfigMapList
	if err := r.List(ctx, &configMapList, matchingLabels); client.IgnoreNotFound(err) != nil {
		logger.Error(err, fmt.Sprintf("failed to get provisioned %s", configMapList.Kind))
		return err
	}

	for _, resource := range configMapList.Items {
		if err := r.Delete(ctx, &resource); client.IgnoreNotFound(err) != nil {
			logger.Error(err, fmt.Sprintf("failed to delete a provisioned %s", resource.Kind), "uuid", resource.GetUID(), "name", resource.GetName())
			return err
		}
		logger.V(1).Info(fmt.Sprintf("%s deleted", resource.Kind), "uuid", resource.GetUID(), "name", resource.GetName())
	}

	var flows flowv1beta1.FlowList
	if err := r.List(ctx, &flows, &client.MatchingLabels{"loggingplumber.isala.me/flowtest": flowTestName}); client.IgnoreNotFound(err) != nil {
		logger.Error(err, fmt.Sprintf("failed to get provisioned %s", flows.Kind))
		//return err
	}

	for _, resource := range flows.Items {
		if err := r.Delete(ctx, &resource); client.IgnoreNotFound(err) != nil {
			logger.Error(err, fmt.Sprintf("failed to delete a provisioned %s", resource.Kind), "uuid", resource.GetUID(), "name", resource.GetName())
			return err
		}
		logger.V(1).Info(fmt.Sprintf("%s deleted", resource.Kind), "uuid", resource.GetUID(), "name", resource.GetName())
	}

	var outputs flowv1beta1.OutputList
	if err := r.List(ctx, &outputs, &client.MatchingLabels{"loggingplumber.isala.me/flowtest": flowTestName}); client.IgnoreNotFound(err) != nil {
		logger.Error(err, fmt.Sprintf("failed to get provisioned %s", outputs.Kind))
		//return err
	}

	for _, resource := range outputs.Items {
		if err := r.Delete(ctx, &resource); client.IgnoreNotFound(err) != nil {
			logger.Error(err, fmt.Sprintf("failed to delete a provisioned %s", resource.Kind), "uuid", resource.GetUID(), "name", resource.GetName())
			return err
		}
		logger.V(1).Info(fmt.Sprintf("%s deleted", resource.Kind), "uuid", resource.GetUID(), "name", resource.GetName())
	}

	return nil
}
