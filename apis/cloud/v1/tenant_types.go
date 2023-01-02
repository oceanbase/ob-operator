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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TenantSpec defines the desired state of Tenant
type TenantSpec struct {
	ClusterID   int              `json:"clusterID"`
	ClusterName string           `json:"clusterName"`
	Topology    []TenantTopology `json:"topology"`

	Charset           string      `json:"charset,omitempty"`
	Collate           string      `json:"collate,omitempty"`
	Mode              string      `json:"mode,omitempty"`
	DefaultTablegroup string      `json:"tablegroup,omitempty"`
	Comment           string      `json:"comment,omitempty"`
	LogonlyReplicaNum int         `json:"logonlyReplicaNum,omitempty"`
	Variables         []Parameter `json:"variables,omitempty"`
}

type TenantTopology struct {
	ZoneName      string       `json:"zone"`
	UnitNumber    int          `json:"unitNum"`
	Priority      int          `json:"priority,omitempty"`
	Type          string       `json:"type"`
	ResourceUnits ResourceUnit `json:"resource"`
}

type ResourceUnit struct {
	MaxCPU        resource.Quantity `json:"maxCPU"`
	MinCPU        resource.Quantity `json:"minCPU"`
	MaxMemory     resource.Quantity `json:"maxMemory"`
	MinMemory     resource.Quantity `json:"minMemory"`
	MaxDiskSize   resource.Quantity `json:"maxDiskSize,omitempty"`
	MaxIops       int               `json:"maxIops,omitempty"`
	MinIops       int               `json:"minIops,omitempty"`
	MaxSessionNum int               `json:"maxSessionNum,omitempty"`
}

// TenantStatus defines the observed state of Tenant
type TenantStatus struct {
	Status            string                 `json:"status"`
	Topology          []TenantTopologyStatus `json:"topology"`
	Charset           string                 `json:"charset,omitempty"`
	ReplicaNum        int                    `json:"replicaNum"`
	LogonlyReplicaNum int                    `json:"logonlyReplicaNum"`
}

type TenantTopologyStatus struct {
	ZoneName      string       `json:"zone"`
	UnitConfigs   []Unit       `json:"units"`
	Priority      int          `json:"priority,omitempty"`
	Type          string       `json:"type"`
	ResourceUnits ResourceUnit `json:"resource"`
	UnitNumber    int          `json:"unitNum"`
}

type Unit struct {
	UnitId     int           `json:"unitId"`
	ServerIP   string        `json:"serverIP"`
	ServerPort int           `json:"serverPort"`
	Status     string        `json:"status"`
	Migrate    MigrateServer `json:"migrate"`
}

type MigrateServer struct {
	ServerIP   string `json:"serverIP"`
	ServerPort int    `json:"serverPort"`
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Tenant is the Schema for the tenants API
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantSpec   `json:"spec,omitempty"`
	Status TenantStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TenantList contains a list of Tenant
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tenant `json:"items"`
}
