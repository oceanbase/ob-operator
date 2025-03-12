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

package handler

import (
	"github.com/gin-gonic/gin"
)

// @ID ListK8sClusters
// @Summary list k8s clusters
// @Description list k8s clusters
// @Tags K8sCluster
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]response.K8sCluster}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/k8s/clusters [GET]
// @Security ApiKeyAuth
func ListK8sClusters(c *gin.Context) (any, error) {
	return nil, nil
}

// @ID GetK8sCluster
// @Summary get k8s cluster
// @Description get k8s cluster
// @Tags K8sCluster
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.K8sCluster}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/k8s/clusters/{name} [GET]
// @Security ApiKeyAuth
func GetK8sCluster(c *gin.Context) (any, error) {
	return nil, nil
}

// @ID CreateK8sCluster
// @Summary create k8s cluster
// @Description create k8s cluster
// @Tags K8sCluster
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/k8s/clusters [POST]
// @Security ApiKeyAuth
func CreateK8sCluster(c *gin.Context) (any, error) {
	return nil, nil
}
