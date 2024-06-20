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

import "github.com/oceanbase/ob-operator/internal/dashboard/model/common"

type K8sObject struct {
	Namespace string `json:"namespace" uri:"namespace" binding:"required"`
	Name      string `json:"name" uri:"name" binding:"required"`
}

type CreateOBProxyParam struct {
	Name             string    `json:"name" binding:"required"`
	Namespace        string    `json:"namespace" binding:"required"`
	ProxyClusterName string    `json:"proxyClusterName" binding:"required"`
	OBCluster        K8sObject `json:"obCluster" binding:"required"`

	// Password should be encrypted
	ProxySysPassword string              `json:"proxySysPassword" binding:"required"`
	Image            string              `json:"image" binding:"required"`
	ServiceType      string              `json:"serviceType" binding:"required" example:"ClusterIP" enums:"ClusterIP,NodePort,LoadBalancer,ExternalName" default:"ClusterIP"`
	Replicas         int32               `json:"replicas" binding:"required"`
	Resource         common.ResourceSpec `json:"resource" binding:"required"`
	Parameters       []common.KVPair     `json:"parameters"`
}

type PatchOBProxyParam struct {
	Image       *string              `json:"image,omitempty"`
	ServiceType *string              `json:"serviceType,omitempty" example:"ClusterIP" enums:"ClusterIP,NodePort,LoadBalancer,ExternalName" default:"ClusterIP"`
	Replicas    *int32               `json:"replicas,omitempty"`
	Resource    *common.ResourceSpec `json:"resource,omitempty"`
	Parameters  []common.KVPair      `json:"parameters,omitempty"`
}
