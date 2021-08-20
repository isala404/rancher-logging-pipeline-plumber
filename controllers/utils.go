package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	flowv1beta1 "github.com/banzaicloud/logging-operator/pkg/sdk/api/v1beta1"
	filters "github.com/banzaicloud/logging-operator/pkg/sdk/model/filter"
	"github.com/banzaicloud/logging-operator/pkg/sdk/model/output"
	loggingplumberv1alpha1 "github.com/mrsupiri/logging-pipeline-plumber/pkg/sdk/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	EventReasonProvision string = "Provision"
	EventReasonCleanup          = "Cleanup"
	EventReasonReconcile        = "Reconcile"
)

func (r *FlowTestReconciler) flowTemplates(flow flowv1beta1.Flow, flowTest loggingplumberv1alpha1.FlowTest) (flowv1beta1.Flow, flowv1beta1.Output) {
	flowTemplate := flowv1beta1.Flow{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "logging.banzaicloud.io/v1beta1",
			Kind:       "Flow",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-test", flow.ObjectMeta.Name),
			Namespace: flow.Namespace,
			Labels:    GetLabels(flow.ObjectMeta.Name, &flowTest),
		},
		Spec: flowv1beta1.FlowSpec{
			LocalOutputRefs: nil,
			Match:           nil,
			Filters: []flowv1beta1.Filter{{
				Grep: &filters.GrepConfig{
					Regexp: []filters.RegexpSection{{
						// Make sure flowtest-uuid is present in logs
						// So logs doesn't get leaked from other pods
						Key:     "kubernetes",
						Pattern: fmt.Sprintf(".*%s.*", flowTest.ObjectMeta.UID),
					}},
				},
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
			Labels:    GetLabels(flow.ObjectMeta.Name, &flowTest),
		},
		Spec: flowv1beta1.OutputSpec{
			HTTPOutput: &output.HTTPOutputConfig{
				Endpoint: fmt.Sprintf("http://logging-plumber-log-aggregator.%s.svc", r.AggregatorNamespace),
				Buffer: &output.Buffer{
					FlushMode:     "interval",
					FlushInterval: "1s",
				},
			},
		},
	}

	return flowTemplate, outTemplate
}

func (r *FlowTestReconciler) clusterFlowTemplates(flow flowv1beta1.ClusterFlow, flowTest loggingplumberv1alpha1.FlowTest) (flowv1beta1.ClusterFlow, flowv1beta1.ClusterOutput) {
	flowTemplate := flowv1beta1.ClusterFlow{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "logging.banzaicloud.io/v1beta1",
			Kind:       "ClusterFlow",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-test", flow.ObjectMeta.Name),
			Namespace: flowTest.Spec.ReferenceFlow.Namespace,
			Labels:    GetLabels(flow.ObjectMeta.Name, &flowTest),
		},
		Spec: flowv1beta1.ClusterFlowSpec{
			GlobalOutputRefs: nil,
			Match:            nil,
			Filters: []flowv1beta1.Filter{{
				Grep: &filters.GrepConfig{
					Regexp: []filters.RegexpSection{{
						// Make sure flowtest-uuid is present in logs
						// So logs doesn't get leaked from other pods
						Key:     "kubernetes",
						Pattern: fmt.Sprintf(".*%s.*", flowTest.ObjectMeta.UID),
					}},
				},
			}},
		},
	}

	outTemplate := flowv1beta1.ClusterOutput{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "logging.banzaicloud.io/v1beta1",
			Kind:       "ClusterOutput",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-test", flow.ObjectMeta.Name),
			Namespace: flowTest.Spec.ReferenceFlow.Namespace,
			Labels:    GetLabels(flow.ObjectMeta.Name, &flowTest),
		},
		Spec: flowv1beta1.ClusterOutputSpec{
			OutputSpec: flowv1beta1.OutputSpec{
				HTTPOutput: &output.HTTPOutputConfig{
					Endpoint: fmt.Sprintf("http://logging-plumber-log-aggregator.%s.svc", r.AggregatorNamespace),
					Buffer: &output.Buffer{
						FlushMode:     "interval",
						FlushInterval: "1s",
					},
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

type Image struct {
	Repository string
	Tag        string
	PullPolicy string
}

// https://stackoverflow.com/a/40326580
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func (r *FlowTestReconciler) checkIndex(ctx context.Context, indexName string) (bool, error) {
	logger := log.FromContext(ctx)

	client := &http.Client{}

	// NOTE: When developing this requires port-forward because controller is running locally
	req, err := http.NewRequest("GET", getEnv("LOG_OUTPUT_ENDPOINT", fmt.Sprintf("http://logging-plumber-log-aggregator.%s.svc/", r.AggregatorNamespace)), nil)
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

func GetLabels(name string, flowTest *loggingplumberv1alpha1.FlowTest, labelsMaps ...map[string]string) map[string]string {
	labels := map[string]string{}

	for _, labelsMap := range labelsMaps {
		for k, v := range labelsMap {
			labels[k] = v
		}
	}

	labels["app.kubernetes.io/name"] = name
	if flowTest != nil {
		labels["loggingplumber.isala.me/flowtest-uuid"] = string(flowTest.ObjectMeta.UID)
		labels["loggingplumber.isala.me/flowtest"] = flowTest.ObjectMeta.Name
	}
	labels["app.kubernetes.io/created-by"] = "logging-plumber"
	labels["app.kubernetes.io/managed-by"] = "logging-pipeline-plumber"
	return labels
}
