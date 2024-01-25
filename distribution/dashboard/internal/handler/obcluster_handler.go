package handler

import (
	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"

	"github.com/oceanbase/oceanbase-dashboard/internal/business/oceanbase"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/param"
	crypto "github.com/oceanbase/oceanbase-dashboard/pkg/crypto"
)

// @ID GetOBClusterStatistic
// @Summary get obcluster statistic
// @Description get obcluster statistic info
// @Tags Obcluster
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]response.OBClusterStastistic}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/statistic [GET]
func GetOBClusterStatistic(c *gin.Context) {
	// return mock data
	obclusterStastics, err := oceanbase.GetOBClusterStatistic(c)
	if err != nil {
		logHandlerError(c, err)
		SendInternalServerErrorResponse(c, nil, err)
	} else {
		SendSuccessfulResponse(c, obclusterStastics)
	}
}

// @ID ListOBClusters
// @Summary list obclusters
// @Description list obclusters
// @Tags Obcluster
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]response.OBCluster}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters [GET]
// @Security ApiKeyAuth
func ListOBClusters(c *gin.Context) {
	obclusters, err := oceanbase.ListOBClusters(c)
	if err != nil {
		logHandlerError(c, err)
		SendInternalServerErrorResponse(c, nil, err)
	} else {
		SendSuccessfulResponse(c, obclusters)
	}
}

// @ID GetOBCluster
// @Summary get obcluster
// @Description get obcluster detailed info
// @Tags Obcluster
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
func GetOBCluster(c *gin.Context) {
	obclusterIdentity := &param.K8sObjectIdentity{}
	err := c.BindUri(obclusterIdentity)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
	}
	obcluster, err := oceanbase.GetOBCluster(c, obclusterIdentity)
	if err != nil {
		logHandlerError(c, err)
		SendInternalServerErrorResponse(c, nil, err)
	} else {
		SendSuccessfulResponse(c, obcluster)
	}
}

// @ID CreateOBCluster
// @Summary create obcluster
// @Description create obcluster
// @Tags Obcluster
// @Accept application/json
// @Produce application/json
// @Param body body param.CreateOBClusterParam true "create obcluster request body"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters [POST]
// @Security ApiKeyAuth
func CreateOBCluster(c *gin.Context) {
	param := &param.CreateOBClusterParam{}
	err := c.Bind(param)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
	}
	param.RootPassword, err = crypto.DecryptWithPrivateKey(param.RootPassword)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
		return
	}
	logger.Debugf("Create obcluster: %v", param)
	err = oceanbase.CreateOBCluster(c, param)
	if err != nil {
		logHandlerError(c, err)
		SendInternalServerErrorResponse(c, nil, err)
	} else {
		SendSuccessfulResponse(c, nil)
	}
}

// @ID UpgradeOBCluster
// @Summary upgrade obcluster
// @Description upgrade obcluster
// @Tags Obcluster
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Param body body param.UpgradeOBClusterParam true "upgrade obcluster request body"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/namespace/{namespace}/name/{name} [POST]
// @Security ApiKeyAuth
func UpgradeOBCluster(c *gin.Context) {
	obclusterIdentity := &param.K8sObjectIdentity{}
	updateParam := &param.UpgradeOBClusterParam{}
	err := c.BindUri(obclusterIdentity)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
	}
	err = c.Bind(updateParam)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
	}
	err = oceanbase.UpgradeObCluster(c, obclusterIdentity, updateParam)
	if err != nil {
		logHandlerError(c, err)
		SendInternalServerErrorResponse(c, nil, err)
	} else {
		SendSuccessfulResponse(c, nil)
	}
}

// @ID DeleteOBCluster
// @Summary delete obcluster
// @Description delete obcluster
// @Tags Obcluster
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/namespace/{namespace}/name/{name} [DELETE]
// @Security ApiKeyAuth
func DeleteOBCluster(c *gin.Context) {
	obclusterIdentity := &param.K8sObjectIdentity{}
	err := c.BindUri(obclusterIdentity)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
	}
	err = oceanbase.DeleteOBCluster(c, obclusterIdentity)
	if err != nil {
		logHandlerError(c, err)
		SendInternalServerErrorResponse(c, nil, err)
	} else {
		SendSuccessfulResponse(c, nil)
	}
}

// @ID AddOBZone
// @Summary add obzone
// @Description add obzone
// @Tags Obcluster
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Param body body param.ZoneTopology true "add obzone request body"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/namespace/{namespace}/name/{name}/obzones [POST]
// @Security ApiKeyAuth
func AddOBZone(c *gin.Context) {
	obclusterIdentity := &param.K8sObjectIdentity{}
	err := c.BindUri(obclusterIdentity)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
	}
	param := &param.ZoneTopology{}
	err = c.Bind(param)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
	}
	err = oceanbase.AddOBZone(c, obclusterIdentity, param)
	if err != nil {
		logHandlerError(c, err)
		SendInternalServerErrorResponse(c, nil, err)
	} else {
		SendSuccessfulResponse(c, nil)
	}
}

// @ID ScaleOBServer
// @Summary scale observer
// @Description scale observer
// @Tags Obcluster
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Param obzoneName path string true "obzone name"
// @Param body body param.ScaleOBServerParam true "scale observer request body"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/namespace/{namespace}/name/{name}/obzones/{obzoneName}/scale [POST]
// @Security ApiKeyAuth
func ScaleOBServer(c *gin.Context) {
	obzoneIdentity := &param.OBZoneIdentity{}
	err := c.BindUri(obzoneIdentity)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
	}
	scaleParam := &param.ScaleOBServerParam{}
	err = c.Bind(scaleParam)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
	}
	err = oceanbase.ScaleOBServer(c, obzoneIdentity, scaleParam)
	if err != nil {
		logHandlerError(c, err)
		SendInternalServerErrorResponse(c, nil, err)
	} else {
		SendSuccessfulResponse(c, nil)
	}
}

// @ID DeleteOBZone
// @Summary delete obzone
// @Description delete obzone
// @Tags Obcluster
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Param obzoneName path string true "obzone name"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/namespace/{namespace}/name/{name}/obzones/{obzoneName} [DELETE]
// @Security ApiKeyAuth
func DeleteOBZone(c *gin.Context) {
	obzoneIdentity := &param.OBZoneIdentity{}
	err := c.BindUri(obzoneIdentity)
	if err != nil {
		SendBadRequestResponse(c, nil, err)
	}
	err = oceanbase.DeleteOBZone(c, obzoneIdentity)
	if err != nil {
		logHandlerError(c, err)
		SendInternalServerErrorResponse(c, nil, err)
	} else {
		SendSuccessfulResponse(c, nil)
	}
}
