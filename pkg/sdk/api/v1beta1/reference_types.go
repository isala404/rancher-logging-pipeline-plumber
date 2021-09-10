package v1beta1

// +kubebuilder:validation:Required
type ReferenceObject struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type FlowStatus string

const (
	Created   FlowStatus = "Created"
	Running   FlowStatus = "Running"
	Completed FlowStatus = "Completed"
	Error     FlowStatus = "Error"
)
