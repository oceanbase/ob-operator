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
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	alarmconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/alarm/constant"
	metricconst "github.com/oceanbase/ob-operator/internal/dashboard/business/metric/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/rule"
	"github.com/oceanbase/ob-operator/pkg/errors"
	promv1 "github.com/prometheus/prometheus/web/api/v1"
	logger "github.com/sirupsen/logrus"
)

func ListRules(filter *rule.RuleFilter) ([]rule.RuleResponse, error) {
	client := resty.New().SetTimeout(time.Duration(alarmconstant.DefaultAlarmQueryTimeout * time.Second))
	ruleDiscovery := &promv1.RuleDiscovery{}
	resp, err := client.R().SetQueryParam("type", "alert").SetHeader("content-type", "application/json").SetResult(ruleDiscovery).Get(fmt.Sprintf("%s%s", metricconst.PrometheusAddress, alarmconstant.RuleUrl))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrExternal, "Query rules from prometheus")
	} else if resp.StatusCode() != http.StatusOK {
		return nil, errors.Newf(errors.ErrExternal, "Query rules from prometheus got unexpected status: %d", resp.StatusCode())
	}
	filteredRules := make([]rule.RuleResponse, 0)
	for _, ruleGroup := range ruleDiscovery.RuleGroups {
		for _, promRule := range ruleGroup.Rules {
			alertingRule, ok := promRule.(promv1.AlertingRule)
			if !ok {
				logger.Errorf("Got an unexpected rule %v", promRule)
			}
			ruleResp := rule.NewRuleResponse(&alertingRule)
			if filterRule(ruleResp, filter) {
				filteredRules = append(filteredRules, *ruleResp)
			}
		}
	}
	return filteredRules, nil
}

func filterRule(rule *rule.RuleResponse, filter *rule.RuleFilter) bool {
	matched := true
	if filter.Keyword != "" {
		matched = matched && strings.Contains(rule.Name, filter.Keyword)
	}
	if filter.InstanceType != "" {
		matched = matched && (rule.InstanceType == filter.InstanceType)
	}
	if filter.Serverity != "" {
		matched = matched && (rule.Serverity == filter.Serverity)
	}
	return matched
}
