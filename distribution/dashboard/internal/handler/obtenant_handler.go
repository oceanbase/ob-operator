package handler

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/internal/business/oceanbase"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/param"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	listOptions := metav1.ListOptions{
		LabelSelector: selector,
	}
	tenants, err := oceanbase.ListAllOBTenants(listOptions)
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
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenant/{namespace}/{name}/unitNumber [PUT]
// @Security ApiKeyAuth
func ModifyUnitNumber(c *gin.Context) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	unitNumberParam := &param.ModifyUnitNumber{}
	err = c.BindJSON(unitNumberParam)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	tenant, err := oceanbase.ModifyOBTenantUnitNumber(types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	}, unitNumberParam.UnitNumber)
	if err != nil {
		SendInternalServerErrorResponse(c, nil, err)
		return
	}
	SendSuccessfulResponse(c, tenant)
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
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenant/{namespace}/{name}/{zone}/unitConfig [PUT]
// @Security ApiKeyAuth
func ModifyUnitConfig(c *gin.Context) {
	nn := struct {
		Name      string `uri:"name"`
		Namespace string `uri:"namespace"`
		Zone      string `uri:"zone"`
	}{}
	err := c.BindUri(&nn)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	unitConfig := param.UnitConfig{}
	err = c.BindJSON(&unitConfig)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	tenant, err := oceanbase.ModifyOBTenantUnitConfig(types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	}, nn.Zone, &unitConfig)
	if err != nil {
		SendInternalServerErrorResponse(c, nil, err)
		return
	}
	SendSuccessfulResponse(c, tenant)
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
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenant/{namespace}/{name}/rootPassword [PUT]
// @Security ApiKeyAuth
func ChangeRootPassword(c *gin.Context) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	passwordParam := &param.ChangeRootPassword{}
	err = c.BindJSON(passwordParam)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}

	tenant, err := oceanbase.ModifyOBTenantRootPassword(types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	}, passwordParam.RootPassword)

	if err != nil {
		SendInternalServerErrorResponse(c, nil, err)
		return
	}
	SendSuccessfulResponse(c, tenant)
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
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenant/{namespace}/{name}/logreplay [POST]
// @Security ApiKeyAuth
func ReplayStandbyLog(c *gin.Context) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	logReplayParam := &param.ReplayStandbyLog{}
	err = c.BindJSON(logReplayParam)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	if logReplayParam.Timestamp == nil {
		SendBadRequestResponse(c, nil, fmt.Errorf("timestamp is required"))
		return
	}
	tenant, err := oceanbase.ReplayStandbyLog(types.NamespacedName{
		Name:      nn.Name,
		Namespace: nn.Namespace,
	}, *logReplayParam.Timestamp)
	if err != nil {
		SendInternalServerErrorResponse(c, nil, err)
		return
	}
	SendSuccessfulResponse(c, tenant)
}

// @ID UpgradeTenantVersion
// @Tags Obtenant
// @Summary Upgrade tenant compatibility version of specific tenant
// @Description Upgrade tenant compatibility version of specific tenant to match the version of cluster
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Router /api/v1/obtenant/{namespace}/{name}/version [POST]
func UpgradeTenantVersion(c *gin.Context) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	tenant, err := oceanbase.UpgradeTenantVersion(types.NamespacedName{
		Name:      nn.Name,
		Namespace: nn.Namespace,
	})
	if err != nil {
		SendInternalServerErrorResponse(c, nil, err)
		return
	}
	SendSuccessfulResponse(c, tenant)
}

// @ID ChangeTenantRole
// @Tags Obtenant
// @Summary Change tenant role of specific tenant
// @Description Change tenant role of specific tenant, if a tenant is a standby tenant, it can be changed to primary tenant, vice versa
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Router /api/v1/obtenant/{namespace}/{name}/role [POST]
func ChangeTenantRole(c *gin.Context) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	p := param.ChangeTenantRole{}
	err = c.BindJSON(&p)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	tenant, err := oceanbase.ChangeTenantRole(types.NamespacedName{
		Name:      nn.Name,
		Namespace: nn.Namespace,
	}, &p)
	if err != nil {
		if oceanbase.Is(err, oceanbase.ErrorTypeBadRequest) {
			SendBadRequestResponse(c, nil, err)
			return
		} else {
			SendInternalServerErrorResponse(c, nil, err)
			return
		}
	}
	SendSuccessfulResponse(c, tenant)
}

// @ID CreateBackupPolicy
// @Tags Obtenant
// @Summary Create backup policy of specific tenant
// @Description Create backup policy of specific tenant, passwords should be encrypted by AES
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.BackupPolicy}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.CreateBackupPolicy true "create backup policy request body"
// @Router /api/v1/obtenant/{namespace}/{name}/backupPolicy [PUT]
func CreateBackupPolicy(c *gin.Context) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	createPolicyParam := &param.CreateBackupPolicy{}
	err = c.BindJSON(createPolicyParam)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	policy, err := oceanbase.CreateTenantBackupPolicy(types.NamespacedName{
		Name:      nn.Name,
		Namespace: nn.Namespace,
	}, createPolicyParam)
	if err != nil {
		if oceanbase.Is(err, oceanbase.ErrorTypeBadRequest) {
			SendBadRequestResponse(c, nil, err)
			return
		} else {
			SendInternalServerErrorResponse(c, nil, err)
			return
		}
	}
	SendSuccessfulResponse(c, policy)
}

// @ID UpdateBackupPolicy
// @Tags Obtenant
// @Summary Update backup policy of specific tenant
// @Description Update backup policy of specific tenant
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.BackupPolicy}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.UpdateBackupPolicy true "update backup policy request body"
// @Router /api/v1/obtenant/{namespace}/{name}/backupPolicy [POST]
func UpdateBackupPolicy(c *gin.Context) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	updatePolicyParam := &param.UpdateBackupPolicy{}
	err = c.BindJSON(updatePolicyParam)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	policy, err := oceanbase.UpdateTenantBackupPolicy(types.NamespacedName{
		Name:      nn.Name,
		Namespace: nn.Namespace,
	}, updatePolicyParam)
	if err != nil {
		if oceanbase.Is(err, oceanbase.ErrorTypeBadRequest) {
			SendBadRequestResponse(c, nil, err)
			return
		} else {
			SendInternalServerErrorResponse(c, nil, err)
			return
		}
	}
	SendSuccessfulResponse(c, policy)
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
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	err = oceanbase.DeleteTenantBackupPolicy(types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	})
	if err != nil {
		SendInternalServerErrorResponse(c, nil, err)
		return
	}
	SendSuccessfulResponse(c, nil)
}

// @ID GetBackupPolicy
// @Tags Obtenant
// @Summary Get backup policy of specific tenant
// @Description Get backup policy of specific tenant
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.BackupPolicy}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Router /api/v1/obtenant/{namespace}/{name}/backupPolicy [GET]
func GetBackupPolicy(c *gin.Context) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	resp, err := oceanbase.GetTenantBackupPolicy(types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	})
	if err != nil {
		SendInternalServerErrorResponse(c, nil, err)
		return
	}
	SendSuccessfulResponse(c, resp)
}

// @ID ListBackupJobs
// @Tags Obtenant
// @Summary List backup jobs of specific tenant
// @Description List backup jobs of specific tenant
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]response.BackupJob}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param type path string true "backup job type" Enums(FULL,INCR,CLEAN,ARCHIVE)
// @Param limit query int false "limit" default(10)
// @Router /api/v1/obtenant/{namespace}/{name}/backup/{type}/jobs [GET]
func ListBackupJobs(c *gin.Context) {
	p := struct {
		Namespace string `uri:"namespace"`
		Name      string `uri:"name"`
		Type      string `uri:"type"`
	}{}
	err := c.BindUri(&p)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	limit := 10
	if c.Query("limit") != "" {
		limit, err = strconv.Atoi(c.Query("limit"))
		if err != nil {
			SendBadRequestResponse(c, nil, err)
			return
		}
	}
	jobs, err := oceanbase.ListBackupJobs(types.NamespacedName{
		Namespace: p.Namespace,
		Name:      p.Name,
	}, p.Type, limit)
	if err != nil {
		SendInternalServerErrorResponse(c, nil, err)
		return
	}
	SendSuccessfulResponse(c, jobs)
}
