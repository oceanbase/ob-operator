package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/internal/business/metric"
	metricconst "github.com/oceanbase/oceanbase-dashboard/internal/business/metric/constant"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/param"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/response"
	buzerr "github.com/oceanbase/oceanbase-dashboard/pkg/errors"
)

// @ID ListAllMetrics
// @Summary list all metrics
// @Description list all metrics meta info, return by groups
// @Tags Metric
// @Accept application/json
// @Produce application/json
// @Param scope query string true "metrics scope" Enums(OBCLUSTER, OBTENANT)
// @Success 200 object response.APIResponse{data=[]response.MetricClass}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/metrics [GET]
// @Security ApiKeyAuth
func ListMetricMetas(c *gin.Context) ([]response.MetricClass, error) {
	// return mock data
	language := c.GetHeader("Accept-Language")
	scope := c.Query("scope")
	if scope != metricconst.ScopeCluster && scope != metricconst.ScopeTenant && scope != metricconst.ScopeClusterOverview {
		err := errors.New("invalid scope")
		return nil, buzerr.NewBadRequest(err.Error())
	}
	metricClasses, err := metric.ListMetricClasses(scope, language)
	if err != nil {
		return nil, err
	}
	return metricClasses, nil
}

// @ID QueryMetrics
// @Summary query metrics
// @Description query metric data
// @Tags Metric
// @Accept application/json
// @Produce application/json
// @Param body body param.MetricQuery true "metric query request body"
// @Success 200 object response.APIResponse{data=[]response.MetricData}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/metrics/query [POST]
// @Security ApiKeyAuth
func QueryMetrics(c *gin.Context) ([]response.MetricData, error) {
	queryParam := &param.MetricQuery{}
	err := c.Bind(queryParam)
	if err != nil {
		return nil, buzerr.NewBadRequest(err.Error())
	}
	metricDatas := metric.QueryMetricData(queryParam)
	return metricDatas, nil
}
