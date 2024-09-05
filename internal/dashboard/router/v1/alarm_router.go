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

var diagnoseGuard = acbiz.AND(
	acbiz.PathGuard(string(acbiz.DomainOBCluster), "*", string(acbiz.ActionRead)),
	acbiz.PathGuard(string(acbiz.DomainSystem), "*", string(acbiz.ActionWrite)),
)

func InitAlarmRoutes(g *gin.RouterGroup) {
	// alert
	g.POST("/alarm/alert/alerts", h.Wrap(h.ListAlerts, acbiz.PathGuard("alarm", "*", "read")))
	g.POST("/alarm/alert/diagnose", h.Wrap(h.DiagnoseAlert, diagnoseGuard))

	// silence
	g.POST("/alarm/silence/silencers", h.Wrap(h.ListSilencers, acbiz.PathGuard("alarm", "*", "read")))
	g.GET("/alarm/silence/silencers/:id", h.Wrap(h.GetSilencer, acbiz.PathGuard("alarm", "*", "read")))
	g.PUT("/alarm/silence/silencers", h.Wrap(h.CreateOrUpdateSilencer, acbiz.PathGuard("alarm", "*", "write")))
	g.DELETE("/alarm/silence/silencers/:id", h.Wrap(h.DeleteSilencer, acbiz.PathGuard("alarm", "*", "write")))

	// rule
	g.POST("/alarm/rule/rules", h.Wrap(h.ListRules, acbiz.PathGuard("alarm", "*", "read")))
	g.GET("/alarm/rule/rules/:name", h.Wrap(h.GetRule, acbiz.PathGuard("alarm", "*", "read")))
	g.PUT("/alarm/rule/rules", h.Wrap(h.CreateOrUpdateRule, acbiz.PathGuard("alarm", "*", "write")))
	g.DELETE("/alarm/rule/rules/:name", h.Wrap(h.DeleteRule, acbiz.PathGuard("alarm", "*", "write")))

	// receiver
	g.POST("/alarm/receiver/receivers", h.Wrap(h.ListReceivers, acbiz.PathGuard("alarm", "*", "read")))
	g.GET("/alarm/receiver/receivers/:name", h.Wrap(h.GetReceiver, acbiz.PathGuard("alarm", "*", "read")))
	g.PUT("/alarm/receiver/receivers", h.Wrap(h.CreateOrUpdateReceiver, acbiz.PathGuard("alarm", "*", "write")))
	g.DELETE("/alarm/receiver/receivers/:name", h.Wrap(h.DeleteReceiver, acbiz.PathGuard("alarm", "*", "write")))
	g.POST("/alarm/receiver/templates", h.Wrap(h.ListReceiverTemplates, acbiz.PathGuard("alarm", "*", "read")))
	g.GET("/alarm/receiver/templates/:type", h.Wrap(h.GetReceiverTemplate, acbiz.PathGuard("alarm", "*", "write")))

	// route
	g.POST("/alarm/route/routes", h.Wrap(h.ListRoutes, acbiz.PathGuard("alarm", "*", "read")))
	g.GET("/alarm/route/routes/:id", h.Wrap(h.GetRoute, acbiz.PathGuard("alarm", "*", "read")))
	g.PUT("/alarm/route/routes", h.Wrap(h.CreateOrUpdateRoute, acbiz.PathGuard("alarm", "*", "write")))
	g.DELETE("/alarm/route/routes/:id", h.Wrap(h.DeleteRoute, acbiz.PathGuard("alarm", "*", "write")))
}
