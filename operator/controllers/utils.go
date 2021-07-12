package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	flowv1beta1 "github.com/banzaicloud/logging-operator/pkg/sdk/api/v1beta1"
	"github.com/banzaicloud/logging-operator/pkg/sdk/model/output"
	loggingplumberv1alpha1 "github.com/mrsupiri/rancher-logging-explorer/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

func banzaiTemplates(flow flowv1beta1.Flow, flowTest loggingplumberv1alpha1.FlowTest, extraLabels map[string]string) (flowv1beta1.Flow, flowv1beta1.Output) {
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
			Match: []flowv1beta1.Match{{
				Select: &flowv1beta1.Select{Labels: extraLabels},
			}},
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
				Endpoint: "http://logging-plumber-log-aggregator.default.svc",
				Buffer: &output.Buffer{
					FlushMode:     "interval",
					FlushInterval: "10s",
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

type Index struct {
	Name     string    `json:"name"`
	FirstLog time.Time `json:"first_log"`
	LastLog  time.Time `json:"last_log"`
	LogCount int       `json:"log_count"`
}

func CheckIndex(ctx context.Context, indexName string) (bool, error) {
	logger := log.FromContext(ctx)

	client := &http.Client{}

	// TODO: UPDATE THIS SVC NAME
	// NOTE: When developing this requires port-forward because controller is running locally
	req, err := http.NewRequest("GET", "http://localhost:8111/", nil)
	if err != nil {
		logger.Error(err, "failed to create request for checking indexes")
		return false, err
	}
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err, "failed to fetch log indexes")
		return false, err
	}
	var indexes []Index
	if err := json.NewDecoder(resp.Body).Decode(&indexes); err != nil {
		logger.Error(err, "failed to fetch log indexes")
		return false, err
	}
	for _, index := range indexes {
		if index.Name == indexName {
			if index.LogCount > 0 {
				return true, nil
			} else {
				return false, nil
			}
		}
	}

	return false, nil
}
