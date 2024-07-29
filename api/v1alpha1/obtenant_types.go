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

	apitypes "github.com/oceanbase/ob-operator/api/types"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized

// OBTenantSpec defines the desired state of OBTenant
type OBTenantSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ClusterName string `json:"obcluster"`
	TenantName  string `json:"tenantName"`
	UnitNumber  int    `json:"unitNum"`

	//+kubebuilder:default=false
	ForceDelete bool `json:"forceDelete,omitempty"`
	//+kubebuilder:default=utf8mb4
	Charset string `json:"charset,omitempty"`
	Collate string `json:"collate,omitempty"` // DB fill collate automatically according to charset
	//+kubebuilder:default=%
	ConnectWhiteList string `json:"connectWhiteList,omitempty"`

	Pools []ResourcePoolSpec `json:"pools"`

	//+kubebuilder:default=PRIMARY
	TenantRole  apitypes.TenantRole `json:"tenantRole,omitempty"`
	Source      *TenantSourceSpec   `json:"source,omitempty"`
	Credentials TenantCredentials   `json:"credentials,omitempty"`
	Scenario    string              `json:"scenario,omitempty"`
}

type TenantCredentials struct {
	Root      string `json:"root,omitempty"`
	StandbyRO string `json:"standbyRo,omitempty"`
}

// Source for restoring or creating standby
type TenantSourceSpec struct {
	Tenant  *string            `json:"tenant,omitempty"`
	Restore *RestoreSourceSpec `json:"restore,omitempty"`
}

type ResourcePoolSpec struct {
	Zone string `json:"zone"`
	//+kubebuilder:default=1
	Priority   int           `json:"priority,omitempty"`
	Type       *LocalityType `json:"type,omitempty"`
	UnitConfig *UnitConfig   `json:"resource"`
}

// TODO Split LocalityType struct to SpecLocalityType and StatusLocalityType
type LocalityType struct {
	Name    string `json:"name"`
	Replica int    `json:"replica"`
	// TODO move isActive to ResourcePoolSpec And ResourcePoolStatus
	IsActive bool `json:"isActive"`
}

// TODO Split UnitConfig struct to SpecUnitConfig and StatusUnitConfig
type UnitConfig struct {
	MaxCPU      resource.Quantity `json:"maxCPU"`
	MemorySize  resource.Quantity `json:"memorySize"`
	MinCPU      resource.Quantity `json:"minCPU,omitempty"`
	MaxIops     int               `json:"maxIops,omitempty"`
	MinIops     int               `json:"minIops,omitempty"`
	IopsWeight  int               `json:"iopsWeight,omitempty"`
	LogDiskSize resource.Quantity `json:"logDiskSize,omitempty"`
}

// +kubebuilder:object:generate=false
// OBTenantStatus defines the observed state of OBTenant
type OBTenantStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status           string                      `json:"status"`
	Pools            []ResourcePoolStatus        `json:"resourcePool"`
	OperationContext *tasktypes.OperationContext `json:"operationContext,omitempty"`
	TenantRecordInfo TenantRecordInfo            `json:"tenantRecordInfo,omitempty"`

	TenantRole  apitypes.TenantRole `json:"tenantRole,omitempty"`
	Source      *TenantSourceStatus `json:"source,omitempty"`
	Credentials TenantCredentials   `json:"credentials,omitempty"`
}

type TenantSourceStatus struct {
	Tenant  *string                `json:"tenant,omitempty"`
	Restore *OBTenantRestoreStatus `json:"restore,omitempty"`
}

func (in *OBTenantStatus) DeepCopyInto(out *OBTenantStatus) {
	*out = *in
	if in.Pools != nil {
		in, out := &in.Pools, &out.Pools
		*out = make([]ResourcePoolStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.OperationContext != nil {
		in, out := &in.OperationContext, &out.OperationContext
		*out = new(tasktypes.OperationContext)
		(*in).DeepCopyInto(*out)
	}
	out.TenantRecordInfo = in.TenantRecordInfo
	if in.Source != nil {
		in, out := &in.Source, &out.Source
		*out = new(TenantSourceStatus)
		(*in).DeepCopyInto(*out)
	}
}

func (in *TenantSourceStatus) DeepCopyInto(out *TenantSourceStatus) {
	*out = *in
	if in.Tenant != nil {
		in, out := &in.Tenant, &out.Tenant
		*out = new(string)
		**out = **in
	}
	if in.Restore != nil {
		in, out := &in.Restore, &out.Restore
		*out = new(OBTenantRestoreStatus)
		**out = **in
	}
}

type ResourcePoolStatus struct {
	ZoneList   string        `json:"zoneList"`
	Units      []UnitStatus  `json:"units"`
	Priority   int           `json:"priority,omitempty"`
	Type       *LocalityType `json:"type"`
	UnitConfig *UnitConfig   `json:"unitConfig"`
	UnitNumber int           `json:"unitNum"`
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

type TenantRecordInfo struct {
	TenantID         int    `json:"tenantID"`
	PrimaryZone      string `json:"primaryZone"`
	Locality         string `json:"locality"`
	PoolList         string `json:"poolList"`
	ConnectWhiteList string `json:"connectWhiteList,omitempty"`
	Charset          string `json:"charset,omitempty"`
	Collate          string `json:"collate,omitempty"`
	UnitNumber       int    `json:"unitNum,omitempty"`
	ZoneList         string `json:"zoneList,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="status",type="string",JSONPath=".status.status"
//+kubebuilder:printcolumn:name="tenantName",type="string",JSONPath=".spec.tenantName"
//+kubebuilder:printcolumn:name="tenantRole",type="string",JSONPath=".status.tenantRole"
//+kubebuilder:printcolumn:name="clusterName",type="string",JSONPath=".spec.obcluster"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="locality",type="string",JSONPath=".status.tenantRecordInfo.locality",priority=1
//+kubebuilder:printcolumn:name="primaryZone",type="string",JSONPath=".status.tenantRecordInfo.primaryZone",priority=1
//+kubebuilder:printcolumn:name="poolList",type="string",JSONPath=".status.tenantRecordInfo.poolList",priority=1
//+kubebuilder:printcolumn:name="charset",type="string",JSONPath=".status.tenantRecordInfo.charset",priority=1

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
