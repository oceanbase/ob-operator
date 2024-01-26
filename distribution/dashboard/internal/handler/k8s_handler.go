package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/internal/business/k8s"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/param"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/response"
	oberr "github.com/oceanbase/oceanbase-dashboard/pkg/errors"
)

// @ID ListK8sEvents
// @Summary list k8s event
// @Description list k8s events
// @Tags Cluster
// @Accept application/json
// @Produce application/json
// @Param objectType query string false "related object types" Enums(OBCLUSTER, OBTENANT)
// @Param type query string false "event level" Enums(NORMAL, WARNING)
// @Param name query string false "Object name" string
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
		logHandlerError(c, err)
		return nil, err
	}
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
func ListK8sNodes(c *gin.Context) ([]response.K8sNode, error) {
	nodes, err := k8s.ListNodes()
	if err != nil {
		logHandlerError(c, err)
		return nil, err
	}
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
func ListK8sNamespaces(c *gin.Context) ([]response.Namespace, error) {
	namespaces, err := k8s.ListNamespaces()
	if err != nil {
		logHandlerError(c, err)
		return nil, err
	}
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
func ListK8sStorageClasses(c *gin.Context) ([]response.StorageClass, error) {
	storageClasses, err := k8s.ListStorageClasses()
	if err != nil {
		logHandlerError(c, err)
		return nil, err
	}
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
		logHandlerError(c, err)
		return nil, oberr.NewBadRequest(err.Error())
	}
	err = k8s.CreateNamespace(param)
	if err != nil {
		logHandlerError(c, err)
		return nil, err
	}
	return nil, nil
}
