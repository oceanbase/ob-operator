package v1

import (
	"github.com/gin-gonic/gin"
	h "github.com/oceanbase/oceanbase-dashboard/internal/handler"
)

func InitK8sRoutes(g *gin.RouterGroup) {
	g.GET("/cluster/events", h.W(h.ListK8sEvents))
	g.GET("/cluster/nodes", h.W(h.ListK8sNodes))
	g.GET("/cluster/namespaces", h.W(h.ListK8sNamespaces))
	g.GET("/cluster/storageClasses", h.W(h.ListK8sStorageClasses))
	g.POST("/cluster/namespaces", h.W(h.CreateK8sNamespace))
}
