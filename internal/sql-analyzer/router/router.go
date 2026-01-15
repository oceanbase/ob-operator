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
	"os"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	docs "github.com/oceanbase/ob-operator/internal/sql-analyzer/generated/swagger"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/handler"
)

func Register(r *gin.Engine) {
	// host docs
	if os.Getenv("ENABLE_SWAGGER_DOC") == "true" {
		docs.SwaggerInfo.BasePath = "/"
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
		r.Use(static.Serve("/api-gen", static.LocalFile("internal/sql-analyzer/generated/swagger", false)))
	}

	apiV1 := r.Group("/api/v1")
	{
		// Debug route
		apiV1.POST("/debug/query", handler.Wrap(handler.DebugQuery))

		tenants := apiV1.Group("/tenants/:tenant_name")
		{
			tenants.POST("/sql-stats", handler.Wrap(handler.QuerySqlStats))
			tenants.POST("/request-stats", handler.Wrap(handler.GetRequestStatistics))
			tenants.POST("/sql-detail", handler.Wrap(handler.GetSqlDetailInfo))
			tenants.POST("/sql-history", handler.Wrap(handler.GetSqlHistoryInfo))
			tenants.POST("/plan_detail", handler.Wrap(handler.GetPlanDetail))
		}
	}
}
