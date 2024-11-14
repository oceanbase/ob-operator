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

	apitypes "github.com/oceanbase/ob-operator/api/types"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OBTenantVariableSpec defines the desired state of OBTenantVariable.
type OBTenantVariableSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of OBTenantVariable. Edit obtenantvariable_types.go to remove/update
	OBCluster string             `json:"obcluster"`
	OBTenant  string             `json:"obtenant"`
	Variable  *apitypes.Variable `json:"variable"`
}

// OBTenantVariableStatus defines the observed state of OBTenantVariable.
type OBTenantVariableStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	OperationContext *tasktypes.OperationContext `json:"operationContext,omitempty"`
	Status           string                      `json:"status"`
	Variable         apitypes.Variable           `json:"variable"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// OBTenantVariable is the Schema for the obtenantvariables API.
type OBTenantVariable struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OBTenantVariableSpec   `json:"spec,omitempty"`
	Status OBTenantVariableStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OBTenantVariableList contains a list of OBTenantVariable.
type OBTenantVariableList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OBTenantVariable `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OBTenantVariable{}, &OBTenantVariableList{})
}
