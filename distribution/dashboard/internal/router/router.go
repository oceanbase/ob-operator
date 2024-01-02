package router

import (
	"os"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/requestid"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/docs"
	"github.com/oceanbase/oceanbase-dashboard/internal/middleware"
	v1 "github.com/oceanbase/oceanbase-dashboard/internal/router/v1"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitRoutes(router *gin.Engine) {
	sessionSecret := "secret"
	if secretEnv := os.Getenv("SESSION_SECRET"); secretEnv != "" {
		sessionSecret = secretEnv
	}
	store := cookie.NewStore([]byte(sessionSecret))
	store.Options(sessions.Options{
		HttpOnly: true,
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
	}

	// login api does not require login
	v1Group := router.Group("/api/v1",
		middleware.LoginRequired(),
		middleware.RefreshExpiration(),
	)

	// init all routes under /api/v1
	v1.InitInfoRoutes(v1Group)
	v1.InitK8sRoutes(v1Group)
	v1.InitMetricRoutes(v1Group)
	v1.InitOBClusterRoutes(v1Group)
	v1.InitOBZoneRoutes(v1Group)
	v1.InitUserRoutes(v1Group)
	v1.InitOBTenantRoutes(v1Group)
}
