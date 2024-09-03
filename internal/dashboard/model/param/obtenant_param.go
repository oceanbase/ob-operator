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

package param

type CreateOBTenantParam struct {
	Name             string `json:"name" binding:"required"`
	Namespace        string `json:"namespace" binding:"required"`
	ClusterName      string `json:"obcluster" binding:"required"`
	TenantName       string `json:"tenantName" binding:"required"`
	UnitNumber       int    `json:"unitNum" binding:"required"`
	RootPassword     string `json:"rootPassword" binding:"required"`
	ConnectWhiteList string `json:"connectWhiteList,omitempty"`
	Charset          string `json:"charset,omitempty"`

	UnitConfig *UnitConfig        `json:"unitConfig" binding:"required"`
	Pools      []ResourcePoolSpec `json:"pools" binding:"required"`

	// Enum: Primary, Standby
	TenantRole TenantRole        `json:"tenantRole,omitempty"`
	Source     *TenantSourceSpec `json:"source,omitempty"`
}

type UpdateOBTenantParam CreateOBTenantParam

type ResourcePoolSpec struct {
	Zone     string `json:"zone" binding:"required"`
	Priority int    `json:"priority,omitempty"`
	// Enum: Readonly, Full
	Type string `json:"type,omitempty"`
}

type TenantSourceSpec struct {
	Tenant  *string            `json:"tenant,omitempty"`
	Restore *RestoreSourceSpec `json:"restore,omitempty"`
}

type RestoreSourceSpec struct {
	// Enum: OSS, NFS
	Type          BackupDestType `json:"type" binding:"required"`
	ArchiveSource string         `json:"archiveSource" binding:"required"`
	BakDataSource string         `json:"bakDataSource" binding:"required"`
	OSSAccessID   string         `json:"ossAccessId,omitempty"`
	OSSAccessKey  string         `json:"ossAccessKey,omitempty"`

	BakEncryptionPassword string              `json:"bakEncryptionPassword,omitempty"`
	Until                 *RestoreUntilConfig `json:"until,omitempty"`
}

type UnitConfig struct {
	CPUCount    string `json:"cpuCount" binding:"required"`
	MemorySize  string `json:"memorySize" binding:"required"`
	MaxIops     int64  `json:"maxIops,omitempty"`
	MinIops     int64  `json:"minIops,omitempty"`
	IopsWeight  int    `json:"iopsWeight,omitempty"`
	LogDiskSize string `json:"logDiskSize,omitempty"`
}

type RestoreUntilConfig struct {
	Timestamp *string `json:"timestamp,omitempty" example:"2024-02-23 17:47:00"`
	Unlimited bool    `json:"unlimited,omitempty"`
}

type ChangeUserPassword struct {
	// Description: The user name of the database account, only root is supported now.
	User     string `json:"user" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ReplayStandbyLog RestoreUntilConfig

type ChangeTenantRole struct {
	Failover   bool `json:"failover,omitempty"`
	Switchover bool `json:"switchover,omitempty"`
}

type PatchUnitConfig struct {
	UnitConfig *UnitConfig        `json:"unitConfig" binding:"required"`
	Pools      []ResourcePoolSpec `json:"pools" binding:"required"`
}

type PatchTenant struct {
	UnitNumber *int `json:"unitNum,omitempty"`
	// Deprecated
	// Description: Deprecated, use PATCH /obtenants/:namespace/:name/pools/:zoneName instead
	UnitConfig *PatchUnitConfig `json:"unitConfig,omitempty"`
}

type TenantPoolSpec struct {
	Priority   int        `json:"priority"`
	UnitConfig UnitConfig `json:"unitConfig"`
}

type TenantPoolName struct {
	NamespacedName `json:",inline"`
	ZoneName       string `json:"zoneName" uri:"zoneName"`
}
