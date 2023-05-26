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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OBParameterSpec defines the desired state of OBParameter
type OBParameterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ClusterName string     `json:"clusterName"`
	ClusterId   int64      `json:"clusterId,omitempty"`
	Parameter   *Parameter `json:"parameter"`
}

// OBParameterStatus defines the observed state of OBParameter
type OBParameterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	OperationContext *OperationContext `json:"operationContext,omitempty"`
	Status           string            `json:"status"`
	Parameter        *[]ParameterValue `json:"parameter"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OBParameter is the Schema for the obparameters API
type OBParameter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OBParameterSpec   `json:"spec,omitempty"`
	Status OBParameterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OBParameterList contains a list of OBParameter
type OBParameterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OBParameter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OBParameter{}, &OBParameterList{})
}
