package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/internal/handler"
)

func InitOBClusterRoutes(g *gin.RouterGroup) {
	g.GET("/obclusters/statistic", handler.GetOBClusterStatistic)
	g.GET("/obclusters", handler.ListOBClusters)
	g.POST("/obclusters", handler.CreateOBCluster)
	g.GET("/obclusters/namespace/:namespace/name/:name", handler.GetOBCluster)
	g.POST("/obclusters/namespace/:namespace/name/:name", handler.UpgradeOBCluster)
	g.DELETE("/obclusters/namespace/:namespace/name/:name", handler.DeleteOBCluster)
	g.POST("/obclusters/namespace/:namespace/name/:name/obzones", handler.AddOBZone)
	g.POST("/obclusters/namespace/:namespace/name/:name/obzones/:obzoneName/scale", handler.ScaleOBServer)
	g.DELETE("/obclusters/namespace/:namespace/name/:name/obzones/:obzoneName", handler.DeleteOBZone)
}
