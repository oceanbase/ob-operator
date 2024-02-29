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

type OBCluster struct {
	Namespace    string     `json:"namespace"`
	Name         string     `json:"name"`
	ClusterName  string     `json:"clusterName"`
	ClusterId    int64      `json:"clusterId"`
	Topology     []OBZone   `json:"topology"`
	Status       string     `json:"status"`
	StatusDetail string     `json:"statusDetail"`
	CreateTime   float64    `json:"createTime"`
	Image        string     `json:"image"`
	Metrics      *OBMetrics `json:"metrics"`
	Version      string     `json:"version"`

	OBClusterExtra `json:",inline"`
}

type OBClusterExtra struct {
	RootPasswordSecret string             `json:"rootPasswordSecret"`
	Parameters         []common.KVPair    `json:"parameters"`
	Monitor            *MonitorSpec       `json:"monitor"`
	BackupVolume       *NFSVolumeSpec     `json:"backupVolume"`
	Mode               common.ClusterMode `json:"mode"`
}

type MonitorSpec struct {
	Image    string              `json:"image"`
	Resource common.ResourceSpec `json:"resource"`
}

type NFSVolumeSpec struct {
	Address string `json:"address"`
	Path    string `json:"path"`
}
