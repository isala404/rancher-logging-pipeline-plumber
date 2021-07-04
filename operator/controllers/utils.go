package controllers

import (
	"context"
	"fmt"
	flowv1beta1 "github.com/banzaicloud/logging-operator/pkg/sdk/api/v1beta1"
	"github.com/banzaicloud/logging-operator/pkg/sdk/model/output"
	loggingplumberv1alpha1 "github.com/mrsupiri/rancher-logging-explorer/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func banzaiTemplates(flow flowv1beta1.Flow, flowTest loggingplumberv1alpha1.FlowTest) (flowv1beta1.Flow, flowv1beta1.Output) {
	flowTemplate := flowv1beta1.Flow{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "logging.banzaicloud.io/v1beta1",
			Kind:       "Flow",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-test", flow.ObjectMeta.Name),
			Namespace: flow.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":                flow.ObjectMeta.Name,
				"app.kubernetes.io/managed-by":          "rancher-logging-explorer",
				"app.kubernetes.io/created-by":          "logging-plumber",
				"loggingplumber.isala.me/flowtest-uuid": string(flowTest.ObjectMeta.UID),
				"loggingplumber.isala.me/flowtest":      flowTest.ObjectMeta.Name,
			},
		},
		Spec: flowv1beta1.FlowSpec{
			LocalOutputRefs: nil,
		},
	}

	outTemplate := flowv1beta1.Output{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "logging.banzaicloud.io/v1beta1",
			Kind:       "Output",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-test", flow.ObjectMeta.Name),
			Namespace: flow.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":                flow.ObjectMeta.Name,
				"app.kubernetes.io/managed-by":          "rancher-logging-explorer",
				"app.kubernetes.io/created-by":          "logging-plumber",
				"loggingplumber.isala.me/flowtest-uuid": string(flowTest.ObjectMeta.UID),
				"loggingplumber.isala.me/flowtest":      flowTest.ObjectMeta.Name,
			},
		},
		Spec: flowv1beta1.OutputSpec{
			HTTPOutput: &output.HTTPOutputConfig{
				Endpoint: fmt.Sprintf("http://logging-plumber-log-aggregator/%s", fmt.Sprintf("%s-slice", flow.ObjectMeta.Name)),
				Buffer: &output.Buffer{
					FlushMode:     "10s",
					FlushInterval: "interval",
				},
			},
		},
	}

	return flowTemplate, outTemplate
}

func (r *FlowTestReconciler) setErrorStatus(ctx context.Context, err error) error {
	logger := log.FromContext(ctx)
	flowTest := ctx.Value("flowTest").(loggingplumberv1alpha1.FlowTest)
	if err != nil {
		flowTest.Status.Status = loggingplumberv1alpha1.Error

		if err := r.Status().Update(ctx, &flowTest); err != nil {
			logger.Error(err, "failed to update flowtest status")
			return err
		}
		return err
	}
	return nil
}
