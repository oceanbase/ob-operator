package v1

import (
	"github.com/gin-gonic/gin"
	ac "github.com/oceanbase/ob-operator/internal/dashboard/business/ac"
	h "github.com/oceanbase/ob-operator/internal/dashboard/handler"
)

func InitLogServiceRoutes(g *gin.RouterGroup) {
	g.GET("/logservices/:namespace/:name/metrics", h.Wrap(h.QueryLogServiceMetrics, ac.PathGuard("oblogservice", ":namespace+:name", "read")))
	g.GET("/obclusters/:namespace/:name/ss-metrics", h.Wrap(h.QuerySSMetrics, ac.PathGuard("obcluster", ":namespace+:name", "read")))
	g.GET("/logservices", h.Wrap(h.ListLogServices, ac.PathGuard("oblogservice", "*", "read")))
	g.POST("/logservices", h.Wrap(h.CreateLogService, ac.PathGuard("oblogservice", "*", "write")))
	g.GET("/logservices/:namespace/:name", h.Wrap(h.GetLogService, ac.PathGuard("oblogservice", ":namespace+:name", "read")))
	g.PATCH("/logservices/:namespace/:name/replicas", h.Wrap(h.ScaleLogService, ac.PathGuard("oblogservice", ":namespace+:name", "write")))
	g.DELETE("/logservices/:namespace/:name", h.Wrap(h.DeleteLogService, ac.PathGuard("oblogservice", ":namespace+:name", "write")))
}
