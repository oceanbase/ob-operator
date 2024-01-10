package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/internal/handler"
)

func InitOBTenantRoutes(g *gin.RouterGroup) {
	g.GET("/obtenants", handler.ListAllTenants)
	g.GET("/obtenant/:namespace/:name", handler.GetTenant)
	g.PUT("/obtenant/:namespace/:name", handler.CreateTenant)
	g.DELETE("/obtenant/:namespace/:name", handler.DeleteTenant)
}
