package v1

import (
	"github.com/gin-gonic/gin"
	h "github.com/oceanbase/oceanbase-dashboard/internal/handler"
)

func InitInfoRoutes(g *gin.RouterGroup) {
	g.GET("/info", h.Wrap(h.GetProcessInfo))
}
