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
	"github.com/oceanbase/ob-operator/api/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OBTenantOperationSpec defines the desired state of OBTenantOperation
type OBTenantOperationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Type       constants.TenantOperationType `json:"type"`
	Switchover *OBTenantOpSwitchoverSpec     `json:"switchover,omitempty"`
	Failover   *OBTenantOpFailoverSpec       `json:"failover,omitempty"`
	ChangePwd  *OBTenantOpChangePwdSpec      `json:"changePwd,omitempty"`
}

type OBTenantOpSwitchoverSpec struct {
	PrimaryTenant string `json:"primaryTenant"`
	StandbyTenant string `json:"standbyTenant"`
}

type OBTenantOpFailoverSpec struct {
	StandbyTenant string `json:"standbyTenant"`
}

type OBTenantOpChangePwdSpec struct {
	Tenant    string                 `json:"tenant"`
	SecretRef corev1.SecretReference `json:"secretRef"`
}

// OBTenantOperationStatus defines the observed state of OBTenantOperation
type OBTenantOperationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status           constants.TenantOperationStatus `json:"status"`
	OperationContext *OperationContext               `json:"operationContext,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OBTenantOperation is the Schema for the obtenantoperations API
type OBTenantOperation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OBTenantOperationSpec   `json:"spec,omitempty"`
	Status OBTenantOperationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OBTenantOperationList contains a list of OBTenantOperation
type OBTenantOperationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OBTenantOperation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OBTenantOperation{}, &OBTenantOperationList{})
}
