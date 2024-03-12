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
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	"github.com/oceanbase/ob-operator/internal/oceanbase"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	"github.com/oceanbase/ob-operator/internal/telemetry/models"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

var (
	Version    = ""
	CommitHash = ""
	BuildTime  = ""
)

// @ID GetProcessInfo
// @Summary Get process info
// @Description Get process info of OceanBase Dashboard, including process name etc.
// @Tags Info
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.DashboardInfo}
// @Failure 500 object response.APIResponse
// @Router /api/v1/info [GET]
func GetProcessInfo(_ *gin.Context) (*response.DashboardInfo, error) {
	pubBytes, err := crypto.PublicKeyToBytes()
	if err != nil {
		return nil, err
	}
	return &response.DashboardInfo{
		AppName:          "oceanbase-dashboard",
		Version:          strings.Join([]string{Version, CommitHash, BuildTime}, "-"),
		PublicKey:        string(pubBytes),
		ReportStatistics: os.Getenv("DISABLE_REPORT_STATISTICS") != "true",
	}, nil
}

// @ID GetStatistics
// @Summary get telemetry data
// @Description get telemetry data
// @Tags Info
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.StatisticDataResponse}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/metrics/telemetry [GET]
// @Security ApiKeyAuth
func GetStatistics(c *gin.Context) (*response.StatisticDataResponse, error) {
	reportData := response.StatisticData{}
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

	return &response.StatisticDataResponse{
		Component: telemetry.TelemetryComponentDashboard,
		Time:      currentTime.Format(time.DateTime),
		Content:   &reportData,
	}, nil
}

// @ID GetStatistics
// @Summary get telemetry data
// @Description get telemetry data
// @Tags Info
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.StatisticDataResponse}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/metrics/telemetry [GET]
// @Security ApiKeyAuth
func GetStatistics(c *gin.Context) (*response.StatisticDataResponse, error) {
	reportData := response.StatisticData{}
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

	return &response.StatisticDataResponse{
		Component: telemetry.TelemetryComponentDashboard,
		Time:      currentTime.Format(time.DateTime),
		Content:   &reportData,
	}, nil
}
