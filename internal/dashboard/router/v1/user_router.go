package v1

import (
	"github.com/gin-gonic/gin"

	h "github.com/oceanbase/ob-operator/internal/dashboard/handler"
)

func InitUserRoutes(g *gin.RouterGroup) {
	g.POST("/login", h.Wrap(h.Login))
	g.POST("/logout", h.Wrap(h.Logout))
}
