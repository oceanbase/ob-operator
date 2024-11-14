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

import "database/sql"

type Tenant struct {
	TenantID         int64  `json:"tenant_id" db:"tenant_id"`
	TenantName       string `json:"tenant_name" db:"tenant_name"`
	PrimaryZone      string `json:"primary_zone" db:"primary_zone"`
	Locality         string `json:"locality" db:"locality"`
	PreviousLocality string `json:"previous_locality" db:"previous_locality"`
	Status           string `json:"status" db:"status"`
	GmtCreate        string `json:"gmt_create" db:"gmt_create"`
}

type Replica struct {
	Type string `json:"type"`
	Num  int    `json:"num"`
	Zone string `json:"zone"`
}

type Pool struct {
	ResourcePoolID int64         `json:"resource_pool_id" db:"resource_pool_id"`
	Name           string        `json:"name" db:"name"`
	UnitNum        int64         `json:"unit_count" db:"unit_count"`
	UnitConfigID   int64         `json:"unit_config_id" db:"unit_config_id"`
	ZoneList       string        `json:"zone_list" db:"zone_list"`
	TenantID       sql.NullInt64 `json:"tenant_id" db:"tenant_id"`
}

type Unit struct {
	UnitID             int64          `json:"unit_id" db:"unit_id"`
	ResourcePoolID     int64          `json:"resource_pool_id" db:"resource_pool_id"`
	Zone               string         `json:"zone" db:"zone"`
	SvrIP              string         `json:"svr_ip" db:"svr_ip"`
	SvrPort            int64          `json:"svr_port" db:"svr_port"`
	MigrateFromSvrIP   sql.NullString `json:"migrate_from_svr_ip" db:"migrate_from_svr_ip"`
	MigrateFromSvrPort sql.NullInt64  `json:"migrate_from_svr_port" db:"migrate_from_svr_port"`
	Status             string         `json:"status" db:"status"`
}

type UnitConfigV4 struct {
	UnitConfigID int64   `json:"unit_config_id" db:"unit_config_id"`
	Name         string  `json:"name" db:"name"`
	MaxCPU       float64 `json:"max_cpu" db:"max_cpu"`
	MinCPU       float64 `json:"min_cpu" db:"min_cpu"`
	MemorySize   int64   `json:"memory_size" db:"memory_size"`
	MaxIops      int64   `json:"max_iops" db:"max_iops"`
	MinIops      int64   `json:"min_iops" db:"min_iops"`
	LogDiskSize  int64   `json:"log_disk_size" db:"log_disk_size"`
	IopsWeight   int64   `json:"iops_weight" db:"iops_weight"`
	Options      string  `json:"options" db:"options"`
}

type ResourceTotal struct {
	CPUTotal  float64 `json:"cpu_total" db:"cpu_capacity"`
	MemTotal  int64   `json:"mem_total" db:"mem_capacity"`
	DiskTotal int64   `json:"disk_total" db:"data_disk_capacity"`
}

type Charset struct {
	Charset string `json:"charset" db:"charset"`
}

type RsJob struct {
	JobID      int64  `json:"job_id" db:"job_id"`
	JobType    string `json:"job_type" db:"job_type"`
	JobStatus  string `json:"job_status" db:"job_status"`
	TenantID   int64  `json:"tenant_id" db:"tenant_id"`
	TenantName string `json:"tenant_name" db:"tenant_name"`
}

type PoolParam struct {
	PoolName string
	ZoneList []string
	// add more properties if needed
}

type TenantSQLParam struct {
	TenantName   string
	PrimaryZone  string
	Locality     string
	PoolList     []string
	Charset      string
	Collate      string
	VariableList string
	UnitNum      int64
}

type PoolSQLParam struct {
	PoolName string
	UnitNum  int64
	ZoneList string
	UnitName string
}

type UnitConfigV4SQLParam struct {
	UnitConfigName string
	MaxCPU         float64
	MinCPU         float64
	MemorySize     int64
	MaxIops        int64
	MinIops        int64
	LogDiskSize    int64
	IopsWeight     int64
}

type TenantAccessPoint struct {
	TenantID   int64  `json:"tenant_id" db:"tenant_id"`
	TenantName string `json:"tenant_name" db:"tenant_name"`
	SvrIP      string `json:"svr_ip" db:"svr_ip"`
	SqlPort    int64  `json:"sql_port" db:"sql_port"`
}

type CreateEmptyStandbyTenantParam struct {
	TenantName    string
	RestoreSource string
	PrimaryZone   string
	Locality      string
	PoolList      []string
}

// Match CDB_OB_LS and CDB_OB_LS_HISTORY
type LSInfo struct {
	LSID int64 `json:"ls_id" db:"ls_id"`
}

// Match GV$OB_LOG_STAT
type LogStat struct {
	LSID     int64 `json:"ls_id" db:"ls_id"`
	BeginLSN int64 `json:"begin_lsn" db:"begin_lsn"`
}
