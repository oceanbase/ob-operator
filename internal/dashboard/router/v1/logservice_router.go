/*
Copyright (c) 2026 OceanBase
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

	"github.com/oceanbase/ob-operator/internal/dashboard/business/ac"
	h "github.com/oceanbase/ob-operator/internal/dashboard/handler"
)

func InitLogServiceRoutes(g *gin.RouterGroup) {
	g.GET("/logservices/:namespace/:name/metrics", h.Wrap(h.QueryLogServiceMetrics, ac.PathGuard("oblogservice", ":namespace+:name", "read")))
	g.GET("/obclusters/:namespace/:name/ss-metrics", h.Wrap(h.QuerySSMetrics, ac.PathGuard("obcluster", ":namespace+:name", "read")))
	g.GET("/logservices", h.Wrap(h.ListLogServices, ac.PathGuard("oblogservice", "*", "read")))
	g.POST("/logservices", h.Wrap(h.CreateLogService, ac.PathGuard("oblogservice", "*", "write")))
	g.GET("/logservices/:namespace/:name", h.Wrap(h.GetLogService, ac.PathGuard("oblogservice", ":namespace+:name", "read")))
	g.PATCH("/logservices/:namespace/:name/replicas", h.Wrap(h.ScaleLogService, ac.PathGuard("oblogservice", ":namespace+:name", "write")))
	g.DELETE("/logservices/:namespace/:name", h.Wrap(h.DeleteLogService, ac.PathGuard("oblogservice", ":namespace+:name", "write")))
}
