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
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/metric"
	metricconst "github.com/oceanbase/ob-operator/internal/dashboard/business/metric/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	"github.com/oceanbase/ob-operator/internal/oceanbase"
	"github.com/oceanbase/ob-operator/internal/store"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	"github.com/oceanbase/ob-operator/internal/telemetry/models"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

// @ID ListAllMetrics
// @Summary list all metrics
// @Description list all metrics meta info, return by groups
// @Tags Metric
// @Accept application/json
// @Produce application/json
// @Param scope query string true "metrics scope" Enums(OBCLUSTER, OBTENANT)
// @Success 200 object response.APIResponse{data=[]response.MetricClass}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/metrics [GET]
// @Security ApiKeyAuth
func ListMetricMetas(c *gin.Context) ([]response.MetricClass, error) {
	// return mock data
	language := c.GetHeader("Accept-Language")
	scope := c.Query("scope")
	if scope != metricconst.ScopeCluster && scope != metricconst.ScopeTenant && scope != metricconst.ScopeClusterOverview {
		err := errors.New("invalid scope")
		return nil, httpErr.NewBadRequest(err.Error())
	}
	metricClasses, err := metric.ListMetricClasses(scope, language)
	if err != nil {
		return nil, err
	}
	return metricClasses, nil
}

// @ID QueryMetrics
// @Summary query metrics
// @Description query metric data
// @Tags Metric
// @Accept application/json
// @Produce application/json
// @Param body body param.MetricQuery true "metric query request body"
// @Success 200 object response.APIResponse{data=[]response.MetricData}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/metrics/query [POST]
// @Security ApiKeyAuth
func QueryMetrics(c *gin.Context) ([]response.MetricData, error) {
	queryParam := &param.MetricQuery{}
	err := c.Bind(queryParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	metricDatas := metric.QueryMetricData(queryParam)
	return metricDatas, nil
}

// @ID GetTelemetryData
// @Summary get telemetry data
// @Description get telemetry data
// @Tags Metric
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.TelemetryReportResponse}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/metrics/telemetry [GET]
// @Security ApiKeyAuth
func GetTelemetryData(c *gin.Context) (*response.TelemetryReportResponse, error) {
	reportData := response.TelemetryData{}
	telemetryIpKey := fmt.Sprintf("get-telemetry-data:%s", c.RemoteIP())
	shouldFetch := true

	latestFetchTime, ok := store.GetCache().Load(telemetryIpKey)
	if ok {
		if timestamp, ok := latestFetchTime.(int64); ok {
			latestFetchedAt := time.Unix(timestamp, 0)
			if latestFetchedAt.Add(10 * time.Minute).After(time.Now()) {
				shouldFetch = false
			}
		}
	}
	if !shouldFetch {
		return nil, nil
	}

	clusterList := v1alpha1.OBClusterList{}
	err := oceanbase.ClusterClient.List(c, corev1.NamespaceAll, &clusterList, metav1.ListOptions{})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	reportData.Clusters = make([]models.OBCluster, 0, len(clusterList.Items))
	for i := range clusterList.Items {
		modelCluster := telemetry.TransformReportOBCluster(&clusterList.Items[i])
		reportData.Clusters = append(reportData.Clusters, *modelCluster)
	}

	zoneList := v1alpha1.OBZoneList{}
	err = oceanbase.ZoneClient.List(c, corev1.NamespaceAll, &zoneList, metav1.ListOptions{})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	reportData.Zones = make([]models.OBZone, 0, len(zoneList.Items))
	for i := range zoneList.Items {
		modelZone := telemetry.TransformReportOBZone(&zoneList.Items[i])
		reportData.Zones = append(reportData.Zones, *modelZone)
	}

	serverList := v1alpha1.OBServerList{}
	err = oceanbase.ServerClient.List(c, corev1.NamespaceAll, &serverList, metav1.ListOptions{})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	reportData.Servers = make([]models.OBServer, 0, len(serverList.Items))
	for i := range serverList.Items {
		modelServer := telemetry.TransformReportOBServer(&serverList.Items[i])
		reportData.Servers = append(reportData.Servers, *modelServer)
	}

	tenantList := v1alpha1.OBTenantList{}
	err = oceanbase.TenantClient.List(c, corev1.NamespaceAll, &tenantList, metav1.ListOptions{})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	reportData.Tenants = make([]models.OBTenant, 0, len(tenantList.Items))
	for i := range tenantList.Items {
		modelTenant := telemetry.TransformReportOBTenant(&tenantList.Items[i])
		reportData.Tenants = append(reportData.Tenants, *modelTenant)
	}

	backupPolicyList := v1alpha1.OBTenantBackupPolicyList{}
	err = oceanbase.BackupPolicyClient.List(c, corev1.NamespaceAll, &backupPolicyList, metav1.ListOptions{})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	reportData.BackupPolicies = make([]models.OBBackupPolicy, 0, len(backupPolicyList.Items))
	for i := range backupPolicyList.Items {
		modelBackupPolicy := telemetry.TransformReportOBBackupPolicy(&backupPolicyList.Items[i])
		reportData.BackupPolicies = append(reportData.BackupPolicies, *modelBackupPolicy)
	}

	clt := client.GetClient()
	eventList, err := clt.ClientSet.CoreV1().Events(corev1.NamespaceAll).List(c, metav1.ListOptions{FieldSelector: "type=Warning"})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	reportData.WarningEvents = make([]models.K8sEvent, 0, len(eventList.Items))
	for i := range eventList.Items {
		modelEvent := &models.K8sEvent{
			Reason:         eventList.Items[i].Reason,
			Message:        eventList.Items[i].Message,
			Name:           eventList.Items[i].Name,
			Namespace:      eventList.Items[i].Namespace,
			LastTimestamp:  eventList.Items[i].LastTimestamp.Format(time.DateTime),
			FirstTimestamp: eventList.Items[i].FirstTimestamp.Format(time.DateTime),
			Count:          eventList.Items[i].Count,
			Kind:           eventList.Items[i].InvolvedObject.Kind,
			ResourceName:   eventList.Items[i].InvolvedObject.Name,
		}
		reportData.WarningEvents = append(reportData.WarningEvents, *modelEvent)
	}
	reportData.Version = Version

	currentTime := time.Now()
	store.GetCache().Store(telemetryIpKey, currentTime.Unix())

	return &response.TelemetryReportResponse{
		Component: telemetry.TelemetryComponentDashboard,
		Time:      currentTime.Format(time.DateTime),
		Content:   &reportData,
	}, nil
}
