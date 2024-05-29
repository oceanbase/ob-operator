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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	alarmconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/alarm/constant"
	metricconst "github.com/oceanbase/ob-operator/internal/dashboard/business/metric/constant"
	rulemodel "github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/rule"
	"github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/prometheus/prometheus/model/rulefmt"
	promv1 "github.com/prometheus/prometheus/web/api/v1"
	logger "github.com/sirupsen/logrus"
)

func CreateOrUpdateRule(rule *rulemodel.Rule) error {
	currentRules, err := ListRules(nil)
	if err != nil {
		return errors.Wrap(err, errors.ErrExternal, "List rules failed")
	}
	configRules := make([]rulefmt.Rule, 0)
	for _, currentRule := range currentRules {
		if rule.Name == currentRule.Name {
			continue
		}
		configRules = append(configRules, *currentRule.ToPromRule())
	}
	configRules = append(configRules, *rule.ToPromRule())
	return updatePrometheusRules(configRules)
}

func DeleteRule(name string) error {
	currentRules, err := ListRules(nil)
	if err != nil {
		return errors.Wrap(err, errors.ErrExternal, "List rules failed")
	}
	configRules := make([]rulefmt.Rule, 0)
	ruleExists := false
	for _, currentRule := range currentRules {
		if name == currentRule.Name {
			ruleExists = true
			continue
		} else {
			configRules = append(configRules, *currentRule.ToPromRule())
		}
	}
	if !ruleExists {
		return errors.NewBadRequest(fmt.Sprintf("Rule %s not exists", name))
	}
	return updatePrometheusRules(configRules)
}

func GetRule(name string) (*rulemodel.RuleResponse, error) {
	rules, err := ListRules(nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrExternal, "Query rules from prometheus")
	}
	for _, rule := range rules {
		if rule.Name == name {
			return &rule, nil
		}
	}
	return nil, errors.New(errors.ErrNotFound, "Rule not found")
}

func ListRules(filter *rulemodel.RuleFilter) ([]rulemodel.RuleResponse, error) {
	client := resty.New().SetTimeout(time.Duration(alarmconstant.DefaultAlarmQueryTimeout * time.Second))
	promRuleResponse := &rulemodel.PromRuleResponse{}
	resp, err := client.R().SetQueryParam("type", "alert").SetHeader("content-type", "application/json").SetResult(promRuleResponse).Get(fmt.Sprintf("%s%s", metricconst.PrometheusAddress, alarmconstant.RuleUrl))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrExternal, "Query rules from prometheus")
	} else if resp.StatusCode() != http.StatusOK {
		return nil, errors.Newf(errors.ErrExternal, "Query rules from prometheus got unexpected status: %d", resp.StatusCode())
	}
	logger.Debugf("Response from prometheus: %v", resp)
	filteredRules := make([]rulemodel.RuleResponse, 0)
	for _, ruleGroup := range promRuleResponse.Data.RuleGroups {
		for _, promRule := range ruleGroup.Rules {
			encodedPromRule, err := json.Marshal(promRule)
			if err != nil {
				logger.Errorf("Got an error when encoding rule %v", promRule)
				continue
			}
			logger.Debugf("Process prometheus rule: %s", string(encodedPromRule))
			alertingRule := &promv1.AlertingRule{}
			err = json.Unmarshal(encodedPromRule, alertingRule)
			if err != nil {
				logger.Errorf("Got an error when decoding rule %v", promRule)
				continue
			}
			ruleResp := rulemodel.NewRuleResponse(alertingRule)
			logger.Debugf("Parsed prometheus rule: %v", ruleResp)
			if filterRule(ruleResp, filter) {
				filteredRules = append(filteredRules, *ruleResp)
			}
		}
	}
	return filteredRules, nil
}

func filterRule(rule *rulemodel.RuleResponse, filter *rulemodel.RuleFilter) bool {
	matched := true
	if filter != nil {
		if filter.Keyword != "" {
			matched = matched && strings.Contains(rule.Name, filter.Keyword)
		}
		if filter.InstanceType != "" {
			matched = matched && (rule.InstanceType == filter.InstanceType)
		}
		if filter.Serverity != "" {
			matched = matched && (rule.Serverity == filter.Serverity)
		}
	}
	return matched
}
