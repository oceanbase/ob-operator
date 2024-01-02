package v1

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitDocRoutes(g *gin.RouterGroup) {
	g.GET("/docs", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
