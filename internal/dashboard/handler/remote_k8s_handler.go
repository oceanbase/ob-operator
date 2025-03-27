/*
Copyright (c) 2025 OceanBase
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

	k8sbiz "github.com/oceanbase/ob-operator/internal/dashboard/business/k8s"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/k8s"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	crypto "github.com/oceanbase/ob-operator/pkg/crypto"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
)

// @ID ListRemoteK8sClusters
// @Summary list remote k8s clusters
// @Description list remote k8s clusters
// @Tags K8sCluster
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]k8s.K8sClusterInfo}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/k8s/clusters [GET]
// @Security ApiKeyAuth
func ListRemoteK8sClusters(c *gin.Context) ([]k8s.K8sClusterInfo, error) {
	return k8sbiz.ListRemoteK8sClusters(c)
}

// @ID GetRemoteK8sCluster
// @Summary get remote k8s cluster
// @Description get remote k8s cluster
// @Tags K8sCluster
// @Accept application/json
// @Produce application/json
// @Param name path string true "k8s cluster name"
// @Success 200 object response.APIResponse{data=k8s.K8sClusterInfo}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/k8s/clusters/{name} [GET]
// @Security ApiKeyAuth
func GetRemoteK8sCluster(c *gin.Context) (*k8s.K8sClusterInfo, error) {
	name := c.Param("name")
	return k8sbiz.GetRemoteK8sCluster(c, name)
}

// @ID DeleteRemoteK8sCluster
// @Summary delete remote k8s cluster
// @Description delete remote k8s cluster
// @Tags K8sCluster
// @Accept application/json
// @Produce application/json
// @Param name path string true "k8s cluster name"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/k8s/clusters/{name} [DELETE]
// @Security ApiKeyAuth
func DeleteRemoteK8sCluster(c *gin.Context) (any, error) {
	name := c.Param("name")
	return nil, k8sbiz.DeleteRemoteK8sCluster(c, name)
}

// @ID PatchRemoteK8sCluster
// @Summary put remote k8s cluster
// @Description put remote k8s cluster
// @Tags K8sCluster
// @Accept application/json
// @Produce application/json
// @Param name path string true "k8s cluster name"
// @Param body body k8s.UpdateK8sClusterParam true "update k8s cluster request body"
// @Success 200 object response.APIResponse{data=k8s.K8sClusterInfo}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/k8s/clusters/{name} [PATCH]
// @Security ApiKeyAuth
func PatchRemoteK8sCluster(c *gin.Context) (*k8s.K8sClusterInfo, error) {
	body := &k8s.UpdateK8sClusterParam{}
	err := c.Bind(body)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	encryptedKey := c.GetHeader(HEADER_ENCRYPTED_KEY)
	if encryptedKey != "" {
		key, err := crypto.DecryptWithPrivateKey(encryptedKey)
		if err != nil {
			return nil, httpErr.NewBadRequest(err.Error())
		}
		body.KubeConfig, err = crypto.AESDescrypt(key, body.KubeConfig)
		if err != nil {
			return nil, httpErr.NewBadRequest(err.Error())
		}
	}
	name := c.Param("name")
	return k8sbiz.UpdateRemoteK8sCluster(c, name, body)
}

// @ID CreateRemoteK8sCluster
// @Summary create remote k8s cluster
// @Description create remote k8s cluster
// @Tags K8sCluster
// @Accept application/json
// @Produce application/json
// @Param body body k8s.CreateK8sClusterParam true "create k8s cluster request body"
// @Success 200 object response.APIResponse{data=k8s.K8sClusterInfo}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/k8s/clusters [POST]
// @Security ApiKeyAuth
func CreateRemoteK8sCluster(c *gin.Context) (*k8s.K8sClusterInfo, error) {
	body := &k8s.CreateK8sClusterParam{}
	err := c.Bind(body)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	encryptedKey := c.GetHeader(HEADER_ENCRYPTED_KEY)
	if encryptedKey != "" {
		key, err := crypto.DecryptWithPrivateKey(encryptedKey)
		if err != nil {
			return nil, httpErr.NewBadRequest(err.Error())
		}
		body.KubeConfig, err = crypto.AESDecrypt(key, body.KubeConfig)
		if err != nil {
			return nil, httpErr.NewBadRequest(err.Error())
		}
	}
	return k8sbiz.CreateRemoteK8sCluster(c, body)
}

// @ID ListRemoteK8sEvents
// @Summary list remote k8s event
// @Description list remote k8s events
// @Tags K8sCluster
// @Accept application/json
// @Produce application/json
// @Param name path string true "k8s cluster name"
// @Param objectType query string false "related object types" Enums(OBCLUSTER, OBTENANT, OBBACKUPPOLICY, OBPROXY)
// @Param type query string false "event level" Enums(NORMAL, WARNING)
// @Param name query string false "Object name" string
// @Param namespace query string false "Namespace" string
// @Success 200 object response.APIResponse{data=[]response.K8sEvent}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/k8s/clusters/{name}/events [GET]
// @Security ApiKeyAuth
func ListRemoteK8sEvents(c *gin.Context) ([]response.K8sEvent, error) {
	queryEventParam := &param.QueryEventParam{
		ObjectType: c.Query("objectType"),
		Type:       c.Query("type"),
		Name:       c.Query("name"),
		Namespace:  c.Query("namespace"),
	}
	k8sClusterName := c.Param("name")
	return k8sbiz.ListRemoteK8sClusterEvents(c, k8sClusterName, queryEventParam)
}

// @ID ListRemoteK8sNodes
// @Summary list remote k8s nodes
// @Description list remote k8s nodes
// @Tags K8sCluster
// @Accept application/json
// @Produce application/json
// @Param name path string true "k8s cluster name"
// @Success 200 object response.APIResponse{data=[]response.K8sNode}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/k8s/clusters/{name}/nodes [GET]
// @Security ApiKeyAuth
func ListRemoteK8sNodes(c *gin.Context) ([]response.K8sNode, error) {
	k8sClusterName := c.Param("name")
	return k8sbiz.ListRemoteK8sClusterNodes(c, k8sClusterName)
}

// @ID PutRemoteK8sNodeLabels
// @Summary update remote k8s node labels
// @Description update remote k8s node labels
// @Tags K8sCluster
// @Accept application/json
// @Produce application/json
// @Param clusterName path string true "k8s cluster name"
// @Param nodeName path string true "node name"
// @Param body body param.NodeLabels true "update node labels request body"
// @Success 200 object response.APIResponse{data=response.K8sNode}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/k8s/clusters/{clusterName}/nodes/{nodeName}/labels [PUT]
// @Security ApiKeyAuth
func PutRemoteK8sNodeLabels(c *gin.Context) (*response.K8sNode, error) {
	clusterName := c.Param("clusterName")
	nodeName := c.Param("nodeName")
	nodeLabels := &param.NodeLabels{}
	err := c.Bind(nodeLabels)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	return k8sbiz.UpdateRemoteK8sClusterNodeLabels(c, clusterName, nodeName, nodeLabels.Labels)
}

// @ID PutRemoteK8sNodeTaints
// @Summary update remote k8s node taints
// @Description update remote k8s node taints
// @Tags K8sCluster
// @Accept application/json
// @Produce application/json
// @Param clusterName path string true "k8s cluster name"
// @Param nodeName path string true "node name"
// @Param body body param.NodeTaints true "update node taints request body"
// @Success 200 object response.APIResponse{data=response.K8sNode}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/k8s/clusters/{clusterName}/nodes/{nodeName}/taints [PUT]
// @Security ApiKeyAuth
func PutRemoteK8sNodeTaints(c *gin.Context) (*response.K8sNode, error) {
	clusterName := c.Param("clusterName")
	nodeName := c.Param("nodeName")
	nodeTaints := &param.NodeTaints{}
	err := c.Bind(nodeTaints)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	return k8sbiz.UpdateRemoteK8sClusterNodeTaints(c, clusterName, nodeName, nodeTaints.Taints)
}

// @ID BatchUpdateRemoteK8sNode
// @Summary batch update remote k8s nodes
// @Description batch update remote k8s nodes taints and labels
// @Tags K8sCluster
// @Accept application/json
// @Produce application/json
// @Param name path string true "k8s cluster name"
// @Param body body param.BatchUpdateNodesParam true "batch update nodes request body"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/k8s/clusters/{name}/nodes/update [POST]
// @Security ApiKeyAuth
func BatchUpdateRemoteK8sNode(c *gin.Context) (any, error) {
	clusterName := c.Param("name")
	updateNodesParam := &param.BatchUpdateNodesParam{}
	err := c.Bind(updateNodesParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	return nil, k8sbiz.BatchUpdateRemoteK8sClusterNodes(c, clusterName, updateNodesParam)
}
