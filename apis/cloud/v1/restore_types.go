/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RestoreSpec defines the desired state of Restore
type RestoreSpec struct {
    Source SourceSpec `json:"source"`
    Dest   DestSpec   `json:"dest"`
    SavePoint SavePointSpec `json:"savePoint"`
	Secret        string            `json:"secret,omitempty"`
	Parameters    []Parameter       `json:"parameters,omitempty"`
}

// SourceSpec defines the source of restore
type SourceSpec struct {
	ClusterID   int    `json:"clusterID"`
	ClusterName string `json:"clusterName"`
    Tennat string `json: "tenant"`
    Path PathSpec `json:"path"`
}

// PathSpec defines the data path, for oceanbase 3.x, use root, for oceanbase 4.x, use data and log
type PathSpec struct {
    Root string `json:"root"`
    Data string `json:"data"`
    Log  string `json:"log"`
}

// SavePointSpec defines the savepoint to restore to
type SavePointSpec struct {
    Type string `json:"type"`
    Value string `json:"value"`
}

// DestSpec defines the dest of restore
type DestSpec struct {
	ClusterID   int    `json:"clusterID"`
	ClusterName string `json:"clusterName"`
    Tennat string `json: "tenant"`
    KmsEncryptInfo stirng `json: "kmsEncryptInfo"`
	Topology    []TenantReplica `json:"topology"`
}

// RestoreStatus defines the observed state of Restore
type RestoreStatus struct {
	RestoreSet []RestoreSetSpec `json:"restoreSet"`
}

type RestoreSetSpec struct {
	JodID            int    `json:"jobID"`
	RestoreClusterID        int    `json:"clusterID"`
	RestoreClusterName      string `json:"clusterName"`
	RestoreTenant       string `json:"tenantName"`
	SourceClusterID        int    `json:"clusterID"`
	SourceClusterName      string `json:"clusterName"`
	SourceTenantName string `json:"backupTenantName"`
	Status           string `json:"status"`
	Timestamp        string `json:"restoreTimestamp"`
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Restore is the Schema for the restores API
type Restore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RestoreSpec   `json:"spec,omitempty"`
	Status RestoreStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RestoreList contains a list of Restore
type RestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Restore `json:"items"`
}
