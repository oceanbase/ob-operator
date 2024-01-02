package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/internal/handler"
)

func InitInfoRoutes(g *gin.RouterGroup) {
	g.GET("/info", handler.GetProcessInfo)
}
