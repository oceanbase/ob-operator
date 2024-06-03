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
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	alarmconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/alarm/constant"
	metricconst "github.com/oceanbase/ob-operator/internal/dashboard/business/metric/constant"
	externalmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/external"
	rulemodel "github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/rule"
	"github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"

	"github.com/go-resty/resty/v2"
	"github.com/prometheus/prometheus/model/rulefmt"
	logger "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var restyClient *resty.Client
var restyOnce sync.Once

func getClient() *resty.Client {
	restyOnce.Do(func() {
		restyClient = resty.New().SetTimeout(time.Duration(alarmconstant.DefaultAlarmQueryTimeout * time.Second))
	})
	return restyClient
}

func getAlertmanagerConfig(ctx context.Context) (*externalmodel.Config, error) {
	content, err := readConfigMapContent(ctx, os.Getenv(alarmconstant.EnvConfigNamespace), os.Getenv(alarmconstant.EnvAlertmanagerConfig), alarmconstant.AlertmanagerConfigFile)
	if err != nil {
		logger.WithError(err).Errorf("Failed to read config from configmap")
		return nil, errors.Wrap(err, errors.ErrInternal, "Get alertmanager config failed")
	}
	config := &externalmodel.Config{}
	err = yaml.Unmarshal([]byte(content), config)
	if err != nil {
		logger.WithError(err).Errorf("Got exception when decode content in configmap")
		return nil, errors.Wrap(err, errors.ErrInvalid, "Invalid config, parse failed")
	}
	return config, nil
}

func updateAlertManagerConfig(ctx context.Context, config *externalmodel.Config) error {

	// jsonContent, _ := json.Marshal(config.Receivers)
	// logger.Debugf("Encode receivers %v", string(jsonContent))
	// logger.Debugf("Url %s", string(config.Receivers[0].WebhookConfigs[0].URL.String()))
	// logger.Debugf("Url %s", string(config.Receivers[1].WebhookConfigs[0].URL.String()))
	// logger.Debugf("Url0 scheme %s", string(config.Receivers[0].WebhookConfigs[0].URL.Scheme))
	// logger.Debugf("Url1 scheme %s", string(config.Receivers[1].WebhookConfigs[0].URL.Scheme))
	// // Encode and decode receivers using gob to keep secret properties
	// receiverContent, err := msgpack.Marshal(config.Receivers)
	// if err != nil {
	// 	logger.WithError(err).Error("Encode receivers failed")
	// 	return errors.Wrap(err, errors.ErrInternal, "Encode receivers using msgpack failed")
	// }
	// var receiverMap map[string]interface{}
	// err = msgpack.Unmarshal(receiverContent, &receiverMap)
	// if err != nil {
	// 	logger.WithError(err).Error("Decode receivers failed")
	// 	return errors.Wrap(err, errors.ErrInternal, "Decode receivers using msgpack failed")
	// }

	// // Encode and Decode config using yaml
	// configContent, err := yaml.Marshal(config)
	// if err != nil {
	// 	logger.WithError(err).Error("Encode config failed")
	// 	return errors.Wrap(err, errors.ErrInternal, "Encode config using yaml failed")
	// }
	// var configMap map[string]interface{}
	// err = yaml.Unmarshal(configContent, &configMap)
	// if err != nil {
	// 	logger.WithError(err).Error("Decode config failed")
	// 	return errors.Wrap(err, errors.ErrInternal, "Decode config using yaml failed")
	// }

	// // Set receivers into configMap
	// configMap["receivers"] = receiverMap

	// Encode using yaml to generate actual content to persist
	content, err := yaml.Marshal(config)
	logger.Debugf("Alertmanager config to persist: %s", string(content))
	if err != nil {
		return errors.Wrap(err, errors.ErrInternal, "Encode config using yaml failed")
	}
	err = os.WriteFile(fmt.Sprintf("%s/%s", alarmconstant.AlertmanagerConfigDir, alarmconstant.AlertmanagerConfigFile), content, 0644)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternal, "Write alertmanager config content failed")
	}

	// Reload config to make config take effect
	err = reloadAlertmanager()
	if err != nil {
		return errors.Wrap(err, errors.ErrInternal, "Reload alertmanager config failed")
	}

	// Write to configmap after reload to ensure keeping a valid version of config
	err = persistToConfigMap(ctx, os.Getenv(alarmconstant.EnvConfigNamespace), os.Getenv(alarmconstant.EnvAlertmanagerConfig), alarmconstant.AlertmanagerConfigFile, string(content))
	if err != nil {
		return errors.Wrap(err, errors.ErrInternal, "Persist alertmanager config to configmap failed")
	}
	return nil
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

func updatePrometheusRules(ctx context.Context, configRules []rulefmt.Rule) error {
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
	err = os.WriteFile(fmt.Sprintf("%s/%s", alarmconstant.RuleConfigDir, alarmconstant.RuleConfigFile), content, 0644)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternal, "Write rule content failed")
	}
	// Write to configmap
	err = persistToConfigMap(ctx, os.Getenv(alarmconstant.EnvConfigNamespace), os.Getenv(alarmconstant.EnvPrometheusRuleConfig), alarmconstant.RuleConfigFile, string(content))
	if err != nil {
		return errors.Wrap(err, errors.ErrInternal, "Persist rule to configmap failed")
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

func readConfigMapContent(ctx context.Context, ns, name, key string) (string, error) {
	cm, err := client.GetClient().ClientSet.CoreV1().ConfigMaps(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return "", errors.NewNotFound("ConfigMap not found")
		}
		return "", errors.NewInternal("Failed to get alertmanager config map, err msg: " + err.Error())
	}
	if cm.Data == nil {
		return "", errors.NewInternal("No data in configmap")
	}
	content, found := cm.Data[key]
	if !found {
		return "", errors.NewInternal(fmt.Sprintf("No data for %s in configmap", key))
	}
	return content, nil
}

func persistToConfigMap(ctx context.Context, ns, name, key, content string) error {
	cm, err := client.GetClient().ClientSet.CoreV1().ConfigMaps(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return errors.NewNotFound("ConfigMap not found")
		}
		return errors.NewInternal("Failed to get alertmanager config map, err msg: " + err.Error())
	}
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	cm.Data[key] = content
	_, err = client.GetClient().ClientSet.CoreV1().ConfigMaps(ns).Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return errors.NewInternal("Failed to update alertmanager config map, err msg: " + err.Error())
	}
	return nil
}
