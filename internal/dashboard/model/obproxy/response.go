/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package obproxy

import (
	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
)

type OBProxy struct {
	OBProxyOverview `json:",inline"`

	ProxySysSecret string                `json:"proxySysSecret" binding:"required"`
	Service        response.K8sService   `json:"service" binding:"required"`
	Resource       common.ResourceSpec   `json:"resource" binding:"required"`
	Parameters     []common.KVPair       `json:"parameters" binding:"required"`
	Pods           []response.K8sPodInfo `json:"pods" binding:"required"`
}

type OBProxyOverview struct {
	Name             string    `json:"name" binding:"required"`
	Namespace        string    `json:"namespace" binding:"required"`
	OBCluster        K8sObject `json:"obCluster" binding:"required"`
	ProxyClusterName string    `json:"proxyClusterName" binding:"required"`
	Image            string    `json:"image" binding:"required"`
	Replicas         int32     `json:"replicas" binding:"required"`
	ServiceType      string    `json:"serviceType" binding:"required"`
	ServiceIP        string    `json:"serviceIp" binding:"required"`
	CreationTime     int64     `json:"creationTime" binding:"required"`
	Status           string    `json:"status" binding:"required" enums:"Running,Pending"`
}

type ConfigItem struct {
	Name         string  `json:"name"`
	Value        string  `json:"value"`
	Info         string  `json:"info"`
	NeedReboot   bool    `json:"needReboot" db:"need_reboot"`
	VisibleLevel string  `json:"visibleLevel" db:"visible_level"`
	Range        *string `json:"range" db:"range"`
	ConfigLevel  *string `json:"configLevel" db:"config_level"`
}
