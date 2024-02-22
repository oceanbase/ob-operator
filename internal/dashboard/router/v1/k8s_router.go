package v1

import (
	"github.com/gin-gonic/gin"
	h "github.com/oceanbase/ob-operator/internal/dashboard/handler"
)

func InitK8sRoutes(g *gin.RouterGroup) {
	g.GET("/cluster/events", h.Wrap(h.ListK8sEvents))
	g.GET("/cluster/nodes", h.Wrap(h.ListK8sNodes))
	g.GET("/cluster/namespaces", h.Wrap(h.ListK8sNamespaces))
	g.GET("/cluster/storageClasses", h.Wrap(h.ListK8sStorageClasses))
	g.POST("/cluster/namespaces", h.Wrap(h.CreateK8sNamespace))
}
