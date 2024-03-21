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
	logger "github.com/sirupsen/logrus"

	"github.com/oceanbase/ob-operator/internal/dashboard/business/k8s"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
)

// @ID ListK8sEvents
// @Summary list k8s event
// @Description list k8s events
// @Tags Cluster
// @Accept application/json
// @Produce application/json
// @Param objectType query string false "related object types" Enums(OBCLUSTER, OBTENANT, OBBACKUPPOLICY)
// @Param type query string false "event level" Enums(NORMAL, WARNING)
// @Param name query string false "Object name" string
// @Param namespace query string false "Namespace" string
// @Success 200 object response.APIResponse{data=[]response.K8sEvent}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/cluster/events [GET]
// @Security ApiKeyAuth
func ListK8sEvents(c *gin.Context) ([]response.K8sEvent, error) {
	queryEventParam := &param.QueryEventParam{
		ObjectType: c.Query("objectType"),
		Type:       c.Query("type"),
		Name:       c.Query("name"),
		Namespace:  c.Query("namespace"),
	}
	events, err := k8s.ListEvents(queryEventParam)
	if err != nil {
		return nil, err
	}
	logger.Debugf("List k8s events: %v", events)
	return events, nil
}

// @ID ListK8sNodes
// @Summary list k8s nodes
// @Description list k8s nodes
// @Tags Cluster
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]response.K8sNode}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/cluster/nodes [GET]
// @Security ApiKeyAuth
func ListK8sNodes(_ *gin.Context) ([]response.K8sNode, error) {
	nodes, err := k8s.ListNodes()
	if err != nil {
		return nil, err
	}
	logger.Debugf("List k8s nodes: %v", nodes)
	return nodes, nil
}

// @ID ListK8sNamespaces
// @Summary list k8s namespaces
// @Description list k8s namespaces
// @Tags Cluster
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]response.Namespace}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/cluster/namespaces [GET]
// @Security ApiKeyAuth
func ListK8sNamespaces(_ *gin.Context) ([]response.Namespace, error) {
	namespaces, err := k8s.ListNamespaces()
	if err != nil {
		return nil, err
	}
	logger.Debugf("List k8s namespaces: %v", namespaces)
	return namespaces, nil
}

// @ID ListK8sStorageClasses
// @Summary list k8s storage classes
// @Description list k8s storage classes
// @Tags Cluster
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]response.StorageClass}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/cluster/storageClasses [GET]
// @Security ApiKeyAuth
func ListK8sStorageClasses(_ *gin.Context) ([]response.StorageClass, error) {
	storageClasses, err := k8s.ListStorageClasses()
	if err != nil {
		return nil, err
	}
	logger.Debugf("List k8s storage classes: %v", storageClasses)
	return storageClasses, nil
}

// @ID CreateK8sNamespace
// @Summary create k8s namespace
// @Description create k8s namespace
// @Tags Cluster
// @Accept application/json
// @Produce application/json
// @Param body body param.CreateNamespaceParam true "create obcluster request body"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/cluster/namespaces [POST]
// @Security ApiKeyAuth
func CreateK8sNamespace(c *gin.Context) (any, error) {
	param := &param.CreateNamespaceParam{}
	err := c.Bind(param)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	logger.Infof("Create k8s namespace: %+v", param)
	err = k8s.CreateNamespace(param)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
