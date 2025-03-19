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

package router

import (
	"os"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/requestid"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	docs "github.com/oceanbase/ob-operator/internal/dashboard/generated/swagger"
	"github.com/oceanbase/ob-operator/internal/dashboard/middleware"
	v1 "github.com/oceanbase/ob-operator/internal/dashboard/router/v1"
	"github.com/oceanbase/ob-operator/internal/dashboard/server/constant"
)

func InitRoutes(router *gin.Engine) {
	sessionSecret := "secret"
	if secretEnv := os.Getenv("SESSION_SECRET"); secretEnv != "" {
		sessionSecret = secretEnv
	}
	store := cookie.NewStore([]byte(sessionSecret))
	store.Options(sessions.Options{
		HttpOnly: true,
		MaxAge:   constant.DefaultSessionExpiration,
		Path:     "/",
	})
	// use gin's crash free middleware
	router.Use(
		gin.Recovery(),
		requestid.New(),
		middleware.Logging(),
		gzip.Gzip(gzip.DefaultCompression),
		sessions.Sessions("cookies", store),
	)

	// host web dist
	router.Use(static.Serve("/", static.LocalFile("ui/dist", false)))
	router.NoRoute(func(c *gin.Context) {
		c.Redirect(302, "/")
	})

	// host docs
	if os.Getenv("ENABLE_SWAGGER_DOC") == "true" {
		docs.SwaggerInfo.BasePath = "/"
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
		router.Use(static.Serve("/api-gen", static.LocalFile("internal/dashboard/generated/swagger", false)))
	}

	v1Group := router.Group("/api/v1")
	// login api does not require login
	if os.Getenv("DEBUG_DASHBOARD") != "true" {
		v1Group = router.Group("/api/v1",
			middleware.LoginRequired(),
			middleware.RefreshExpiration(),
		)
	}

	// init all routes under /api/v1
	v1.InitWebhookRoutes(v1Group)
	v1.InitInfoRoutes(v1Group)
	v1.InitClusterRoutes(v1Group)
	v1.InitK8sClusterRoutes(v1Group)
	v1.InitMetricRoutes(v1Group)
	v1.InitOBClusterRoutes(v1Group)
	v1.InitUserRoutes(v1Group)
	v1.InitOBTenantRoutes(v1Group)
	v1.InitTerminalRoutes(v1Group)
	v1.InitAlarmRoutes(v1Group)
	v1.InitOBProxyRoutes(v1Group)
	v1.InitAccessControlRoutes(v1Group)
	v1.InitInspectionRoutes(v1Group)
	v1.InitSqlRoutes(v1Group)
	v1.InitJobRoutes(v1Group)
	v1.InitStorageRoutes(v1Group)
}
