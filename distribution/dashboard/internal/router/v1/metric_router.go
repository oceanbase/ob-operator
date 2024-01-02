package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/internal/handler"
)

func InitMetricRoutes(g *gin.RouterGroup) {
	g.GET("/metrics", handler.ListMetricMetas)
	g.POST("/metrics/query", handler.QueryMetrics)
}
