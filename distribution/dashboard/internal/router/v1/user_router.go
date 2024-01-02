package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/internal/handler"
)

func InitUserRoutes(g *gin.RouterGroup) {
	g.POST("/login", handler.Login)
	g.POST("/logout", handler.Logout)
}
