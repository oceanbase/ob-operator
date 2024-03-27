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

	"github.com/oceanbase/ob-operator/internal/dashboard/business/oceanbase"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	crypto "github.com/oceanbase/ob-operator/pkg/crypto"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
)

// @ID GetOBClusterStatistic
// @Summary get obcluster statistic
// @Description get obcluster statistic info
// @Tags OBCluster
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]response.OBClusterStastistic}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/statistic [GET]
func GetOBClusterStatistic(c *gin.Context) ([]response.OBClusterStastistic, error) {
	obclusterStastics, err := oceanbase.GetOBClusterStatistic(c)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Get obcluster statistic: %v", obclusterStastics)
	return obclusterStastics, nil
}

// @ID ListOBClusters
// @Summary list obclusters
// @Description list obclusters
// @Tags OBCluster
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]response.OBClusterOverview}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters [GET]
// @Security ApiKeyAuth
func ListOBClusters(c *gin.Context) ([]response.OBClusterOverview, error) {
	obclusters, err := oceanbase.ListOBClusters(c)
	if err != nil {
		return nil, err
	}
	logger.Debugf("List obclusters: %v", obclusters)
	return obclusters, nil
}

// @ID GetOBCluster
// @Summary get obcluster
// @Description get obcluster detailed info
// @Tags OBCluster
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Success 200 object response.APIResponse{data=response.OBCluster}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/namespace/{namespace}/name/{name} [GET]
// @Security ApiKeyAuth
func GetOBCluster(c *gin.Context) (*response.OBCluster, error) {
	obclusterIdentity := &param.K8sObjectIdentity{}
	err := c.BindUri(obclusterIdentity)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	return oceanbase.GetOBCluster(c, obclusterIdentity)
}

// @ID CreateOBCluster
// @Summary create obcluster
// @Description create obcluster
// @Tags OBCluster
// @Accept application/json
// @Produce application/json
// @Param body body param.CreateOBClusterParam true "create obcluster request body"
// @Success 200 object response.APIResponse{data=response.OBCluster}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters [POST]
// @Security ApiKeyAuth
func CreateOBCluster(c *gin.Context) (*response.OBCluster, error) {
	param := &param.CreateOBClusterParam{}
	err := c.Bind(param)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	param.RootPassword, err = crypto.DecryptWithPrivateKey(param.RootPassword)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	logger.Infof("Create obcluster with param: %+v", param)
	return oceanbase.CreateOBCluster(c, param)
}

// @ID UpgradeOBCluster
// @Summary upgrade obcluster
// @Description upgrade obcluster
// @Tags OBCluster
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Param body body param.UpgradeOBClusterParam true "upgrade obcluster request body"
// @Success 200 object response.APIResponse{data=response.OBCluster}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/namespace/{namespace}/name/{name} [POST]
// @Security ApiKeyAuth
func UpgradeOBCluster(c *gin.Context) (*response.OBCluster, error) {
	obclusterIdentity := &param.K8sObjectIdentity{}
	updateParam := &param.UpgradeOBClusterParam{}
	err := c.BindUri(obclusterIdentity)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	err = c.Bind(updateParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	logger.Infof("Upgrade obcluster with param: %+v", updateParam)
	return oceanbase.UpgradeObCluster(c, obclusterIdentity, updateParam)
}

// @ID DeleteOBCluster
// @Summary delete obcluster
// @Description delete obcluster
// @Tags OBCluster
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Success 200 object response.APIResponse{data=bool}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/namespace/{namespace}/name/{name} [DELETE]
// @Security ApiKeyAuth
func DeleteOBCluster(c *gin.Context) (bool, error) {
	obclusterIdentity := &param.K8sObjectIdentity{}
	err := c.BindUri(obclusterIdentity)
	if err != nil {
		return false, httpErr.NewBadRequest(err.Error())
	}
	return oceanbase.DeleteOBCluster(c, obclusterIdentity)
}

// @ID AddOBZone
// @Summary add obzone
// @Description add obzone
// @Tags OBCluster
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Param body body param.ZoneTopology true "add obzone request body"
// @Success 200 object response.APIResponse{data=response.OBCluster}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/namespace/{namespace}/name/{name}/obzones [POST]
// @Security ApiKeyAuth
func AddOBZone(c *gin.Context) (*response.OBCluster, error) {
	obclusterIdentity := &param.K8sObjectIdentity{}
	err := c.BindUri(obclusterIdentity)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	param := &param.ZoneTopology{}
	err = c.Bind(param)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	logger.Infof("Add obzone with param: %+v", param)
	return oceanbase.AddOBZone(c, obclusterIdentity, param)
}

// @ID ScaleOBServer
// @Summary scale observer
// @Description scale observer
// @Tags OBCluster
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Param obzoneName path string true "obzone name"
// @Param body body param.ScaleOBServerParam true "scale observer request body"
// @Success 200 object response.APIResponse{data=response.OBCluster}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/namespace/{namespace}/name/{name}/obzones/{obzoneName}/scale [POST]
// @Security ApiKeyAuth
func ScaleOBServer(c *gin.Context) (*response.OBCluster, error) {
	obzoneIdentity := &param.OBZoneIdentity{}
	err := c.BindUri(obzoneIdentity)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	scaleParam := &param.ScaleOBServerParam{}
	err = c.Bind(scaleParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	if scaleParam.Replicas <= 0 {
		return nil, httpErr.NewBadRequest("Replicas must be greater than 0")
	}
	logger.Infof("Scale observer with param: %+v", scaleParam)
	return oceanbase.ScaleOBServer(c, obzoneIdentity, scaleParam)
}

// @ID DeleteOBZone
// @Summary delete obzone
// @Description delete obzone
// @Tags OBCluster
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Param obzoneName path string true "obzone name"
// @Success 200 object response.APIResponse{data=response.OBCluster}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/namespace/{namespace}/name/{name}/obzones/{obzoneName} [DELETE]
// @Security ApiKeyAuth
func DeleteOBZone(c *gin.Context) (any, error) {
	obzoneIdentity := &param.OBZoneIdentity{}
	err := c.BindUri(obzoneIdentity)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	return oceanbase.DeleteOBZone(c, obzoneIdentity)
}

// @ID ListOBClusterResources
// @Summary list resource usages, the old router ending with /essential-parameters is deprecated
// @Description list resource usages of specific obcluster, such as cpu, memory, storage, etc. The old router ending with /essential-parameters is deprecated
// @Tags OBCluster
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Success 200 object response.APIResponse{data=response.OBClusterResources}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/{namespace}/{name}/resource-usages [GET]
// @Security ApiKeyAuth
func ListOBClusterResources(c *gin.Context) (*response.OBClusterResources, error) {
	obclusterIdentity := &param.K8sObjectIdentity{}
	err := c.BindUri(obclusterIdentity)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	usages, err := oceanbase.GetOBClusterUsages(c, obclusterIdentity)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Get resource usages of obcluster: %v", obclusterIdentity)
	return usages, nil
}
