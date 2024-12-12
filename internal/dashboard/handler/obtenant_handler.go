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
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/clients"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	acbiz "github.com/oceanbase/ob-operator/internal/dashboard/business/ac"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/oceanbase"
	acmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/ac"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	crypto "github.com/oceanbase/ob-operator/pkg/crypto"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

// @ID ListAllTenants
// @Tags OBTenant
// @Summary List all tenants
// @Description List all tenants and return them
// @Accept application/json
// @Produce application/json
// @Param obcluster query string false "obcluster to filter"
// @Param ns query string false "namespace to filter"
// @Success 200 object response.APIResponse{data=[]response.OBTenantOverview}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenants [GET]
// @Security ApiKeyAuth
func ListAllTenants(c *gin.Context) ([]*response.OBTenantOverview, error) {
	selector := ""
	if queryCluster := c.Query("obcluster"); queryCluster != "" {
		selector = fmt.Sprintf("ref-obcluster=%s", queryCluster)
	}
	listOptions := metav1.ListOptions{
		LabelSelector: selector,
	}
	ns := ""
	if queryNs := c.Query("ns"); queryNs != "" {
		ns = queryNs
	}
	tenants, err := oceanbase.ListAllOBTenants(c, ns, listOptions)
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	if len(tenants) == 0 && c.Query("obcluster") != "" {
		allTenants, err := oceanbase.ListAllOBTenants(c, ns, metav1.ListOptions{})
		if err != nil {
			return nil, httpErr.NewInternal(err.Error())
		}
		for i := range allTenants {
			if allTenants[i].ClusterName == c.Query("obcluster") {
				tenants = append(tenants, allTenants[i])
			}
		}
	}
	logger.Debugf("List all tenants: %+v", tenants)
	return tenants, nil
}

// @ID GetTenant
// @Tags OBTenant
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
// @Router /api/v1/obtenants/{namespace}/{name} [GET]
// @Security ApiKeyAuth
func GetTenant(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	tenant, err := oceanbase.GetOBTenant(c, types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, httpErr.NewNotFound(err.Error())
		}
		return nil, httpErr.NewInternal(err.Error())
	}
	logger.Debugf("Get obtenant: %+v", tenant)
	return tenant, nil
}

// @ID CreateTenant
// @Tags OBTenant
// @Summary Create tenant
// @Description Create an obtenant in a specific namespace, passwords should be encrypted by AES
// @Accept application/json
// @Produce application/json
// @Param body body param.CreateOBTenantParam true "create obtenant request body"
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenants [PUT]
// @Security ApiKeyAuth
func CreateTenant(c *gin.Context) (*response.OBTenantDetail, error) {
	tenantParam := &param.CreateOBTenantParam{}
	err := c.BindJSON(tenantParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	username, ok := c.Get("username")
	if !ok {
		return nil, httpErr.NewBadRequest("Unauthorized")
	}
	ok, err = acbiz.Enforce(c, username.(string), &acmodel.Policy{
		Domain: acbiz.DomainOBCluster,
		Object: acmodel.Object(tenantParam.ClusterName),
		Action: acbiz.ActionWrite,
	})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	if !ok {
		return nil, httpErr.New(httpErr.ErrPermissionDenied, "Permission denied")
	}

	tenantParam.RootPassword, err = crypto.DecryptWithPrivateKey(tenantParam.RootPassword)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	if tenantParam.Source != nil && tenantParam.Source.Restore != nil {
		if tenantParam.Source.Restore.Type == "OSS" {
			tenantParam.Source.Restore.OSSAccessID, err = crypto.DecryptWithPrivateKey(tenantParam.Source.Restore.OSSAccessID)
			if err != nil {
				return nil, httpErr.NewBadRequest(err.Error())
			}
			tenantParam.Source.Restore.OSSAccessKey, err = crypto.DecryptWithPrivateKey(tenantParam.Source.Restore.OSSAccessKey)
			if err != nil {
				return nil, httpErr.NewBadRequest(err.Error())
			}
		}
		if tenantParam.Source.Restore.BakEncryptionPassword != "" {
			tenantParam.Source.Restore.BakEncryptionPassword, err = crypto.DecryptWithPrivateKey(tenantParam.Source.Restore.BakEncryptionPassword)
			if err != nil {
				return nil, httpErr.NewBadRequest(err.Error())
			}
		}
	}
	loggingCreateOBTenantParam(tenantParam)
	tenant, err := oceanbase.CreateOBTenant(c, types.NamespacedName{
		Namespace: tenantParam.Namespace,
		Name:      tenantParam.Name,
	}, tenantParam)
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	logger.Debugf("Create obtenant: %+v", tenant)
	return tenant, nil
}

// @ID DeleteTenant
// @Tags OBTenant
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
// @Router /api/v1/obtenants/{namespace}/{name} [DELETE]
// @Security ApiKeyAuth
func DeleteTenant(c *gin.Context) (any, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	err = oceanbase.DeleteOBTenant(c, types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, httpErr.NewNotFound(err.Error())
		}
		return nil, httpErr.NewInternal(err.Error())
	}
	return nil, nil
}

// @ID PatchTenant
// @Tags OBTenant
// @Summary Patch tenant's configuration
// @Description Patch tenant's configuration
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.PatchTenant true "patch tenant body"
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenants/{namespace}/{name} [PATCH]
// @Security ApiKeyAuth
func PatchTenant(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := param.NamespacedName{}
	err := c.BindUri(&nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	patch := param.PatchTenant{}
	err = c.BindJSON(&patch)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	if patch.UnitNumber == nil && patch.UnitConfig == nil {
		return nil, httpErr.NewBadRequest("unitNumber or unitConfig is required")
	}
	logger.Infof("Patch obtenant with param: %+v", patch)
	tenant, err := oceanbase.PatchTenant(c, types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	}, &patch)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Patch obtenant: %+v", tenant)
	return tenant, nil
}

// @ID ChangeUserPassword
// @Tags OBTenant
// @Summary Change root password of specific tenant
// @Description Change root password of specific tenant, encrypted by AES
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.ChangeUserPassword true "new password"
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenants/{namespace}/{name}/userCredentials [POST]
// @Security ApiKeyAuth
func ChangeUserPassword(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	passwordParam := &param.ChangeUserPassword{}
	err = c.BindJSON(passwordParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	logger.Infof("Change obtenant root password")
	if passwordParam.User != "root" {
		return nil, httpErr.NewBadRequest("only root user is supported")
	}
	passwordParam.Password, err = crypto.DecryptWithPrivateKey(passwordParam.Password)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	tenant, err := oceanbase.ModifyOBTenantRootPassword(c, types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	}, passwordParam.Password)

	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	logger.Debugf("Change obtenant root password: %+v", tenant)
	return tenant, nil
}

// @ID ReplayStandbyLog
// @Tags OBTenant
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
// @Router /api/v1/obtenants/{namespace}/{name}/logreplay [POST]
// @Security ApiKeyAuth
func ReplayStandbyLog(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	logReplayParam := &param.ReplayStandbyLog{}
	err = c.BindJSON(logReplayParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	if !logReplayParam.Unlimited && logReplayParam.Timestamp == nil {
		return nil, httpErr.NewBadRequest("timestamp is required if the restore is limited")
	}
	logger.Infof("Replay standby log with param: %+v", logReplayParam)
	tenant, err := oceanbase.ReplayStandbyLog(c, types.NamespacedName{
		Name:      nn.Name,
		Namespace: nn.Namespace,
	}, logReplayParam)
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	logger.Debugf("Replay standby log: %+v", tenant)
	return tenant, nil
}

// @ID UpgradeTenantVersion
// @Tags OBTenant
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
// @Router /api/v1/obtenants/{namespace}/{name}/version [POST]
// @Security ApiKeyAuth
func UpgradeTenantVersion(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	tenant, err := oceanbase.UpgradeTenantVersion(c, types.NamespacedName{
		Name:      nn.Name,
		Namespace: nn.Namespace,
	})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	logger.Debugf("Upgrade tenant version: %+v", tenant)
	return tenant, nil
}

// @ID ChangeTenantRole
// @Tags OBTenant
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
// @Param body body param.ChangeTenantRole true "target role to change to"
// @Router /api/v1/obtenants/{namespace}/{name}/role [POST]
// @Security ApiKeyAuth
func ChangeTenantRole(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	p := param.ChangeTenantRole{}
	err = c.BindJSON(&p)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	if !p.Failover != p.Switchover {
		return nil, httpErr.NewBadRequest("one and only one of failover and switchover can be true")
	}
	logger.Infof("Change tenant role with param: %+v", p)
	tenant, err := oceanbase.ChangeTenantRole(c, types.NamespacedName{
		Name:      nn.Name,
		Namespace: nn.Namespace,
	}, &p)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Change tenant role: %+v", tenant)
	return tenant, nil
}

// @ID CreateBackupPolicy
// @Tags OBTenant
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
// @Router /api/v1/obtenants/{namespace}/{name}/backupPolicy [PUT]
// @Security ApiKeyAuth
func CreateBackupPolicy(c *gin.Context) (*response.BackupPolicy, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	createPolicyParam := &param.CreateBackupPolicy{}
	err = c.BindJSON(createPolicyParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	if createPolicyParam.DestType != param.BackupDestNFS {
		if createPolicyParam.Host == "" {
			return nil, httpErr.NewBadRequest("Host is required for non-NFS type destination")
		}
		createPolicyParam.OSSAccessID, err = crypto.DecryptWithPrivateKey(createPolicyParam.OSSAccessID)
		if err != nil {
			return nil, httpErr.NewBadRequest(err.Error())
		}
		createPolicyParam.OSSAccessKey, err = crypto.DecryptWithPrivateKey(createPolicyParam.OSSAccessKey)
		if err != nil {
			return nil, httpErr.NewBadRequest(err.Error())
		}
		if createPolicyParam.DestType == param.BackupDestCOS && createPolicyParam.AppID == "" {
			return nil, httpErr.NewBadRequest("AppID is required for COS type destination")
		}
		if createPolicyParam.DestType == param.BackupDestS3 && createPolicyParam.Region == "" {
			return nil, httpErr.NewBadRequest("Region is required for S3 type destination")
		}
	}
	if createPolicyParam.BakEncryptionPassword != "" {
		createPolicyParam.BakEncryptionPassword, err = crypto.DecryptWithPrivateKey(createPolicyParam.BakEncryptionPassword)
		if err != nil {
			return nil, httpErr.NewBadRequest(err.Error())
		}
	}
	logger.Infof("Create backup policy with param: %+v", createPolicyParam)
	policy, err := oceanbase.CreateTenantBackupPolicy(c, types.NamespacedName{
		Name:      nn.Name,
		Namespace: nn.Namespace,
	}, createPolicyParam)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Create backup policy: %+v", policy)
	return policy, nil
}

// @ID UpdateBackupPolicy
// @Tags OBTenant
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
// @Router /api/v1/obtenants/{namespace}/{name}/backupPolicy [PATCH]
// @Security ApiKeyAuth
func UpdateBackupPolicy(c *gin.Context) (*response.BackupPolicy, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	updatePolicyParam := &param.UpdateBackupPolicy{}
	err = c.BindJSON(updatePolicyParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	logger.Infof("Update backup policy with param: %+v", updatePolicyParam)
	policy, err := oceanbase.UpdateTenantBackupPolicy(c, types.NamespacedName{
		Name:      nn.Name,
		Namespace: nn.Namespace,
	}, updatePolicyParam)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Update backup policy: %+v", policy)
	return policy, nil
}

// @ID DeleteBackupPolicy
// @Tags OBTenant
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
// @Param force query string false "force delete" default(false)
// @Router /api/v1/obtenants/{namespace}/{name}/backupPolicy [DELETE]
// @Security ApiKeyAuth
func DeleteBackupPolicy(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	err = oceanbase.DeleteTenantBackupPolicy(c, types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	}, c.Query("force") == "true")
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	return nil, nil
}

// @ID GetBackupPolicy
// @Tags OBTenant
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
// @Router /api/v1/obtenants/{namespace}/{name}/backupPolicy [GET]
// @Security ApiKeyAuth
func GetBackupPolicy(c *gin.Context) (*response.BackupPolicy, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	policy, err := oceanbase.GetTenantBackupPolicy(c, types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	logger.Debugf("Get backup policy: %+v", policy)
	return policy, nil
}

// @ID ListBackupJobs
// @Tags OBTenant
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
// @Param type path string true "backup job type" Enums(FULL,INC,CLEAN,ARCHIVE)
// @Param limit query int false "limit" default(10)
// @Router /api/v1/obtenants/{namespace}/{name}/backup/{type}/jobs [GET]
// @Security ApiKeyAuth
func ListBackupJobs(c *gin.Context) ([]*response.BackupJob, error) {
	p := struct {
		Namespace string `uri:"namespace"`
		Name      string `uri:"name"`
		Type      string `uri:"type"`
	}{}
	err := c.BindUri(&p)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	p.Type = strings.ToUpper(p.Type)
	limit := 10
	if c.Query("limit") != "" {
		limit, err = strconv.Atoi(c.Query("limit"))
		if err != nil {
			return nil, httpErr.NewBadRequest(err.Error())
		}
	}
	jobs, err := oceanbase.ListBackupJobs(c, types.NamespacedName{
		Namespace: p.Namespace,
		Name:      p.Name,
	}, p.Type, limit)
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	logger.Debugf("List backup jobs: %+v", jobs)
	return jobs, nil
}

// @ID GetOBTenantStatistic
// @Tags OBTenant
// @Summary List statistics information of tenants
// @Description List statistics information of tenants
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]response.OBTenantStatistic}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenants/statistic [GET]
// @Security ApiKeyAuth
func GetOBTenantStatistic(c *gin.Context) ([]response.OBTenantStatistic, error) {
	tenants, err := oceanbase.GetOBTenantStatistics(c)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Get obtenant statistics: %+v", tenants)
	return tenants, nil
}

// @ID CreateOBTenantPool
// @Tags OBTenant
// @Summary Create obtenant pool
// @Description Create an obtenant pool in a specific namespace
// @Accept application/json
// @Produce application/json
// @Param body body param.TenantPoolSpec true "create obtenant pool request body"
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param zoneName path string true "obzone name"
// @Router /api/v1/obtenants/{namespace}/{name}/pools/{zoneName} [PUT]
// @Security ApiKeyAuth
func CreateOBTenantPool(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := param.TenantPoolName{}
	err := c.BindUri(&nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}

	p := param.TenantPoolSpec{}
	err = c.BindJSON(&p)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	logger.Infof("Create obtenant pool with param: %+v", p)
	return oceanbase.CreateTenantPool(c, nn, &p)
}

// @ID DeleteOBTenantPool
// @Tags OBTenant
// @Summary Delete obtenant pool
// @Description Delete an obtenant pool in a specific namespace
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param zoneName path string true "obzone name"
// @Router /api/v1/obtenants/{namespace}/{name}/pools/{zoneName} [DELETE]
// @Security ApiKeyAuth
func DeleteOBTenantPool(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := param.TenantPoolName{}
	err := c.BindUri(&nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}

	return oceanbase.DeleteTenantPool(c, nn)
}

// @ID PatchOBTenantPool
// @Tags OBTenant
// @Summary Patch obtenant pool
// @Description Patch an obtenant pool in a specific namespace
// @Accept application/json
// @Produce application/json
// @Param body body param.TenantPoolSpec true "patch obtenant pool request body"
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param zoneName path string true "obzone name"
// @Router /api/v1/obtenants/{namespace}/{name}/pools/{zoneName} [PATCH]
// @Security ApiKeyAuth
func PatchOBTenantPool(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := param.TenantPoolName{}
	err := c.BindUri(&nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}

	p := param.TenantPoolSpec{}
	err = c.BindJSON(&p)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	logger.Infof("Patch obtenant pool with param: %+v", p)
	return oceanbase.PatchTenantPool(c, nn, &p)
}

// @ID ListOBTenantRelatedEvents
// @Tags OBTenant
// @Summary List related events of specific tenant
// @Description List related events of specific tenant, including restore, backup and backup policy events
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]response.K8sEvent}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Router /api/v1/obtenants/{namespace}/{name}/related-events [GET]
// @Security ApiKeyAuth
func ListOBTenantRelatedEvents(c *gin.Context) ([]response.K8sEvent, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}

	obtenant, err := clients.TenantClient.Get(c, nn.Namespace, nn.Name, metav1.GetOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, httpErr.NewBadRequest("obtenant not found")
		}
		return nil, httpErr.NewInternal(err.Error())
	}
	events := []response.K8sEvent{}
	// Get related events of obtenant

	restoreList := &v1alpha1.OBTenantRestoreList{}
	err = clients.RestoreJobClient.List(c, nn.Namespace, restoreList, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", oceanbaseconst.LabelTenantName, obtenant.Name),
	})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	if len(restoreList.Items) > 0 {
		names := make([]string, 0, len(restoreList.Items))
		for _, r := range restoreList.Items {
			names = append(names, r.Name)
		}
		events = append(events, GetScopedEvents(c, nn.Namespace, "OBTenantRestore", names)...)
	}

	policy, err := oceanbase.GetTenantBackupPolicy(c, types.NamespacedName{Namespace: nn.Namespace, Name: nn.Name})
	if err != nil {
		if !kubeerrors.IsNotFound(err) {
			return nil, httpErr.NewInternal(err.Error())
		}
	} else if policy != nil {
		policyEvents, err := client.GetClient().ClientSet.CoreV1().Events(policy.Namespace).List(c, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("involvedObject.kind=OBTenantBackupPolicy,involvedObject.name=%s", policy.Name),
		})
		if err != nil {
			return nil, httpErr.NewInternal(err.Error())
		}

		for _, e := range policyEvents.Items {
			events = append(events, response.K8sEvent{
				Namespace:  e.Namespace,
				Reason:     e.Reason,
				Message:    e.Message,
				Type:       e.Type,
				Object:     e.InvolvedObject.Kind + "/" + e.InvolvedObject.Name,
				FirstOccur: e.FirstTimestamp.Unix(),
				LastSeen:   e.LastTimestamp.Unix(),
			})
		}

		backupJobs := &v1alpha1.OBTenantBackupList{}
		err = clients.BackupJobClient.List(c, policy.Namespace, backupJobs, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", oceanbaseconst.LabelRefBackupPolicy, policy.Name),
		})
		if err != nil {
			return nil, httpErr.NewInternal(err.Error())
		}

		if len(backupJobs.Items) > 0 {
			names := make([]string, 0, len(backupJobs.Items))
			for _, b := range backupJobs.Items {
				names = append(names, b.Name)
			}
			events = append(events, GetScopedEvents(c, policy.Namespace, "OBTenantBackup", names)...)
		}
	}

	logger.Debugf("Get related events of obtenant: %v", nn)
	return events, nil
}

// @ID ListOBTenantUsers
// @Tags OBTenant
// @Summary List obtenant users
// @Description List all the users under obtenant
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]string}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Router /api/v1/obtenants/{namespace}/{name}/users [GET]
// @Security ApiKeyAuth
func ListOBTenantUsers(_ *gin.Context) ([]string, error) {
	return nil, httpErr.NewNotImplemented("")
}

// @ID ListOBTenantDatabases
// @Tags OBTenant
// @Summary List obtenant databases
// @Description List all the databases under obtenant
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]string}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Router /api/v1/obtenants/{namespace}/{name}/databases [GET]
// @Security ApiKeyAuth
func ListOBTenantDatabases(_ *gin.Context) ([]string, error) {
	return nil, httpErr.NewNotImplemented("")
}
