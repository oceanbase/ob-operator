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
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	alarmconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/alarm/constant"
	metricconst "github.com/oceanbase/ob-operator/internal/dashboard/business/metric/constant"
	rulemodel "github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/rule"
	"github.com/oceanbase/ob-operator/pkg/errors"
	logger "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	apimodels "github.com/prometheus/alertmanager/api/v2/models"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/prometheus/prometheus/model/rulefmt"
)

func GetAlertmanagerConfig() (*amconfig.Config, error) {
	statusResp := &apimodels.AlertmanagerStatus{}
	client := resty.New().SetTimeout(time.Duration(alarmconstant.DefaultAlarmQueryTimeout * time.Second))
	resp, err := client.R().SetHeader("content-type", "application/json").SetResult(statusResp).Get(fmt.Sprintf("%s%s", alarmconstant.AlertManagerAddress, alarmconstant.StatusUrl))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrExternal, "Query status from alertmanager")
	} else if resp.StatusCode() != http.StatusOK {
		return nil, errors.Newf(errors.ErrExternal, "Query status from alertmanager got unexpected status: %d", resp.StatusCode())
	}
	content := statusResp.Config.Original
	config := &amconfig.Config{}
	err = yaml.Unmarshal([]byte(*content), config)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInvalid, "Invalid config, parse failed")
	}
	return config, nil
}

func updateAlertManagerConfig(config *amconfig.Config) error {
	content, err := yaml.Marshal(config)
	logger.Debugf("Alertmanager config to persist: %s", string(content))
	if err != nil {
		return errors.Wrap(err, errors.ErrInternal, "Encode config content failed")
	}
	err = ioutil.WriteFile(alarmconstant.AlertmanagerConfigFile, content, 0644)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternal, "Write alertmanager config content failed")
	}
	return reloadAlertmanager()
}

func reloadAlertmanager() error {
	client := resty.New().SetTimeout(time.Duration(alarmconstant.DefaultAlarmQueryTimeout * time.Second))
	resp, err := client.R().SetHeader("content-type", "application/json").Post(fmt.Sprintf("%s%s", alarmconstant.AlertManagerAddress, alarmconstant.AlertmanagerReloadUrl))
	if err != nil {
		return errors.Wrap(err, errors.ErrExternal, "Reload alertmanager failed")
	} else if resp.StatusCode() != http.StatusOK {
		return errors.Newf(errors.ErrExternal, "Reload alertmanager got unexpected status: %d", resp.StatusCode())
	}
	return nil
}

func updatePrometheusRules(configRules []rulefmt.Rule) error {
	ruleGroup := rulemodel.ConfigRuleGroup{
		Name:  alarmconstant.OBRuleGroupName,
		Rules: configRules,
	}
	ruleGroups := rulemodel.ConfigRuleGroups{
		Groups: []rulemodel.ConfigRuleGroup{ruleGroup},
	}
	content, err := yaml.Marshal(ruleGroups)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternal, "Encode rule content failed")
	}
	err = ioutil.WriteFile(alarmconstant.RuleConfigFile, content, 0644)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternal, "Write rule content failed")
	}
	return reloadPrometheus()
}

func reloadPrometheus() error {
	client := resty.New().SetTimeout(time.Duration(alarmconstant.DefaultAlarmQueryTimeout * time.Second))
	resp, err := client.R().SetHeader("content-type", "application/json").Post(fmt.Sprintf("%s%s", metricconst.PrometheusAddress, alarmconstant.PrometheusReloadUrl))
	if err != nil {
		return errors.Wrap(err, errors.ErrExternal, "Reload prometheus failed")
	} else if resp.StatusCode() != http.StatusOK {
		return errors.Newf(errors.ErrExternal, "Reload prometheus got unexpected status: %d", resp.StatusCode())
	}
	return nil
}
