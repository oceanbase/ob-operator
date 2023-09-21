/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OBTenantRestoreSpec defines the desired state of OBTenantRestore
type OBTenantRestoreSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	TargetTenant  string               `json:"targetTenant"`
	TargetCluster string               `json:"targetCluster"`
	RestoreRole   constants.TenantRole `json:"restoreRole"`
	Source        TenantSourceSpec     `json:"source"`
}

// +kubebuilder:object:generate=false
// OBTenantRestoreStatus defines the observed state of OBTenantRestore
type OBTenantRestoreStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status           constants.RestoreJobStatus `json:"status"`
	RestoreProgress  *model.RestoreHistory      `json:"restoreProgress,omitempty"`
	OperationContext *OperationContext          `json:"operationContext,omitempty"`
}

func (in *OBTenantRestoreStatus) DeepCopyInto(out *OBTenantRestoreStatus) {
	*out = *in
	if in.RestoreProgress != nil {
		in, out := &in.RestoreProgress, &out.RestoreProgress
		*out = new(model.RestoreHistory)
		**out = **in
	}
	if in.OperationContext != nil {
		in, out := &in.OperationContext, &out.OperationContext
		*out = new(OperationContext)
		(*in).DeepCopyInto(*out)
	}
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OBTenantRestore is the Schema for the obtenantrestores API
// An instance of OBTenantRestore stands for a tenant restore job
type OBTenantRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OBTenantRestoreSpec   `json:"spec,omitempty"`
	Status OBTenantRestoreStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OBTenantRestoreList contains a list of OBTenantRestore
type OBTenantRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OBTenantRestore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OBTenantRestore{}, &OBTenantRestoreList{})
}
