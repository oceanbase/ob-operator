package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/internal/handler"
)

func InitK8sRoutes(g *gin.RouterGroup) {
	g.GET("/cluster/events", handler.ListK8sEvents)
	g.GET("/cluster/nodes", handler.ListK8sNodes)
	g.GET("/cluster/namespaces", handler.ListK8sNamespaces)
	g.GET("/cluster/storageClasses", handler.ListK8sStorageClasses)
	g.POST("/cluster/namespaces", handler.CreateK8sNamespace)
}
