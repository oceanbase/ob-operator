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
// @Tags Obcluster
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]response.OBClusterStastistic}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/statistic [GET]
func GetOBClusterStatistic(c *gin.Context) ([]response.OBClusterStastistic, error) {
	// return mock data
	obclusterStastics, err := oceanbase.GetOBClusterStatistic(c)
	if err != nil {
		return nil, err
	}
	return obclusterStastics, nil
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
func ListOBClusters(c *gin.Context) ([]response.OBCluster, error) {
	obclusters, err := oceanbase.ListOBClusters(c)
	if err != nil {
		return nil, err
	}
	return obclusters, nil
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
func GetOBCluster(c *gin.Context) (*response.OBCluster, error) {
	obclusterIdentity := &param.K8sObjectIdentity{}
	err := c.BindUri(obclusterIdentity)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	obcluster, err := oceanbase.GetOBCluster(c, obclusterIdentity)
	if err != nil {
		return nil, err
	}
	return obcluster, nil
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
func CreateOBCluster(c *gin.Context) (any, error) {
	param := &param.CreateOBClusterParam{}
	err := c.Bind(param)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	param.RootPassword, err = crypto.DecryptWithPrivateKey(param.RootPassword)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	logger.Debugf("Create obcluster: %v", param)
	err = oceanbase.CreateOBCluster(c, param)
	if err != nil {
		return nil, err
	}
	return nil, nil
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
func UpgradeOBCluster(c *gin.Context) (any, error) {
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
	err = oceanbase.UpgradeObCluster(c, obclusterIdentity, updateParam)
	if err != nil {
		return nil, err
	}
	return nil, nil
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
func DeleteOBCluster(c *gin.Context) (any, error) {
	obclusterIdentity := &param.K8sObjectIdentity{}
	err := c.BindUri(obclusterIdentity)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	err = oceanbase.DeleteOBCluster(c, obclusterIdentity)
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	return nil, nil
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
func AddOBZone(c *gin.Context) (any, error) {
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
	err = oceanbase.AddOBZone(c, obclusterIdentity, param)
	if err != nil {
		return nil, err
	}
	return nil, nil
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
func ScaleOBServer(c *gin.Context) (any, error) {
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
	err = oceanbase.ScaleOBServer(c, obzoneIdentity, scaleParam)
	if err != nil {
		return nil, err
	}
	return nil, nil
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
func DeleteOBZone(c *gin.Context) (any, error) {
	obzoneIdentity := &param.OBZoneIdentity{}
	err := c.BindUri(obzoneIdentity)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	err = oceanbase.DeleteOBZone(c, obzoneIdentity)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
