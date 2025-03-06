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
	"github.com/oceanbase/ob-operator/internal/dashboard/model/k8s"
)

type K8sEvent struct {
	Namespace  string `json:"namespace" binding:"required"`
	Type       string `json:"type" binding:"required"`
	Count      int32  `json:"count" binding:"required"`
	FirstOccur int64  `json:"firstOccur" binding:"required"`
	LastSeen   int64  `json:"lastSeen" binding:"required"`
	Reason     string `json:"reason" binding:"required"`
	Object     string `json:"object" binding:"required"`
	Message    string `json:"message" binding:"required"`
}

type K8sNodeCondition struct {
	Type    string `json:"type" binding:"required"`
	Reason  string `json:"reason" binding:"required"`
	Message string `json:"message" binding:"required"`
}

type K8sNodeInfo struct {
	Name       string             `json:"name" binding:"required"`
	Status     string             `json:"status" binding:"required"`
	Conditions []K8sNodeCondition `json:"conditions" binding:"required"`
	Roles      []string           `json:"roles" binding:"required"`
	Labels     []common.KVPair    `json:"labels" binding:"required"`
	Taints     []k8s.Taint        `json:"taints" binding:"required"`
	Uptime     int64              `json:"uptime" binding:"required"`
	Version    string             `json:"version" binding:"required"`
	InternalIP string             `json:"internalIP" binding:"required"`
	ExternalIP string             `json:"externalIP" binding:"required"`
	OS         string             `json:"os" binding:"required"`
	Kernel     string             `json:"kernel" binding:"required"`
	CRI        string             `json:"cri" binding:"required"`
}

type K8sNodeResource struct {
	CpuTotal    float64 `json:"cpuTotal" binding:"required"`
	CpuUsed     float64 `json:"cpuUsed" binding:"required"`
	CpuFree     float64 `json:"cpuFree" binding:"required"`
	MemoryTotal float64 `json:"memoryTotal" binding:"required"`
	MemoryUsed  float64 `json:"memoryUsed" binding:"required"`
	MemoryFree  float64 `json:"memoryFree" binding:"required"`
}

type K8sNode struct {
	Info     *K8sNodeInfo     `json:"info"`
	Resource *K8sNodeResource `json:"resource"`
}

type Namespace struct {
	Namespace string `json:"namespace" binding:"required"`
	Status    string `json:"status" binding:"required"`
}

type StorageClass struct {
	Name                 string          `json:"name" binding:"required"`
	Provisioner          string          `json:"provisioner" binding:"required"`
	ReclaimPolicy        string          `json:"reclaimPolicy" binding:"required"`
	VolumeBindingMode    string          `json:"volumeBindingMode" binding:"required"`
	AllowVolumeExpansion bool            `json:"allowVolumeExpansion" binding:"required"`
	MountOptions         []string        `json:"mountOptions,omitempty"`
	Parameters           []common.KVPair `json:"parameters,omitempty"`
}

type K8sPodInfo struct {
	Name       string          `json:"name" binding:"required"`
	Namespace  string          `json:"namespace" binding:"required"`
	NodeName   string          `json:"nodeName" binding:"required"`
	Status     string          `json:"status" binding:"required"`
	PodIP      string          `json:"podIP" binding:"required"`
	Message    string          `json:"message" binding:"required"`
	Reason     string          `json:"reason" binding:"required"`
	StartTime  string          `json:"startTime" binding:"required"`
	Containers []ContainerInfo `json:"containers" binding:"required"`
}

type ContainerInfo struct {
	Name         string              `json:"name" binding:"required"`
	Image        string              `json:"image" binding:"required"`
	RestartCount int32               `json:"restartCount" binding:"required"`
	Ready        bool                `json:"ready" binding:"required"`
	StartTime    string              `json:"startTime" binding:"required"`
	Ports        []int32             `json:"ports" binding:"required"`
	Requests     common.ResourceSpec `json:"requests" binding:"required"`
	Limits       common.ResourceSpec `json:"limits" binding:"required"`
}

type K8sService struct {
	Name       string           `json:"name" binding:"required"`
	Namespace  string           `json:"namespace" binding:"required"`
	Type       string           `json:"type" binding:"required"`
	ClusterIP  string           `json:"clusterIP" binding:"required"`
	ExternalIP string           `json:"externalIP"`
	Ports      []K8sServicePort `json:"ports" binding:"required"`
}

type K8sServicePort struct {
	Name       string `json:"name,omitempty"`
	Port       int32  `json:"port" binding:"required"`
	TargetPort int32  `json:"targetPort" binding:"required"`
	NodePort   int32  `json:"nodePort,omitempty"`
}
