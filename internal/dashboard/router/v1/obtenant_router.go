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

func InitOBTenantRoutes(g *gin.RouterGroup) {
	g.GET("/obtenants", h.Wrap(h.ListAllTenants, acbiz.PathGuard("obtenant", "", "read")))
	g.GET("/obtenants/:namespace/:name", h.Wrap(h.GetTenant, acbiz.PathGuard("obtenant", ":namespace+:name", "read")))
	g.PUT("/obtenants", h.Wrap(h.CreateTenant, acbiz.PathGuard("obtenant", "*", "write")))
	g.DELETE("/obtenants/:namespace/:name", h.Wrap(h.DeleteTenant, acbiz.PathGuard("obtenant", ":namespace+:name", "write")))
	g.PATCH("/obtenants/:namespace/:name", h.Wrap(h.PatchTenant, acbiz.PathGuard("obtenant", ":namespace+:name", "write")))
	g.POST("/obtenants/:namespace/:name/userCredentials", h.Wrap(h.ChangeUserPassword, acbiz.PathGuard("obtenant", ":namespace+:name", "write")))
	g.POST("/obtenants/:namespace/:name/logreplay", h.Wrap(h.ReplayStandbyLog, acbiz.PathGuard("obtenant", ":namespace+:name", "write")))
	g.POST("/obtenants/:namespace/:name/version", h.Wrap(h.UpgradeTenantVersion, acbiz.PathGuard("obtenant", ":namespace+:name", "write")))
	g.POST("/obtenants/:namespace/:name/role", h.Wrap(h.ChangeTenantRole, acbiz.PathGuard("obtenant", ":namespace+:name", "write")))
	g.GET("/obtenants/:namespace/:name/backupPolicy", h.Wrap(h.GetBackupPolicy, acbiz.PathGuard("obtenant", ":namespace+:name", "write")))
	g.PUT("/obtenants/:namespace/:name/backupPolicy", h.Wrap(h.CreateBackupPolicy, acbiz.PathGuard("obtenant", ":namespace+:name", "write")))
	g.PATCH("/obtenants/:namespace/:name/backupPolicy", h.Wrap(h.UpdateBackupPolicy, acbiz.PathGuard("obtenant", ":namespace+:name", "write")))
	g.DELETE("/obtenants/:namespace/:name/backupPolicy", h.Wrap(h.DeleteBackupPolicy, acbiz.PathGuard("obtenant", ":namespace+:name", "write")))
	g.GET("/obtenants/:namespace/:name/backup/:type/jobs", h.Wrap(h.ListBackupJobs, acbiz.PathGuard("obtenant", ":namespace+:name", "read")))
	g.GET("/obtenants/statistic", h.Wrap(h.GetOBTenantStatistic, acbiz.PathGuard("obtenant", "*", "read")))
	g.PUT("/obtenants/:namespace/:name/pools/:zoneName", h.Wrap(h.CreateOBTenantPool, acbiz.PathGuard("obtenant", ":namespace+:name", "write")))
	g.DELETE("/obtenants/:namespace/:name/pools/:zoneName", h.Wrap(h.DeleteOBTenantPool, acbiz.PathGuard("obtenant", ":namespace+:name", "write")))
	g.PATCH("/obtenants/:namespace/:name/pools/:zoneName", h.Wrap(h.PatchOBTenantPool, acbiz.PathGuard("obtenant", ":namespace+:name", "write")))
	g.GET("/obtenants/:namespace/:name/related-events", h.Wrap(h.ListOBTenantRelatedEvents, acbiz.PathGuard("obtenant", ":namespace+:name", "read")))
}
