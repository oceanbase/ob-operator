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

import (
	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
)

type OBClusterStastistic struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type OBServer struct {
	Namespace    string     `json:"namespace"`
	Name         string     `json:"name"`
	Status       string     `json:"status"`
	StatusDetail string     `json:"statusDetail"`
	Address      string     `json:"address"`
	Metrics      *OBMetrics `json:"metrics"`
}

type OBZone struct {
	Namespace    string          `json:"namespace"`
	Name         string          `json:"name"`
	Zone         string          `json:"zone"`
	Replicas     int             `json:"replicas"`
	Status       string          `json:"status"`
	StatusDetail string          `json:"statusDetail"`
	RootService  string          `json:"rootService,omitempty"`
	OBServers    []OBServer      `json:"observers,omitempty"`
	NodeSelector []common.KVPair `json:"nodeSelector,omitempty"`

	Tolerations []common.KVPair       `json:"tolerations,omitempty"`
	Affinities  []common.AffinitySpec `json:"affinities,omitempty"`
}

type OBMetrics struct {
	CpuPercent    int `json:"cpuPercent"`
	MemoryPercent int `json:"memoryPercent"`
	DiskPercent   int `json:"diskPercent"`
}

type OBClusterBrief struct {
	Name         string   `json:"name"`
	Namespace    string   `json:"namespace"`
	ClusterName  string   `json:"clusterName"`
	ClusterId    int64    `json:"clusterId"`
	Status       string   `json:"status"`
	StatusDetail string   `json:"statusDetail"`
	CreateTime   int64    `json:"createTime"`
	Image        string   `json:"image"`
	Topology     []OBZone `json:"topology"`
}

type OBCluster struct {
	OBClusterBrief `json:",inline"`

	Metrics *OBMetrics `json:"metrics"`
	Version string     `json:"version"`

	OBClusterExtra `json:",inline"`
}

type ResourceSpecRender struct {
	Cpu      int64  `json:"cpu"`
	MemoryGB string `json:"memory"`
}

type OBClusterExtra struct {
	Resource ResourceSpecRender `json:"resource"`
	Storage  OBServerStorage    `json:"storage"`

	RootPasswordSecret string             `json:"rootPasswordSecret"`
	Parameters         []common.KVPair    `json:"parameters"`
	Monitor            *MonitorSpec       `json:"monitor"`
	BackupVolume       *NFSVolumeSpec     `json:"backupVolume"`
	Mode               common.ClusterMode `json:"mode"`
}

type MonitorSpec struct {
	Image    string             `json:"image"`
	Resource ResourceSpecRender `json:"resource"`
}

type NFSVolumeSpec struct {
	Address string `json:"address"`
	Path    string `json:"path"`
}

type OBServerStorage struct {
	DataStorage    StorageSpec `json:"dataStorage"`
	RedoLogStorage StorageSpec `json:"redoLogStorage"`
	SysLogStorage  StorageSpec `json:"sysLogStorage"`
}

type StorageSpec struct {
	StorageClass string `json:"storageClass"`
	SizeGB       string `json:"size"`
}

type OBClusterResources struct {
	MinPoolMemory     int64                              `json:"minPoolMemory" example:"2147483648"`
	OBServerResources []OBServerAvailableResource        `json:"obServerResources"`
	OBZoneResourceMap map[string]*OBZoneAvaiableResource `json:"obZoneResourceMap"`
}

type OBServerAvailableResource struct {
	OBServerIP             string `json:"obServerIP"`
	OBZoneAvaiableResource `json:",inline"`
}

type OBZoneAvaiableResource struct {
	ServerCount       int64  `json:"serverCount" example:"3"`
	OBZone            string `json:"obZone" example:"zone1"`
	AvailableLogDisk  int64  `json:"availableLogDisk" example:"5368709120"`
	AvailableDataDisk int64  `json:"availableDataDisk" example:"16106127360"`
	AvailableMemory   int64  `json:"availableMemory" example:"5368709120"`
	AvailableCPU      int64  `json:"availableCPU" example:"12"`
}
