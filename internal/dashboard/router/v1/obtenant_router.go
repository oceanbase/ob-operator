package v1

import (
	"github.com/gin-gonic/gin"

	h "github.com/oceanbase/ob-operator/internal/dashboard/handler"
)

func InitOBTenantRoutes(g *gin.RouterGroup) {
	g.GET("/obtenants", h.Wrap(h.ListAllTenants))
	g.GET("/obtenants/:namespace/:name", h.Wrap(h.GetTenant))
	g.PUT("/obtenants", h.Wrap(h.CreateTenant))
	g.DELETE("/obtenants/:namespace/:name", h.Wrap(h.DeleteTenant))
	g.PATCH("/obtenants/:namespace/:name", h.Wrap(h.PatchTenant))
	g.POST("/obtenants/:namespace/:name/userCredentials", h.Wrap(h.ChangeUserPassword))
	g.POST("/obtenants/:namespace/:name/logreplay", h.Wrap(h.ReplayStandbyLog))
	g.POST("/obtenants/:namespace/:name/version", h.Wrap(h.UpgradeTenantVersion))
	g.POST("/obtenants/:namespace/:name/role", h.Wrap(h.ChangeTenantRole))
	g.GET("/obtenants/:namespace/:name/backupPolicy", h.Wrap(h.GetBackupPolicy))
	g.PUT("/obtenants/:namespace/:name/backupPolicy", h.Wrap(h.CreateBackupPolicy))
	g.PATCH("/obtenants/:namespace/:name/backupPolicy", h.Wrap(h.UpdateBackupPolicy))
	g.DELETE("/obtenants/:namespace/:name/backupPolicy", h.Wrap(h.DeleteBackupPolicy))
	g.GET("/obtenants/:namespace/:name/backup/:type/jobs", h.Wrap(h.ListBackupJobs))
}
