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
type OBTenantBrief struct {
	Name        string            `json:"name"`        // Name of the resource
	Namespace   string            `json:"namespace"`   // Namespace of the resource
	TenantName  string            `json:"tenantName"`  // Name of the tenant in the database
	ClusterName string            `json:"clusterName"` // Name of the cluster belonging to
	TenantRole  string            `json:"tenantRole"`  // Enum: Primary, Standby
	UnitNumber  int               `json:"unitNumber"`  // Number of units in every zone
	Topology    []OBTenantReplica `json:"topology"`    // Topology of the tenant
	Status      string            `json:"status"`      // Status of the tenant
	CreateTime  string            `json:"createTime"`  // Creation time of the tenant
	Locality    string            `json:"locality"`    // Locality of the tenant units
	Charset     string            `json:"charset"`     // Charset of the tenant
}

type OBTenantDetail struct {
	OBTenantBrief       `json:",inline"`
	RootCredential      string `json:"rootCredential"`
	StandbyROCredentail string `json:"standbyROCredentail"`
	Version             string `json:"version"`

	PrimaryTenant string         `json:"primaryTenant"`
	RestoreSource *RestoreSource `json:"restoreSource,omitempty"`
}

type OBTenantReplica struct {
	Zone     string `json:"zone"`
	Priority int    `json:"priority"`
	// Enum: Readonly, Full
	Type        string `json:"type"`
	MaxCPU      string `json:"maxCPU"`
	MemorySize  string `json:"memorySize"`
	MinCPU      string `json:"minCPU,omitempty"`
	MaxIops     int    `json:"maxIops,omitempty"`
	MinIops     int    `json:"minIops,omitempty"`
	IopsWeight  int    `json:"iopsWeight,omitempty"`
	LogDiskSize string `json:"logDiskSize,omitempty"`
}

type RestoreSource struct {
	// Enum: OSS, NFS
	Type                string `json:"type"`
	ArchiveSource       string `json:"archiveSource"`
	BakDataSource       string `json:"bakDataSource"`
	OssAccessSecret     string `json:"ossAccessSecret,omitempty"`
	BakEncryptionSecret string `json:"bakEncryptionSecret,omitempty"`
	Until               string `json:"until,omitempty"`
}
