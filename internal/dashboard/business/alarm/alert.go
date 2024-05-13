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

package alarm

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	alarmconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/alarm/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/alert"
	"github.com/oceanbase/ob-operator/pkg/errors"

	apimodels "github.com/prometheus/alertmanager/api/v2/models"
)

func ListAlerts(filter *alert.AlertFilter) ([]alert.Alert, error) {
	client := resty.New().SetTimeout(time.Duration(alarmconstant.DefaultAlarmQueryTimeout * time.Second))
	gettableAlerts := make([]*apimodels.GettableAlert, 0)
	resp, err := client.R().SetQueryParams(map[string]string{
		"active":      "true",
		"silenced":    "true",
		"inhibited":   "true",
		"unprocessed": "true",
		"receiver":    "",
	}).SetHeader("content-type", "application/json").SetResult(gettableAlerts).Get(fmt.Sprintf("%s%s", alarmconstant.AlertManagerAddress, alarmconstant.AlertQueryUrl))

	if err != nil || resp.StatusCode() != http.StatusOK {
		return nil, errors.Wrap(err, errors.ErrExternal, "Query alerts from alertmanager")
	}

	filteredAlerts := filterAlerts(gettableAlerts, filter)

	return filteredAlerts, nil
}

func filterAlerts(alerts apimodels.GettableAlerts, filter *alert.AlertFilter) []alert.Alert {
	filteredAlerts := make([]alert.Alert, 0)
	for _, alert := range alerts {
		if matchAlert(alert, filter) {
			filteredAlerts = append(filteredAlerts, alert.NewAlert(alert))
		}
	}
	return filteredAlerts
}

func matchAlert(alert apimodels.GettableAlert, filter *alert.AlertFilter) bool {

}

// func QueryMetricData(queryParam *param.MetricQuery) []response.MetricData {
// 	client := resty.New().SetTimeout(time.Duration(metricconst.DefaultMetricQueryTimeout * time.Second))
// 	metricDatas := make([]response.MetricData, 0, len(queryParam.Metrics))
// 	wg := sync.WaitGroup{}
// 	metricDataCh := make(chan []response.MetricData, len(queryParam.Metrics))
// 	for _, metric := range queryParam.Metrics {
// 		exprTemplate, found := metricExprConfig[metric]
// 		if found {
// 			wg.Add(1)
// 			go func(metric string, ch chan []response.MetricData) {
// 				defer wg.Done()
// 				expr := replaceQueryVariables(exprTemplate, queryParam.Labels, queryParam.GroupLabels, queryParam.QueryRange.Step)
// 				logger.Infof("Query with expr: %s, range: %v", expr, queryParam.QueryRange)
// 				queryRangeResp := &external.PrometheusQueryRangeResponse{}
// 				resp, err := client.R().SetQueryParams(map[string]string{
// 					"start": strconv.FormatFloat(queryParam.QueryRange.StartTimestamp, 'f', 3, 64),
// 					"end":   strconv.FormatFloat(queryParam.QueryRange.EndTimestamp, 'f', 3, 64),
// 					"step":  strconv.FormatInt(queryParam.QueryRange.Step, 10),
// 					"query": expr,
// 				}).SetHeader("content-type", "application/json").
// 					SetResult(queryRangeResp).
// 					Get(fmt.Sprintf("%s%s", metricconst.PrometheusAddress, metricconst.MetricRangeQueryUrl))
// 				if err != nil {
// 					logger.Errorf("Query expression expr got error: %v", err)
// 				} else if resp.StatusCode() == http.StatusOK {
// 					ch <- extractMetricData(metric, queryRangeResp)
// 				}
// 			}(metric, metricDataCh)
// 		} else {
// 			logger.Warnf("Metric %s expression not found", metric)
// 		}
// 	}
// 	wg.Wait()
// 	close(metricDataCh)
// 	for metricDataArray := range metricDataCh {
// 		metricDatas = append(metricDatas, metricDataArray...)
// 	}
// 	return metricDatas
// }
