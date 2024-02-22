package v1

import (
	"github.com/gin-gonic/gin"
	h "github.com/oceanbase/ob-operator/internal/dashboard/handler"
)

func InitMetricRoutes(g *gin.RouterGroup) {
	g.GET("/metrics", h.Wrap(h.ListMetricMetas))
	g.POST("/metrics/query", h.Wrap(h.QueryMetrics))
}
