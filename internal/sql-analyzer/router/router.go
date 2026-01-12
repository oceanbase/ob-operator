/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/handler"
)

func Register(r *gin.Engine) {
	apiV1 := r.Group("/api/v1")
	{
		// Debug route
		apiV1.POST("/debug/query", handler.Wrap(handler.DebugQuery))

		tenants := apiV1.Group("/tenants/:tenant_name")
		{
			// The openapi definition is in the handler
			tenants.POST("/sql-stats", handler.Wrap(handler.QuerySqlStats))
			tenants.POST("/request-stats", handler.Wrap(handler.GetRequestStatistics))
			tenants.POST("/sql-detail", handler.Wrap(handler.GetSqlDetailInfo))
			tenants.POST("/sql-history", handler.Wrap(handler.GetSqlHistoryInfo))
			// Add plan detail router
			tenants.POST("/plan_detail", handler.Wrap(handler.GetPlanDetail))
		}
	}
}
