package response

import (
	"github.com/oceanbase/oceanbase-dashboard/internal/model/common"
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
}
