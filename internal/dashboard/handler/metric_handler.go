/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package handler

import (
	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"

	"github.com/oceanbase/ob-operator/internal/dashboard/business/metric"
	metricconst "github.com/oceanbase/ob-operator/internal/dashboard/business/metric/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
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
	switch scope {
	case
		metricconst.ScopeCluster,
		metricconst.ScopeTenant,
		metricconst.ScopeClusterOverview,
		metricconst.ScopeTenantOverview,
		metricconst.ScopeOBProxy,
		metricconst.ScopeOBProxyOverview:
	default:
		return nil, httpErr.NewBadRequest("invalid scope")
	}
	metricClasses, err := metric.ListMetricClasses(scope, language)
	if err != nil {
		return nil, err
	}
	logger.Debugf("List metric classes: %+v", metricClasses)
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
		return nil, httpErr.NewBadRequest(err.Error())
	}
	logger.Infof("Query metric data with param: %+v", queryParam)
	metricDatas := metric.QueryMetricData(queryParam)
	logger.Debugf("Query metric data: %+v", metricDatas)
	return metricDatas, nil
}
