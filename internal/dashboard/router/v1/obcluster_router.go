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

	h "github.com/oceanbase/ob-operator/internal/dashboard/handler"
)

func InitOBClusterRoutes(g *gin.RouterGroup) {
	g.GET("/obclusters/statistic", h.Wrap(h.GetOBClusterStatistic))
	g.GET("/obclusters", h.Wrap(h.ListOBClusters))
	g.POST("/obclusters", h.Wrap(h.CreateOBCluster))
	g.GET("/obclusters/namespace/:namespace/name/:name", h.Wrap(h.GetOBCluster))
	g.POST("/obclusters/namespace/:namespace/name/:name", h.Wrap(h.UpgradeOBCluster))
	g.DELETE("/obclusters/namespace/:namespace/name/:name", h.Wrap(h.DeleteOBCluster))
	g.POST("/obclusters/namespace/:namespace/name/:name/obzones", h.Wrap(h.AddOBZone))
	g.POST("/obclusters/namespace/:namespace/name/:name/obzones/:obzoneName/scale", h.Wrap(h.ScaleOBServer))
	g.DELETE("/obclusters/namespace/:namespace/name/:name/obzones/:obzoneName", h.Wrap(h.DeleteOBZone))
	g.GET("/obclusters/:namespace/:name/essential-parameters", h.Wrap(h.ListOBClusterResources))
	g.GET("/obclusters/:namespace/:name/resource-usages", h.Wrap(h.ListOBClusterResources))
}
