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

package v1

import (
	"github.com/gin-gonic/gin"

	acbiz "github.com/oceanbase/ob-operator/internal/dashboard/business/ac"
	h "github.com/oceanbase/ob-operator/internal/dashboard/handler"
)

func InitOBClusterRoutes(g *gin.RouterGroup) {
	g.GET("/obclusters/statistic", h.Wrap(h.GetOBClusterStatistic, acbiz.PathGuard("obcluster", "*", "read")))
	g.GET("/obclusters", h.Wrap(h.ListOBClusters, acbiz.PathGuard("obcluster", "", "read")))
	g.POST("/obclusters", h.Wrap(h.CreateOBCluster, acbiz.PathGuard("obcluster", "*", "read")))
	g.GET("/obclusters/namespace/:namespace/name/:name", h.Wrap(h.GetOBCluster, acbiz.PathGuard("obcluster", ":namespace+:name", "read")))
	g.POST("/obclusters/namespace/:namespace/name/:name", h.Wrap(h.UpgradeOBCluster, acbiz.PathGuard("obcluster", ":namespace+:name", "write")))
	g.DELETE("/obclusters/namespace/:namespace/name/:name", h.Wrap(h.DeleteOBCluster, acbiz.PathGuard("obcluster", ":namespace+:name", "write")))
	g.POST("/obclusters/namespace/:namespace/name/:name/obzones", h.Wrap(h.AddOBZone, acbiz.PathGuard("obcluster", ":namespace+:name", "write")))
	g.POST("/obclusters/namespace/:namespace/name/:name/obzones/:obzoneName/scale", h.Wrap(h.ScaleOBServer, acbiz.PathGuard("obcluster", ":namespace+:name", "write")))
	g.DELETE("/obclusters/namespace/:namespace/name/:name/obzones/:obzoneName", h.Wrap(h.DeleteOBZone, acbiz.PathGuard("obcluster", ":namespace+:name", "write")))
	g.GET("/obclusters/:namespace/:name/resource-usages", h.Wrap(h.ListOBClusterResources, acbiz.PathGuard("obcluster", ":namespace+:name", "read")))
	g.GET("/obclusters/:namespace/:name/related-events", h.Wrap(h.ListOBClusterRelatedEvents, acbiz.PathGuard("obcluster", ":namespace+:name", "read")))
}
