package v1

import (
	"github.com/gin-gonic/gin"
	h "github.com/oceanbase/ob-operator/internal/dashboard/handler"
)

func InitInfoRoutes(g *gin.RouterGroup) {
	g.GET("/info", h.Wrap(h.GetProcessInfo))
}
