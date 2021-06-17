package v1alpha1

type ReferenceObject struct {
	// +kubebuilder:validation:Required
	Kind string `json:"kind"`
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// +kubebuilder:validation:Required
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
