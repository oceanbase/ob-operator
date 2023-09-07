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
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OBTenantRestoreSpec defines the desired state of OBTenantRestore
type OBTenantRestoreSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ObClusterName     string `json:"obClusterName"`
	RestoreTenantName string `json:"restoreTenantName"`
	Type              string `json:"type"`
	SourceUri         string `json:"sourceUri"`
	Until             string `json:"until,omitempty"`
}

// OBTenantRestoreStatus defines the observed state of OBTenantRestore
type OBTenantRestoreStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status           RestoreJobStatus  `json:"status"`
	JobStatus        batchv1.JobStatus `json:"jobStatus"`
	Progress         string            `json:"progress"`
	OperationContext *OperationContext `json:"operationContext,omitempty"`
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

type RestoreJobType string

const (
	RestoreJobRestore  RestoreJobType = "RESTORE"
	RestoreJobActivate RestoreJobType = "ACTIVATE"
)

type RestoreJobStatus string

const (
	RestoreJobRunning    RestoreJobStatus = "RUNNING"
	RestoreJobFailed     RestoreJobStatus = "FAILED"
	RestoreJobSuccessful RestoreJobStatus = "SUCCESSFUL"
	RestoreJobCanceled   RestoreJobStatus = "CANCELED"
)
