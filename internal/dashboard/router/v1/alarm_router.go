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

	h "github.com/oceanbase/ob-operator/internal/dashboard/handler"
)

func InitAlarmRoutes(g *gin.RouterGroup) {
	// alert
	g.POST("/alarm/alert/alerts", h.Wrap(h.ListAlerts))

	// silence
	g.POST("/alarm/silence/silencers", h.Wrap(h.ListSilencers))
	g.GET("/alarm/silence/silencers/:id", h.Wrap(h.GetSilencer))
	g.PUT("/alarm/silence/silencers", h.Wrap(h.CreateOrUpdateSilencer))
	g.DELETE("/alarm/silence/silencers/:id", h.Wrap(h.DeleteSilencer))

	// rule
	g.POST("/alarm/rule/rules", h.Wrap(h.ListRules))
	g.GET("/alarm/rule/rules/:name", h.Wrap(h.GetRule))
	g.PUT("/alarm/rule/rules", h.Wrap(h.CreateOrUpdateRule))
	g.DELETE("/alarm/rule/rules/:name", h.Wrap(h.DeleteRule))

	// receiver
	g.POST("/alarm/receiver/receivers", h.Wrap(h.ListReceivers))
	g.GET("/alarm/receiver/receivers/:name", h.Wrap(h.GetReceiver))
	g.PUT("/alarm/receiver/receivers", h.Wrap(h.CreateOrUpdateReceiver))
	g.DELETE("/alarm/receiver/receivers/:name", h.Wrap(h.DeleteReceiver))
	g.POST("/alarm/receiver/templates", h.Wrap(h.ListReceiverTemplates))
	g.GET("/alarm/receiver/templates/:type", h.Wrap(h.GetReceiverTemplate))

	// route
	g.POST("/alarm/route/routes", h.Wrap(h.ListRoutes))
	g.GET("/alarm/route/routes/:id", h.Wrap(h.GetRoute))
	g.PUT("/alarm/route/routes", h.Wrap(h.CreateOrUpdateRoute))
	g.DELETE("/alarm/route/routes/:id", h.Wrap(h.DeleteRoute))
}
