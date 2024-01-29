package v1

import (
	"github.com/gin-gonic/gin"
	h "github.com/oceanbase/oceanbase-dashboard/internal/handler"
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
}
