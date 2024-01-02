package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/internal/business/k8s"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/param"
)

// @ID ListK8sEvents
// @Summary list k8s event
// @Description list k8s events
// @Tags Cluster
// @Accept application/json
// @Produce application/json
// @Param objectType query string false "related object types" Enums(OBCLUSTER, OBTENANT)
// @Param type query string false "event level" Enums(NORMAL, WARNING)
// @Success 200 object response.APIResponse{data=[]response.K8sEvent}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/cluster/events [GET]
// @Security ApiKeyAuth
func ListK8sEvents(c *gin.Context) {
	queryEventParam := &param.QueryEventParam{
		ObjectType: c.Query("objectType"),
		Type:       c.Query("type"),
	}
	events, err := k8s.ListEvents(queryEventParam)
	if err != nil {
		logHandlerError(c, err)
		SendInternalServerErrorResponse(c, nil, err)
	} else {
		SendSuccessfulResponse(c, events)
	}
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
func ListK8sNodes(c *gin.Context) {
	nodes, err := k8s.ListNodes()
	if err != nil {
		logHandlerError(c, err)
		SendInternalServerErrorResponse(c, nil, err)
	} else {
		SendSuccessfulResponse(c, nodes)
	}
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
func ListK8sNamespaces(c *gin.Context) {
	namespaces, err := k8s.ListNamespaces()
	if err != nil {
		logHandlerError(c, err)
		SendInternalServerErrorResponse(c, nil, err)
	} else {
		SendSuccessfulResponse(c, namespaces)
	}
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
func ListK8sStorageClasses(c *gin.Context) {
	storageClasses, err := k8s.ListStorageClasses()
	if err != nil {
		logHandlerError(c, err)
		SendInternalServerErrorResponse(c, nil, err)
	} else {
		SendSuccessfulResponse(c, storageClasses)
	}
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
func CreateK8sNamespace(c *gin.Context) {
	param := &param.CreateNamespaceParam{}
	err := c.Bind(param)
	if err != nil {
		logHandlerError(c, err)
		SendBadRequestResponse(c, nil, err)
	} else {
		err = k8s.CreateNamespace(param)
		if err != nil {
			logHandlerError(c, err)
			SendInternalServerErrorResponse(c, nil, err)
		} else {
			SendSuccessfulResponse(c, nil)
		}
	}
}
