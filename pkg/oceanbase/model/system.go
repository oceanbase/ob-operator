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

package model

// OBTenant is the tenant model of OB system
type OBTenant struct {
	TenantId                 int64  `json:"tenant_id" db:"tenant_id"`
	TenantName               string `json:"tenant_name" db:"tenant_name"`
	TenantType               string `json:"tenant_type" db:"tenant_type"`
	CreateTime               string `json:"create_time" db:"create_time"`
	ModifyTime               string `json:"modify_time" db:"modify_time"`
	PrimaryZone              string `json:"primary_zone" db:"primary_zone"`
	Locality                 string `json:"locality" db:"locality"`
	PreviousLocality         string `json:"previous_locality" db:"previous_locality"`
	CompatibilityMode        string `json:"compatibility_mode" db:"compatibility_mode"`
	Status                   string `json:"status" db:"status"`
	InRecyclebin             string `json:"in_recyclebin" db:"in_recyclebin"`
	Locked                   string `json:"locked" db:"locked"`
	TenantRole               string `json:"tenant_role" db:"tenant_role"`
	SyncScn                  int64  `json:"sync_scn" db:"sync_scn"`
	ReplayableScn            int64  `json:"replayable_scn" db:"replayable_scn"`
	ReadableScn              int64  `json:"readable_scn" db:"readable_scn"`
	RecoveryUntilScn         int64  `json:"recovery_until_scn" db:"recovery_until_scn"`
	LogMode                  string `json:"log_mode" db:"log_mode"`
	ArbitrationServiceStatus string `json:"arbitration_service_status" db:"arbitration_service_status"`
}

// OBUnit is the unit model of OB system
type OBUnit struct {
	UnitId         int64  `json:"unit_id" db:"unit_id"`
	TenantId       int64  `json:"tenant_id" db:"tenant_id"`
	Status         string `json:"status" db:"status"`
	ResourcePoolId int64  `json:"resource_pool_id" db:"resource_pool_id"`
	UnitGroupId    int64  `json:"unit_group_id" db:"unit_group_id"`
	CreateTime     string `json:"create_time" db:"create_time"`
	ModifyTime     string `json:"modify_time" db:"modify_time"`
	Zone           string `json:"zone" db:"zone"`
	SvrIp          string `json:"svr_ip" db:"svr_ip"`
	SvrPort        int64  `json:"svr_port" db:"svr_port"`
	UnitConfigId   int64  `json:"unit_config_id" db:"unit_config_id"`
	MaxCpu         int64  `json:"max_cpu" db:"max_cpu"`
	MinCpu         int64  `json:"min_cpu" db:"min_cpu"`
	MemorySize     int64  `json:"memory_size" db:"memory_size"`
	LogDiskSize    int64  `json:"log_disk_size" db:"log_disk_size"`
	MaxIops        int64  `json:"max_iops" db:"max_iops"`
	MinIops        int64  `json:"min_iops" db:"min_iops"`
	IopsWeight     int64  `json:"iops_weight" db:"iops_weight"`
}
