package v1

import (
	"github.com/gin-gonic/gin"
	h "github.com/oceanbase/oceanbase-dashboard/internal/handler"
)

func InitMetricRoutes(g *gin.RouterGroup) {
	g.GET("/metrics", h.W(h.ListMetricMetas))
	g.POST("/metrics/query", h.W(h.QueryMetrics))
}
