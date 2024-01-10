package handler

import (
	"fmt"

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
// @Param obcluster query string false "obcluster to filter"
// @Success 200 object response.APIResponse{data=[]response.OBTenantBrief}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenants [GET]
// @Security ApiKeyAuth
func ListAllTenants(c *gin.Context) {
	selector := ""
	if c.Query("obcluster") != "" {
		selector = fmt.Sprintf("ref-obcluster=%s", c.Query("obcluster"))
	}
	tenants, err := oceanbase.ListAllOBTenants(selector)
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
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
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
// @Description Create an obtenant in a specific namespace, passwords should be encrypted by AES
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.CreateOBTenantParam true "create obtenant request body"
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
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

// @ID ModifyUnitNumber
// @Tags Obtenant
// @Summary Modify unit number of specific tenant
// @Description Modify unit number of specific tenant
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.ModifyUnitNumber true "param containing unit number to modify"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenant/{namespace}/{name}/unitNumber [PUT]
// @Security ApiKeyAuth
func ModifyUnitNumber(c *gin.Context) {
	SendNotImplementedResponse(c, nil, nil)
}

// @ID ModifyUnitConfig
// @Tags Obtenant
// @Summary Modify unit config of specific tenant
// @Description Modify unit config of specific tenant
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param zone path string true "target zone"
// @Param body body param.UnitConfig true "new unit config"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenant/{namespace}/{name}/{zone}/unitConfig [PUT]
// @Security ApiKeyAuth
func ModifyUnitConfig(c *gin.Context) {
	SendNotImplementedResponse(c, nil, nil)
}

// @ID ChangeRootPassword
// @Tags Obtenant
// @Summary Change root password of specific tenant
// @Description Change root password of specific tenant, encrypted by AES
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.ChangeRootPassword true "new password"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenant/{namespace}/{name}/rootPassword [PUT]
// @Security ApiKeyAuth
func ChangeRootPassword(c *gin.Context) {
	SendNotImplementedResponse(c, nil, nil)
}

// @ID ReplayStandbyLog
// @Tags Obtenant
// @Summary Replay standby log of specific standby tenant
// @Description Replay standby log of specific standby tenant
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.ReplayStandbyLog true "target timestamp to replay to"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenant/{namespace}/{name}/logreplay [POST]
// @Security ApiKeyAuth
func ReplayStandbyLog(c *gin.Context) {
	SendNotImplementedResponse(c, nil, nil)
}

// @ID ChangeTenantRole
// @Tags Obtenant
// @Summary Change tenant role of specific tenant
// @Description Change tenant role of specific tenant, if a tenant is a standby tenant, it can be changed to primary tenant, vice versa
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Router /api/v1/obtenant/{namespace}/{name}/role [POST]
func ChangeTenantRole(c *gin.Context) {
	SendNotImplementedResponse(c, nil, nil)
}

// @ID UpgradeTenantVersion
// @Tags Obtenant
// @Summary Upgrade tenant compatibility version of specific tenant
// @Description Upgrade tenant compatibility version of specific tenant to match the version of cluster
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Router /api/v1/obtenant/{namespace}/{name}/version [POST]
func UpgradeTenantVersion(c *gin.Context) {
	SendNotImplementedResponse(c, nil, nil)
}

// @ID CreateBackupPolicy
// @Tags Obtenant
// @Summary Create backup policy of specific tenant
// @Description Create backup policy of specific tenant, passwords should be encrypted by AES
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.CreateBackupPolicy true "create backup policy request body"
// @Router /api/v1/obtenant/{namespace}/{name}/backupPolicy [PUT]
func CreateBackupPolicy(c *gin.Context) {
	SendNotImplementedResponse(c, nil, nil)
}

// @ID UpdateBackupPolicy
// @Tags Obtenant
// @Summary Update backup policy of specific tenant
// @Description Update backup policy of specific tenant
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.UpdateBackupPolicy true "update backup policy request body"
// @Router /api/v1/obtenant/{namespace}/{name}/backupPolicy [POST]
func UpdateBackupPolicy(c *gin.Context) {
	SendNotImplementedResponse(c, nil, nil)
}

// @ID DeleteBackupPolicy
// @Tags Obtenant
// @Summary Delete backup policy of specific tenant
// @Description Delete backup policy of specific tenant
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Router /api/v1/obtenant/{namespace}/{name}/backupPolicy [DELETE]
func DeleteBackupPolicy(c *gin.Context) {
	SendNotImplementedResponse(c, nil, nil)
}
