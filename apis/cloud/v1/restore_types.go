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
	DestTenant      string                `json:"destTenant"`
	SourceTenant    string                `json:"sourceTenant"`
	Timestamp       string                `json:"timestamp"`
	Path            string                `json:"sourcePath"`
	Locality        string                `json:"locality,omitempty"`
	SourceCluster   RestoreClusterSpec    `json:"source"`
	ResourceUnit    ResourceUnitSpec      `json:"resourceUnit"`
	ResourcePool    ResourcePoolSpec      `json:"resourcePool"`
	Volume          []VolumeSpec          `json:"vloume,omitempty"`
	RestorePassword []RestorePasswordSpec `json:"restorePassword,omitempty"`
	Parameters      []Parameter           `json:"parameters,omitempty"`
}

// RestoreCluster defines the restore cluster
type RestoreClusterSpec struct {
	ClusterID   int    `json:"clusterID"`
	ClusterName string `json:"clusterName"`
}

// ResourceUnitSpec defines the resource unit config
type ResourceUnitSpec struct {
	Name          string `json:"name,omitempty"`
	MaxCPU        int    `json:"maxCPU"`
	MaxMemory     string `json:"maxMemory"`
	MaxIops       int    `json:"maxIops"`
	MaxDiskSize   string `json:"maxDiskSize"`
	MaxSessionNum int    `json:"maxSessionNum"`
	MinCPU        int    `json:"minCPU"`
	MinMemory     string `json:"minMemory"`
	MinIops       int    `json:"minIops"`
}

// ResourcePoolSpec defines the resources pool config
type ResourcePoolSpec struct {
	Name     string   `json:"name,omitempty"`
	UnitNum  int      `json:"unitNum"`
	ZoneList []string `json:"zoneList"`
}

// ResourcePoolSpec defines the restore password config
type RestorePasswordSpec struct {
	DatabasePassword        string `json:"databasePassword"`
	DatabasePasswordInfo    string `json:"databasePasswordInfo"`
	IncrementalPassword     string `json:"incrementalPassword"`
	IncrementalPasswordInfo string `json:"incrementalPasswordInfo"`
}

// RestoreStatus defines the observed state of Restore
type RestoreStatus struct {
	RestoreSet []RestoreSetSpec `json:"restoreSet"`
}

type RestoreSetSpec struct {
	JodID            int    `json:"jobID"`
	ClusterID        int    `json:"clusterID"`
	ClusterName      string `json:"clusterName"`
	TenantName       string `json:"tenantName"`
	BackupTenantName string `json:"backupTenantName"`
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
