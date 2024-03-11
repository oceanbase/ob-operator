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

// use as param
type ServerInfo struct {
	Ip   string
	Port int64
}

// use as response
type OBServer struct {
	Id               int64  `json:"id" db:"id"`
	Zone             string `json:"zone" db:"zone"`
	Ip               string `json:"svr_ip" db:"svr_ip"`
	Port             int64  `json:"svr_port" db:"svr_port"`
	SqlPort          int64  `json:"inner_port" db:"inner_port"`
	WithRootserver   int64  `json:"with_rootserver" db:"with_rootserver"`
	WithPartition    int64  `json:"with_partition" db:"with_partition"`
	Status           string `json:"status" db:"status"`
	StartServiceTime int64  `json:"start_service_time" db:"start_service_time"`
	BuildVersion     string `json:"build_version" db:"build_version"`
}

// GVOBServer shows the usage info of the server
type GVOBServer struct {
	ServerIP             string `json:"svrIp" db:"svr_ip"`
	Port                 int64  `json:"svrPort" db:"svr_port"`
	Zone                 string `json:"zone" db:"zone"`
	SQLPort              int64  `json:"sqlPort" db:"sql_port"`
	CPUCapacity          int64  `json:"cpuCapacity" db:"cpu_capacity"`
	CPUCapacityMax       int64  `json:"cpuCapacityMax" db:"cpu_capacity_max"`
	CPUAssigned          int64  `json:"cpuAssigned" db:"cpu_assigned"`
	CPUAssignedMax       int64  `json:"cpuAssignedMax" db:"cpu_assigned_max"`
	MemCapacity          int64  `json:"memCapacity" db:"mem_capacity"`
	MemAssigned          int64  `json:"memAssigned" db:"mem_assigned"`
	MemoryLimit          int64  `json:"memoryLimit" db:"memory_limit"`
	LogDiskCapacity      int64  `json:"logDiskCapacity" db:"log_disk_capacity"`
	LogDiskAssigned      int64  `json:"logDiskAssigned" db:"log_disk_assigned"`
	DataDiskCapacity     int64  `json:"dataDiskCapacity" db:"data_disk_capacity"`
	DataDiskAllocated    int64  `json:"dataDiskAllocated" db:"data_disk_allocated"`
	DataDiskInUse        int64  `json:"dataDiskInUse" db:"data_disk_in_use"`
	DataDiskHealthStatus string `json:"dataDiskHealthStatus" db:"data_disk_health_status"`
}
