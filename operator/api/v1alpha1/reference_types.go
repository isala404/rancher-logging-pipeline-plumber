package v1alpha1

// +kubebuilder:validation:Required
type ReferenceObject struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type FlowStatus string

const (
	Created FlowStatus = "Created"
	Running FlowStatus = "Running"
	Skipped FlowStatus = "Skipped"
	Failed  FlowStatus = "Failed"
	Passed  FlowStatus = "Passed"
)
