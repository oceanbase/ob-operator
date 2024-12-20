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

package response

// @Description Brief information about OBTenant
type OBTenantOverview struct {
	UID         string            `json:"uid" binding:"required"`                 // Unique identifier of the resource
	Name        string            `json:"name" binding:"required"`                // Name of the resource
	Namespace   string            `json:"namespace" binding:"required"`           // Namespace of the resource
	TenantName  string            `json:"tenantName" binding:"required"`          // Name of the tenant in the database
	ClusterName string            `json:"clusterResourceName" binding:"required"` // Name of the cluster belonging to
	TenantRole  string            `json:"tenantRole" binding:"required"`          // Enum: Primary, Standby
	UnitNumber  int               `json:"unitNumber" binding:"required"`          // Number of units in every zone
	Topology    []OBTenantReplica `json:"topology"`                               // Topology of the tenant
	Status      string            `json:"status" binding:"required"`              // Status of the tenant
	CreateTime  string            `json:"createTime" binding:"required"`          // Creation time of the tenant
	Locality    string            `json:"locality" binding:"required"`            // Locality of the tenant units
	Charset     string            `json:"charset" binding:"required"`             // Charset of the tenant
	PrimaryZone string            `json:"primaryZone" binding:"required"`         // Primary zone of the tenant

	DeletionProtection bool `json:"deletionProtection" binding:"required"` // Whether the tenant is protected from deletion
}

type OBTenantDetail struct {
	OBTenantOverview    `json:",inline"`
	RootCredential      string `json:"rootCredential"`
	StandbyROCredential string `json:"standbyROCredential"`
	Version             string `json:"version"`

	PrimaryTenant string         `json:"primaryTenant"`
	RestoreSource *RestoreSource `json:"restoreSource,omitempty"`
}

type OBTenantReplica struct {
	Zone     string `json:"zone" binding:"required"`
	Priority int    `json:"priority" binding:"required"`
	// Enum: Readonly, Full
	Type        string `json:"type" binding:"required"`
	MaxCPU      int64  `json:"maxCPU" binding:"required"`
	MemorySize  int64  `json:"memorySize" binding:"required"`
	MinCPU      int64  `json:"minCPU,omitempty" binding:"required"`
	MaxIops     int64  `json:"maxIops,omitempty" binding:"required"`
	MinIops     int64  `json:"minIops,omitempty" binding:"required"`
	IopsWeight  int    `json:"iopsWeight,omitempty" binding:"required"`
	LogDiskSize int64  `json:"logDiskSize,omitempty" binding:"required"`
}

type RestoreSource struct {
	// Enum: OSS, NFS
	Type                string `json:"type" binding:"required"`
	ArchiveSource       string `json:"archiveSource" binding:"required"`
	BakDataSource       string `json:"bakDataSource" binding:"required"`
	OssAccessSecret     string `json:"ossAccessSecret,omitempty"`
	BakEncryptionSecret string `json:"bakEncryptionSecret,omitempty"`
	Until               string `json:"until,omitempty"`
}

type OBTenantStatistic OBClusterStatistic
