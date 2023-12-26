/*
Copyright 2023.

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
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OBResourceRescueSpec defines the desired state of OBResourceRescue
type OBResourceRescueSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	TargetKind    string `json:"targetKind"`
	TargetResName string `json:"targetResName"`
	Type          string `json:"type"`
	TargetGV      string `json:"targetGV,omitempty"`
	Namespace     string `json:"namespace,omitempty"`
	TargetStatus  string `json:"targetStatus,omitempty"`
}

// OBResourceRescueStatus defines the observed state of OBResourceRescue
type OBResourceRescueStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	OperationContext *tasktypes.OperationContext `json:"operationContext,omitempty"`
	Status           string                      `json:"status"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// OBResourceRescue is the Schema for the obresourcerescues API
type OBResourceRescue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OBResourceRescueSpec   `json:"spec,omitempty"`
	Status OBResourceRescueStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OBResourceRescueList contains a list of OBResourceRescue
type OBResourceRescueList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OBResourceRescue `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OBResourceRescue{}, &OBResourceRescueList{})
}
