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

	g.PUT("/obtenant/:namespace/:name/unitNumber", handler.ModifyUnitNumber)
	g.PUT("/obtenant/:namespace/:name/:zone/unitConfig", handler.ModifyUnitConfig)
	g.PUT("/obtenant/:namespace/:name/rootPassword", handler.ChangeRootPassword)

	g.POST("/obtenant/:namespace/:name/logreplay", handler.ReplayStandbyLog)
	g.POST("/obtenant/:namespace/:name/version", handler.UpgradeTenantVersion)
	g.POST("/obtenant/:namespace/:name/role", handler.ChangeTenantRole)

	g.GET("/obtenant/:namespace/:name/backupPolicy", handler.GetBackupPolicy)
	g.PUT("/obtenant/:namespace/:name/backupPolicy", handler.CreateBackupPolicy)
	g.POST("/obtenant/:namespace/:name/backupPolicy", handler.UpdateBackupPolicy)
	g.DELETE("/obtenant/:namespace/:name/backupPolicy", handler.DeleteBackupPolicy)
	g.GET("/obtenant/:namespace/:name/:type/backupJobs", handler.ListBackupJobs)
}
