package v1

import (
	"github.com/gin-gonic/gin"
	h "github.com/oceanbase/oceanbase-dashboard/internal/handler"
)

func InitOBTenantRoutes(g *gin.RouterGroup) {
	g.GET("/obtenants", h.W(h.ListAllTenants))
	g.GET("/obtenant/:namespace/:name", h.W(h.GetTenant))
	g.PUT("/obtenant/:namespace/:name", h.W(h.CreateTenant))
	g.DELETE("/obtenant/:namespace/:name", h.W(h.DeleteTenant))
	g.PATCH("/obtenant/:namespace/:name", h.W(h.PatchTenant))
	g.POST("/obtenant/:namespace/:name/userCredentials", h.W(h.ChangeUserPassword))
	g.POST("/obtenant/:namespace/:name/logreplay", h.W(h.ReplayStandbyLog))
	g.POST("/obtenant/:namespace/:name/version", h.W(h.UpgradeTenantVersion))
	g.POST("/obtenant/:namespace/:name/role", h.W(h.ChangeTenantRole))
	g.GET("/obtenant/:namespace/:name/backupPolicy", h.W(h.GetBackupPolicy))
	g.PUT("/obtenant/:namespace/:name/backupPolicy", h.W(h.CreateBackupPolicy))
	g.POST("/obtenant/:namespace/:name/backupPolicy", h.W(h.UpdateBackupPolicy))
	g.DELETE("/obtenant/:namespace/:name/backupPolicy", h.W(h.DeleteBackupPolicy))
	g.GET("/obtenant/:namespace/:name/backup/:type/jobs", h.W(h.ListBackupJobs))
}
