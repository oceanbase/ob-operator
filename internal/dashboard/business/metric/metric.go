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

package metric

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	logger "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	bizcommon "github.com/oceanbase/ob-operator/internal/dashboard/business/common"
	bizconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	metricconst "github.com/oceanbase/ob-operator/internal/dashboard/business/metric/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/generated/bindata"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/external"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
)

var metricExprConfig map[string]string

func init() {
	metricExprConfig = make(map[string]string)
	metricExprConfigContent, err := bindata.Asset(metricconst.MetricExprConfigFile)
	if err != nil {
		panic(errors.Wrap(err, "load metric expr config failed"))
	}
	err = yaml.Unmarshal(metricExprConfigContent, &metricExprConfig)
	if err != nil {
		panic(errors.Wrap(err, "parse metric expr config data failed"))
	}
}

func ListMetricClasses(scope, language string) ([]response.MetricClass, error) {
	metricClasses := make([]response.MetricClass, 0)
	configFile := metricconst.MetricConfigFileEnUS
	switch language {
	case bizconstant.LANGUAGE_EN_US:
		configFile = metricconst.MetricConfigFileEnUS
	case bizconstant.LANGUAGE_ZH_CN:
		configFile = metricconst.MetricConfigFileZhCN
	default:
		logger.Infof("Not supported language %s, return default", language)
	}

	metricConfigContent, err := bindata.Asset(configFile)
	if err != nil {
		return metricClasses, err
	}
	logger.Debugf("metric configs contents: %s", string(metricConfigContent))
	metricConfigMap := make(map[string][]response.MetricClass)
	// TODO: Do not unmarshal the file every time, cache the result
	err = yaml.Unmarshal(metricConfigContent, &metricConfigMap)
	if err != nil {
		return metricClasses, err
	}
	logger.Debugf("metric configs: %v", metricConfigMap)
	metricClasses, found := metricConfigMap[scope]
	if !found {
		err = errors.Errorf("metric config for scope %s not found", scope)
	}
	return metricClasses, err
}

func replaceQueryVariables(exprTemplate string, labels []common.KVPair, groupLabels []string, step int64) string {
	labelStrParts := make([]string, 0, len(labels))
	for _, label := range labels {
		labelStrParts = append(labelStrParts, fmt.Sprintf("%s=\"%s\"", label.Key, label.Value))
	}
	labelStr := strings.Join(labelStrParts, ",")
	groupLabelStr := strings.Join(groupLabels, ",")
	replacer := strings.NewReplacer(metricconst.KeyInterval, fmt.Sprintf("%ss", strconv.FormatInt(step, 10)), metricconst.KeyLabels, labelStr, metricconst.KeyGroupLabels, groupLabelStr)
	return replacer.Replace(exprTemplate)
}

func extractMetricData(name string, resp *external.PrometheusQueryRangeResponse) []response.MetricData {
	metricDatas := make([]response.MetricData, 0)
	for _, result := range resp.Data.Result {
		values := make([]response.MetricValue, 0)
		metric := response.Metric{
			Name:   name,
			Labels: bizcommon.MapToKVs(result.Metric),
		}
		for _, value := range result.Values {
			t := value[0].(float64)
			v, err := strconv.ParseFloat(value[1].(string), 64)
			if err != nil {
				logger.Warnf("failed to parse value %v", value)
				v = 0
			} else if math.IsNaN(v) {
				logger.Debugf("value at timestamp %f is NaN, set to 0", t)
				v = 0
			}
			values = append(values, response.MetricValue{
				Timestamp: t,
				Value:     v,
			})
		}
		lenValues := len(values)
		if lenValues == 0 {
			continue
		}
		for i := range values {
			// interpolate zero slot with average of previous and next value
			if values[i].Value == 0 {
				switch i {
				case 0:
					if lenValues > 1 {
						values[i].Value = values[i+1].Value
					}
				case lenValues - 1:
					if lenValues > 1 {
						values[i].Value = values[i-1].Value
					}
				default:
					values[i].Value = (values[i-1].Value + values[i+1].Value) / 2
				}
			}
		}
		metricDatas = append(metricDatas, response.MetricData{
			Metric: metric,
			Values: values,
		})
	}
	return metricDatas
}

func QueryMetricData(queryParam *param.MetricQuery) []response.MetricData {
	client := resty.New().SetTimeout(time.Duration(metricconst.DefaultMetricQueryTimeout * time.Second))
	metricDatas := make([]response.MetricData, 0, len(queryParam.Metrics))
	wg := sync.WaitGroup{}
	metricDataCh := make(chan []response.MetricData, len(queryParam.Metrics))
	for _, metric := range queryParam.Metrics {
		exprTemplate, found := metricExprConfig[metric]
		if found {
			wg.Add(1)
			go func(metric string, ch chan []response.MetricData) {
				defer wg.Done()
				expr := replaceQueryVariables(exprTemplate, queryParam.Labels, queryParam.GroupLabels, queryParam.QueryRange.Step)
				logger.Infof("query with expr: %s, range: %v", expr, queryParam.QueryRange)
				queryRangeResp := &external.PrometheusQueryRangeResponse{}
				resp, err := client.R().SetQueryParams(map[string]string{
					"start": strconv.FormatFloat(queryParam.QueryRange.StartTimestamp, 'f', 3, 64),
					"end":   strconv.FormatFloat(queryParam.QueryRange.EndTimestamp, 'f', 3, 64),
					"step":  strconv.FormatInt(queryParam.QueryRange.Step, 10),
					"query": expr,
				}).SetHeader("content-type", "application/json").
					SetResult(queryRangeResp).
					Get(fmt.Sprintf("%s%s", metricconst.PrometheusAddress, metricconst.MetricRangeQueryUrl))
				if err != nil {
					logger.Errorf("Query expression expr got error: %v", err)
				} else if resp.StatusCode() == http.StatusOK {
					ch <- extractMetricData(metric, queryRangeResp)
				}
			}(metric, metricDataCh)
		} else {
			logger.Warnf("Metric %s expression not found", metric)
		}
	}
	wg.Wait()
	close(metricDataCh)
	for metricDataArray := range metricDataCh {
		metricDatas = append(metricDatas, metricDataArray...)
	}
	return metricDatas
}
