package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	ls "github.com/oceanbase/ob-operator/internal/dashboard/business/logservice"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/ssmonitor"
	oberr "github.com/oceanbase/ob-operator/pkg/errors"
	"net"
	"strconv"
	"time"
)

// Prometheus is a sidecar and calls the loopback interface. Do not trust
// X-Forwarded-For, bypass login on external requests, or expose SQL credentials.
func CollectSSMetrics(c *gin.Context) {
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil || !net.ParseIP(host).IsLoopback() {
		c.AbortWithStatus(403)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 24*time.Second)
	defer cancel()
	text, err := ssmonitor.Collect(ctx, ls.Default())
	if err != nil {
		c.String(503, "SS collection failed\n")
		return
	}
	c.Data(200, "text/plain; version=0.0.4", []byte(text))
}
func ssMetricQuery(c *gin.Context, kind string) ([]ssmonitor.Series, error) {
	minutes := 60
	if q := c.Query("minutes"); q != "" {
		var err error
		minutes, err = strconv.Atoi(q)
		if err != nil {
			return nil, oberr.NewBadRequest("Invalid time range")
		}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()
	result, err := ssmonitor.Query(ctx, "http://127.0.0.1:9090", kind, c.Param("namespace"), c.Param("name"), minutes)
	if err != nil {
		return nil, oberr.NewBadRequest(err.Error())
	}
	return result, nil
}
func QueryLogServiceMetrics(c *gin.Context) ([]ssmonitor.Series, error) {
	return ssMetricQuery(c, "logservice")
}
func QuerySSMetrics(c *gin.Context) ([]ssmonitor.Series, error) { return ssMetricQuery(c, "obcluster") }
