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
	"context"

	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/clients"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/oceanbase"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/job"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

// @ID GetOBClusterStatistic
// @Summary get obcluster statistic
// @Description get obcluster statistic info
// @Tags OBCluster
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]response.OBClusterStatistic}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/statistic [GET]
func GetOBClusterStatistic(c *gin.Context) ([]response.OBClusterStatistic, error) {
	obclusterStatistics, err := oceanbase.GetOBClusterStatistic(c)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Get obcluster statistic: %v", obclusterStatistics)
	return obclusterStatistics, nil
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
	err = extractPassword(param)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	loggingCreateOBClusterParam(param)
	return nil, oceanbase.CreateOBCluster(c, param)
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

// @ID ListOBClusterRelatedEvents
// @Summary list related events
// @Description list related events of specific obcluster, including obzone and observer.
// @Tags OBCluster
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Success 200 object response.APIResponse{data=[]response.K8sEvent}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/{namespace}/{name}/related-events [GET]
// @Security ApiKeyAuth
func ListOBClusterRelatedEvents(c *gin.Context) ([]response.K8sEvent, error) {
	nn := &param.K8sObjectIdentity{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	obcluster, err := clients.ClusterClient.Get(c, nn.Namespace, nn.Name, metav1.GetOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, httpErr.NewBadRequest("obcluster not found")
		}
		return nil, httpErr.NewInternal(err.Error())
	}
	obzoneList := &v1alpha1.OBZoneList{}
	err = clients.ZoneClient.List(c, nn.Namespace, obzoneList, metav1.ListOptions{
		LabelSelector: oceanbaseconst.LabelRefOBCluster + "=" + obcluster.Name,
	})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	observerList := &v1alpha1.OBServerList{}
	err = clients.ServerClient.List(c, nn.Namespace, observerList, metav1.ListOptions{
		LabelSelector: oceanbaseconst.LabelRefOBCluster + "=" + obcluster.Name,
	})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	var events []response.K8sEvent

	if len(obzoneList.Items) > 0 {
		names := make([]string, 0, len(obzoneList.Items))
		for _, obzone := range obzoneList.Items {
			names = append(names, obzone.Name)
		}
		events = append(events, GetScopedEvents(c, nn.Namespace, "OBZone", names)...)
	}

	if len(observerList.Items) > 0 {
		names := make([]string, 0, len(observerList.Items))
		for _, obzone := range observerList.Items {
			names = append(names, obzone.Name)
		}
		events = append(events, GetScopedEvents(c, nn.Namespace, "OBServer", names)...)
		events = append(events, GetScopedEvents(c, nn.Namespace, "Pod", names)...)
	}

	logger.Debugf("Get related events of obcluster: %v", nn)
	return events, nil
}

func GetScopedEvents(ctx context.Context, ns, kind string, scoped []string) []response.K8sEvent {
	eventList, err := client.GetClient().ClientSet.CoreV1().Events(ns).List(ctx, metav1.ListOptions{
		FieldSelector: "involvedObject.kind=" + kind,
	})
	if err != nil {
		return nil
	}
	existMapping := make(map[string]struct{})
	for _, item := range scoped {
		existMapping[item] = struct{}{}
	}
	var events []response.K8sEvent
	for _, event := range eventList.Items {
		if _, ok := existMapping[event.InvolvedObject.Name]; ok {
			events = append(events, response.K8sEvent{
				Namespace:  event.Namespace,
				Message:    event.Message,
				Reason:     event.Reason,
				Type:       event.Type,
				Object:     event.InvolvedObject.Kind + "/" + event.InvolvedObject.Name,
				Count:      event.Count,
				FirstOccur: event.FirstTimestamp.Unix(),
				LastSeen:   event.LastTimestamp.Unix(),
			})
		}
	}
	return events
}

// @ID PatchOBCluster
// @Summary patch obcluster
// @Description patch obcluster configuration including resources, storage, monitor and parameters
// @Tags OBCluster
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Param body body param.PatchOBClusterParam true "patch obcluster request body"
// @Success 200 object response.APIResponse{data=response.OBCluster}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/namespace/{namespace}/name/{name} [PATCH]
// @Security ApiKeyAuth
func PatchOBCluster(c *gin.Context) (*response.OBCluster, error) {
	obclusterIdentity := &param.K8sObjectIdentity{}
	err := c.BindUri(obclusterIdentity)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	patchParam := &param.PatchOBClusterParam{}
	err = c.Bind(patchParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	logger.Infof("Patch obcluster with param: %+v", patchParam)
	return oceanbase.PatchOBCluster(c, obclusterIdentity, patchParam)
}

// @ID RestartOBServers
// @Summary restart observers
// @Description restart specified observers in the obcluster
// @Tags OBCluster
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Param body body param.RestartOBServersParam true "restart observers request body"
// @Success 200 object response.APIResponse{data=response.OBCluster}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/namespace/{namespace}/name/{name}/restart [POST]
// @Security ApiKeyAuth
func RestartOBServers(c *gin.Context) (*response.OBCluster, error) {
	obclusterIdentity := &param.K8sObjectIdentity{}
	err := c.BindUri(obclusterIdentity)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	restartParam := &param.RestartOBServersParam{}
	err = c.Bind(restartParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	logger.Infof("Restart observers with param: %+v", restartParam)
	return oceanbase.RestartOBServers(c, obclusterIdentity, restartParam)
}

// @ID DeleteOBServers
// @Summary delete observers
// @Description delete specified observers from the obcluster
// @Tags OBCluster
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Param body body param.DeleteOBServersParam true "delete observers request body"
// @Success 200 object response.APIResponse{data=response.OBCluster}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/namespace/{namespace}/name/{name}/observers [DELETE]
// @Security ApiKeyAuth
func DeleteOBServers(c *gin.Context) (*response.OBCluster, error) {
	obclusterIdentity := &param.K8sObjectIdentity{}
	err := c.BindUri(obclusterIdentity)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	deleteParam := &param.DeleteOBServersParam{}
	err = c.Bind(deleteParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	logger.Infof("Delete observers with param: %+v", deleteParam)
	return oceanbase.DeleteOBServers(c, obclusterIdentity, deleteParam)
}

// @ID ListOBClusterParameters
// @Summary List OBCluster Parameters
// @Description List OBCluster Parameters by namespace and name
// @Tags OBCluster
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "namespace of obcluster resource"
// @Param name path string true "name of obcluster resource"
// @Success 200 object response.APIResponse{data=[]response.AggregatedParameter}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/namespace/{namespace}/name/{name}/parameters [GET]
func ListOBClusterParameters(c *gin.Context) ([]response.AggregatedParameter, error) {
	nn := &param.K8sObjectIdentity{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	return oceanbase.ListOBClusterParameters(c, nn)
}

// @ID DownloadOBClusterLog
// @Summary Download obcluster log
// @Description Download obcluster log
// @Tags OBCluster
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Param startTime query string true "start time"
// @Param endTime query string true "end time"
// @Success 200 object response.APIResponse{data=job.Job}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/namespace/{namespace}/name/{name}/log [GET]
// @Security ApiKeyAuth
func DownloadOBClusterLog(c *gin.Context) (*job.Job, error) {
	nn := &param.K8sObjectIdentity{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	startTime := c.Query("startTime")
	endTime := c.Query("endTime")
	if startTime == "" || endTime == "" {
		return nil, httpErr.NewBadRequest("startTime and endTime are required")
	}
	return oceanbase.DownloadOBClusterLog(c, nn.Namespace, nn.Name, startTime, endTime)
}
