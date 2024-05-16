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
	logger "github.com/sirupsen/logrus"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
)

func loggingCreateOBClusterParam(param *param.CreateOBClusterParam) {
	logger.
		WithField("Name", param.Name).
		WithField("Namespace", param.Namespace).
		WithField("ClusterName", param.ClusterName).
		WithField("ClusterId", param.ClusterId).
		WithField("Mode", param.Mode).
		WithField("Topology", param.Topology).
		WithField("OBServer", param.OBServer).
		WithField("Monitor", param.Monitor).
		WithField("Parameters", param.Parameters).
		Infof("Create OBCluster param")
}

func loggingCreateOBTenantParam(param *param.CreateOBTenantParam) {
	logger.
		WithField("Name", param.Name).
		WithField("Namespace", param.Namespace).
		WithField("ClusterName", param.ClusterName).
		WithField("TenantName", param.TenantName).
		WithField("UnitNumber", param.UnitNumber).
		WithField("ConnectWhiteList", param.ConnectWhiteList).
		WithField("Charset", param.Charset).
		WithField("UnitConfig", param.UnitConfig).
		WithField("Pools", param.Pools).
		WithField("TenantRole", param.TenantRole).
		WithField("Source", param.Source).
		Infof("Create OBTenant param")
}
