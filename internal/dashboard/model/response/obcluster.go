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

type OBClusterStatistic struct {
	Status string `json:"status" binding:"required"`
	Count  int    `json:"count" binding:"required"`
}

type OBServer struct {
	Namespace    string     `json:"namespace" binding:"required"`
	Name         string     `json:"name" binding:"required"`
	Status       string     `json:"status" binding:"required"`
	StatusDetail string     `json:"statusDetail" binding:"required"`
	Address      string     `json:"address" binding:"required"`
	Metrics      *OBMetrics `json:"metrics"`
}

type OBZone struct {
	Namespace    string          `json:"namespace" binding:"required"`
	Name         string          `json:"name" binding:"required"`
	Zone         string          `json:"zone" binding:"required"`
	Replicas     int             `json:"replicas" binding:"required"`
	Status       string          `json:"status" binding:"required"`
	StatusDetail string          `json:"statusDetail" binding:"required"`
	RootService  string          `json:"rootService,omitempty"`
	OBServers    []OBServer      `json:"observers,omitempty"`
	NodeSelector []common.KVPair `json:"nodeSelector,omitempty"`

	Tolerations []common.TolerationSpec `json:"tolerations,omitempty"`
	Affinities  []common.AffinitySpec   `json:"affinities,omitempty"`
}

type OBMetrics struct {
	CpuPercent    int `json:"cpuPercent" binding:"required"`
	MemoryPercent int `json:"memoryPercent" binding:"required"`
	DiskPercent   int `json:"diskPercent" binding:"required"`
}

type OBClusterMeta struct {
	UID         string             `json:"uid" binding:"required"`
	Name        string             `json:"name" binding:"required"`
	Namespace   string             `json:"namespace" binding:"required"`
	ClusterName string             `json:"clusterName" binding:"required"`
	ClusterId   int64              `json:"clusterId" binding:"required"`
	Mode        common.ClusterMode `json:"mode" binding:"required"`

	SupportStaticIP    bool `json:"supportStaticIP" binding:"required"`
	DeletionProtection bool `json:"deletionProtection" binding:"required"`
	PvcIndependent     bool `json:"pvcIndependent" binding:"required"`
}

type OBClusterOverview struct {
	OBClusterMeta `json:",inline"`
	Status        string   `json:"status" binding:"required"`
	StatusDetail  string   `json:"statusDetail" binding:"required"`
	CreateTime    int64    `json:"createTime" binding:"required"`
	Image         string   `json:"image" binding:"required"`
	Topology      []OBZone `json:"topology" binding:"required"`
}

type OBCluster struct {
	OBClusterOverview `json:",inline"`

	Metrics *OBMetrics `json:"metrics"`
	Version string     `json:"version"`

	OBClusterExtra `json:",inline"`
}

type ResourceSpecRender struct {
	Cpu    int64 `json:"cpu" binding:"required"`
	Memory int64 `json:"memory" binding:"required"`
}

type ParameterSpec struct {
	Name      string `json:"name" binding:"required"`
	SpecValue string `json:"specValue" binding:"required"`
	Value     string `json:"value" binding:"required"`
}

type OBClusterExtra struct {
	Resource ResourceSpecRender `json:"resource" binding:"required"`
	Storage  OBServerStorage    `json:"storage" binding:"required"`

	RootPasswordSecret string          `json:"rootPasswordSecret" binding:"required"`
	Parameters         []ParameterSpec `json:"parameters" binding:"required"`
	Monitor            *MonitorSpec    `json:"monitor"`
	BackupVolume       *NFSVolumeSpec  `json:"backupVolume"`
	Annotations        []common.KVPair `json:"annotations"`
}

type MonitorSpec struct {
	Image    string             `json:"image" binding:"required"`
	Resource ResourceSpecRender `json:"resource" binding:"required"`
}

type NFSVolumeSpec struct {
	Address string `json:"address" binding:"required"`
	Path    string `json:"path" binding:"required"`
}

type OBServerStorage struct {
	DataStorage    StorageSpec `json:"dataStorage" binding:"required"`
	RedoLogStorage StorageSpec `json:"redoLogStorage" binding:"required"`
	SysLogStorage  StorageSpec `json:"sysLogStorage" binding:"required"`
}

type StorageSpec struct {
	StorageClass string `json:"storageClass" binding:"required"`
	Size         int64  `json:"size" binding:"required"`
}

type OBClusterResources struct {
	MinPoolMemory     int64                               `json:"minPoolMemory" example:"2147483648" binding:"required"`
	OBServerResources []OBServerAvailableResource         `json:"obServerResources"`
	OBZoneResourceMap map[string]*OBZoneAvailableResource `json:"obZoneResourceMap"`
}

type OBServerAvailableResource struct {
	OBServerIP              string `json:"obServerIP" binding:"required"`
	OBZoneAvailableResource `json:",inline"`
}

type OBZoneAvailableResource struct {
	ServerCount       int64  `json:"serverCount" example:"3" binding:"required"`
	OBZone            string `json:"obZone" example:"zone1" binding:"required"`
	AvailableLogDisk  int64  `json:"availableLogDisk" example:"5368709120" binding:"required"`
	AvailableDataDisk int64  `json:"availableDataDisk" example:"16106127360" binding:"required"`
	AvailableMemory   int64  `json:"availableMemory" example:"5368709120" binding:"required"`
	AvailableCPU      int64  `json:"availableCPU" example:"12" binding:"required"`
}
