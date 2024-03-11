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

func InitMetricRoutes(g *gin.RouterGroup) {
	g.GET("/metrics", h.Wrap(h.ListMetricMetas))
	g.POST("/metrics/query", h.Wrap(h.QueryMetrics))
	g.GET("/metrics/telemetry", h.Wrap(h.GetTelemetryData))
}
