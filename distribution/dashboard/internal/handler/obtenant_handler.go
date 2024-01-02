package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/internal/business/oceanbase"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/param"
	logger "github.com/sirupsen/logrus"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// @ID ListAllTenants
// @Tags Obtenant
// @Summary List all tenants
// @Description List all tenants and return them
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenants [GET]
// @Security ApiKeyAuth
func ListAllTenants(c *gin.Context) {
	tenants, err := oceanbase.ListAllOBTenants()
	if err != nil {
		SendInternalServerErrorResponse(c, nil, err)
		return
	}
	SendSuccessfulResponse(c, tenants)
}

// @ID GetTenant
// @Tags Obtenant
// @Summary Get tenant
// @Description Get an obtenant in a specific namespace
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenant/{namespace}/{name} [GET]
// @Security ApiKeyAuth
func GetTenant(c *gin.Context) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	tenant, err := oceanbase.GetOBTenant(types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			SendNotFoundResponse(c, nil, err)
			return
		}
		SendInternalServerErrorResponse(c, nil, err)
		return
	}
	SendSuccessfulResponse(c, tenant)
}

// @ID CreateTenant
// @Tags Obtenant
// @Summary Create tenant
// @Description Create an obtenant in a specific namespace
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.CreateOBTenantParam true "create obtenant request body"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenant/{namespace}/{name} [PUT]
// @Security ApiKeyAuth
func CreateTenant(c *gin.Context) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	tenantParam := &param.CreateOBTenantParam{}
	err = c.BindJSON(tenantParam)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	logger.Infof("Create obtenant: %+v", tenantParam)
	tenant, err := oceanbase.CreateOBTenant(types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	}, tenantParam)
	if err != nil {
		SendInternalServerErrorResponse(c, nil, err)
		return
	}
	SendSuccessfulResponse(c, tenant)
}

// @ID UpdateTenant
// @Tags Obtenant
// @Summary Update tenant
// @Description Update an obtenant in a specific namespace
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.UpdateOBTenantParam true "update obtenant request body, the same as CreateOBTenantParam"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenant/{namespace}/{name} [POST]
// @Security ApiKeyAuth
func UpdateTenant(c *gin.Context) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	tenantParam := &param.CreateOBTenantParam{}
	err = c.BindJSON(tenantParam)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	logger.Infof("Update obtenant: %+v", tenantParam)
	tenant, err := oceanbase.UpdateOBTenant(types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	}, tenantParam)
	if err != nil {
		SendInternalServerErrorResponse(c, nil, err)
		return
	}
	SendSuccessfulResponse(c, tenant)
}

// @ID DeleteTenant
// @Tags Obtenant
// @Summary Delete tenant
// @Description Delete an obtenant in a specific namespace, ask user to confrim the deletion carefully
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenant/{namespace}/{name} [DELETE]
// @Security ApiKeyAuth
func DeleteTenant(c *gin.Context) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	err = oceanbase.DeleteOBTenant(types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	})
	if err != nil {
		logHandlerError(c, err)
		if kubeerrors.IsNotFound(err) {
			SendNotFoundResponse(c, nil, err)
			return
		}
		SendInternalServerErrorResponse(c, nil, err)
		return
	}
	SendSuccessfulResponse(c, nil)
}
