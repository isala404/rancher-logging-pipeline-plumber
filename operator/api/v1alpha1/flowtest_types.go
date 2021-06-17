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

package v1alpha1

import (
	flowv1beta1 "github.com/banzaicloud/logging-operator/pkg/sdk/api/v1beta1/"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FlowTestSpec defines the desired state of FlowTest
type FlowTestSpec struct {
	// +kubebuilder:validation:Required
	ReferencePod ReferenceObject `json:"referencePod"`
	// +kubebuilder:validation:Required
	ReferenceFlow ReferenceObject `json:"referenceFlow"`
	// +kubebuilder:validation:Required
	Messages []string `json:"messages"`
	// +kubebuilder:validation:Required
	Match []flowv1beta1.Match `json:"match,omitempty"`
	// +kubebuilder:validation:Required
	Filters []flowv1beta1.Filter `json:"filters,omitempty"`
}

// FlowTestStatus defines the observed state of FlowTest
type FlowTestStatus struct {
	SimulationPod  ReferenceObject `json:"simulationPod"`
	SimulationFlow ReferenceObject `json:"simulationFlow"`
	OutputIndex    string          `json:"outputIndex"`
	LogView        string          `json:"logView"`
	// +kubebuilder:validation:Enum=Created,Running,Failed,Skipped,Passed
	Status FlowStatus `json:"status"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.simulationPod.name",name="SimulationPod",type="string"
// +kubebuilder:printcolumn:JSONPath=".status.simulationFlow.name",name="SimulationFlow",type="string"
// +kubebuilder:printcolumn:JSONPath=".status.logView",name="LogView",type="string"
// +kubebuilder:printcolumn:JSONPath=".status.logView",name="LogView",type="string"

// FlowTest is the Schema for the flowtests API
type FlowTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlowTestSpec   `json:"spec,omitempty"`
	Status FlowTestStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FlowTestList contains a list of FlowTest
type FlowTestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FlowTest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FlowTest{}, &FlowTestList{})
}
