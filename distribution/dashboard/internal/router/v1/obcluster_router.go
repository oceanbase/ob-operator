package v1

import (
	"github.com/gin-gonic/gin"
	h "github.com/oceanbase/oceanbase-dashboard/internal/handler"
)

func InitOBClusterRoutes(g *gin.RouterGroup) {
	g.GET("/obclusters/statistic", h.W(h.GetOBClusterStatistic))
	g.GET("/obclusters", h.W(h.ListOBClusters))
	g.POST("/obclusters", h.W(h.CreateOBCluster))
	g.GET("/obclusters/namespace/:namespace/name/:name", h.W(h.GetOBCluster))
	g.POST("/obclusters/namespace/:namespace/name/:name", h.W(h.UpgradeOBCluster))
	g.DELETE("/obclusters/namespace/:namespace/name/:name", h.W(h.DeleteOBCluster))
	g.POST("/obclusters/namespace/:namespace/name/:name/obzones", h.W(h.AddOBZone))
	g.POST("/obclusters/namespace/:namespace/name/:name/obzones/:obzoneName/scale", h.W(h.ScaleOBServer))
	g.DELETE("/obclusters/namespace/:namespace/name/:name/obzones/:obzoneName", h.W(h.DeleteOBZone))
}
