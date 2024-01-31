package v1

import (
	"github.com/gin-gonic/gin"
	h "github.com/oceanbase/oceanbase-dashboard/internal/handler"
)

func InitUserRoutes(g *gin.RouterGroup) {
	g.POST("/login", h.Wrap(h.Login))
	g.POST("/logout", h.Wrap(h.Logout))
}
