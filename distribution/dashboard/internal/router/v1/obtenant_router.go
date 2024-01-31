package v1

import (
	"github.com/gin-gonic/gin"
	h "github.com/oceanbase/oceanbase-dashboard/internal/handler"
)

func InitOBTenantRoutes(g *gin.RouterGroup) {
	g.GET("/obtenants", h.Wrap(h.ListAllTenants))
	g.GET("/obtenant/:namespace/:name", h.Wrap(h.GetTenant))
	g.PUT("/obtenant/:namespace/:name", h.Wrap(h.CreateTenant))
	g.DELETE("/obtenant/:namespace/:name", h.Wrap(h.DeleteTenant))
	g.PATCH("/obtenant/:namespace/:name", h.Wrap(h.PatchTenant))
	g.POST("/obtenant/:namespace/:name/userCredentials", h.Wrap(h.ChangeUserPassword))
	g.POST("/obtenant/:namespace/:name/logreplay", h.Wrap(h.ReplayStandbyLog))
	g.POST("/obtenant/:namespace/:name/version", h.Wrap(h.UpgradeTenantVersion))
	g.POST("/obtenant/:namespace/:name/role", h.Wrap(h.ChangeTenantRole))
	g.GET("/obtenant/:namespace/:name/backupPolicy", h.Wrap(h.GetBackupPolicy))
	g.PUT("/obtenant/:namespace/:name/backupPolicy", h.Wrap(h.CreateBackupPolicy))
	g.POST("/obtenant/:namespace/:name/backupPolicy", h.Wrap(h.UpdateBackupPolicy))
	g.DELETE("/obtenant/:namespace/:name/backupPolicy", h.Wrap(h.DeleteBackupPolicy))
	g.GET("/obtenant/:namespace/:name/backup/:type/jobs", h.Wrap(h.ListBackupJobs))
}
