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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OBTenantSpec defines the desired state of OBTenant
type OBTenantSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ClusterName string `json:"clusterName"`
	TenantName  string `json:"tenantName"`
	UnitNumber int          `json:"unitNum"`

	//+kubebuilder:default=false
	ForceDelete bool `json:"forceDelete,omitempty"`
	Charset          string `json:"charset,omitempty"`
	Collate          string `json:"collate,omitempty"`
	Mode             string `json:"mode,omitempty"`
	ConnectWhiteList string `json:"connectWhiteList,omitempty"`

	Pools []ResourcePoolSpec `json:"pools"`
}

type UnitConfig struct {
	MaxCPU     resource.Quantity `json:"maxCPU"`
	MemorySize resource.Quantity `json:"memorySize"`
	MinCPU     resource.Quantity `json:"minCPU,omitempty"`
	MaxIops    int               `json:"maxIops,omitempty"`
	MinIops    int               `json:"minIops,omitempty"`
	IopsWeight  int               `json:"iopsWeight,omitempty"`
	LogDiskSize resource.Quantity `json:"logDiskSize,omitempty"`
}

// OBTenantStatus defines the observed state of OBTenant
type OBTenantStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status           string               `json:"status"`
	TenantID		 int 					`json:"tenantID"`
	Pools            []ResourcePoolStatus `json:"resourcePool"`
	ConnectWhiteList string               `json:"connectWhiteList,omitempty"`
	Charset          string               `json:"charset,omitempty"`
	UnitNumber 		 int          		`json:"unitNum"`
	OperationContext *OperationContext     `json:"operationContext,omitempty"`
}

type ResourcePoolStatus struct {
	ZoneList string       `json:"zoneList"`
	Units    []UnitStatus `json:"units"`
	Priority int          `json:"priority,omitempty"`
	Type       LocalityType `json:"type"`
	UnitConfig UnitConfig       `json:"unitConfig"`
	UnitNumber int          `json:"unitNum"`
}

type UnitStatus struct {
	UnitId     int                 `json:"unitId"`
	ServerIP   string              `json:"serverIP"`
	ServerPort int                 `json:"serverPort"`
	Status     string              `json:"status"`
	Migrate    MigrateServerStatus `json:"migrate"`
}

type MigrateServerStatus struct {
	ServerIP   string `json:"serverIP"`
	ServerPort int    `json:"serverPort"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OBTenant is the Schema for the obtenants API
type OBTenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OBTenantSpec   `json:"spec,omitempty"`
	Status OBTenantStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OBTenantList contains a list of OBTenant
type OBTenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OBTenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OBTenant{}, &OBTenantList{})
}
