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

	alarmconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/alarm/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/generated/bindata"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/receiver"
	"github.com/oceanbase/ob-operator/pkg/errors"

	amconfig "github.com/prometheus/alertmanager/config"
	logger "github.com/sirupsen/logrus"
)

var receiverConfigFiles = map[receiver.ReceiverType]string{
	receiver.TypeDiscord:   "discord_config.yaml",
	receiver.TypeEmail:     "email_config.yaml",
	receiver.TypePagerduty: "pagerduty_config.yaml",
	receiver.TypeSlack:     "slack_config.yaml",
	receiver.TypeWebhook:   "webhook_config.yaml",
	receiver.TypeOpsGenie:  "opsgenie_config.yaml",
	receiver.TypeWechat:    "wechat_config.yaml",
	receiver.TypePushover:  "pushover_config.yaml",
	receiver.TypeVictorOps: "victorops_config.yaml",
	receiver.TypeSNS:       "sns_config.yaml",
	receiver.TypeTelegram:  "telegram_config.yaml",
	receiver.TypeWebex:     "webex_config.yaml",
	receiver.TypeMSTeams:   "msteams_config.yaml",
}

func GetReceiver(ctx context.Context, name string) (*receiver.Receiver, error) {
	receivers, err := ListReceivers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrExternal, "Failed to get receivers")
	}
	for _, receiver := range receivers {
		if receiver.Name == name {
			return &receiver, nil
		}
	}
	return nil, errors.NewNotFound("Receiver not found")
}

func ListReceivers(ctx context.Context) ([]receiver.Receiver, error) {
	config, err := getAlertmanagerConfig(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrExternal, "Failed to get config")
	}
	receivers := make([]receiver.Receiver, 0, len(config.Receivers))
	for _, amreceiver := range config.Receivers {
		receiver := receiver.NewReceiver(&amreceiver)
		if receiver != nil {
			receivers = append(receivers, *receiver)
		}
	}
	return receivers, nil
}

func CreateOrUpdateReceiver(ctx context.Context, r *receiver.Receiver) error {
	config, err := getAlertmanagerConfig(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrExternal, "Failed to get config")
	}

	configReceivers := make([]amconfig.Receiver, 0, len(config.Receivers))
	for _, amreceiver := range config.Receivers {
		if amreceiver.Name == r.Name {
			continue
		}
		configReceivers = append(configReceivers, amreceiver)
	}
	amreceiver, err := r.ToAmReceiver()
	if err != nil {
		return errors.Wrap(err, errors.ErrBadRequest, "Failed to convert receiver to alertmanager's model")
	}
	logger.Debugf("Add receiver %s: %v", amreceiver.Name, amreceiver)
	configReceivers = append(configReceivers, *amreceiver)
	config.Receivers = configReceivers
	return updateAlertManagerConfig(ctx, config)
}

func DeleteReceiver(ctx context.Context, name string) error {
	config, err := getAlertmanagerConfig(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrExternal, "Failed to get config")
	}

	configReceivers := make([]amconfig.Receiver, 0, len(config.Receivers))
	foundReceiver := false
	for _, amreceiver := range config.Receivers {
		if amreceiver.Name == name {
			foundReceiver = true
			continue
		}
		configReceivers = append(configReceivers, amreceiver)
	}
	if !foundReceiver {
		return errors.NewBadRequest(fmt.Sprintf("Receiver %s not exists", name))
	}
	config.Receivers = configReceivers
	return updateAlertManagerConfig(ctx, config)
}

func ListReceiverTemplates() ([]receiver.Template, error) {
	receiverTemplates := make([]receiver.Template, 0)
	for receiverType, configFile := range receiverConfigFiles {
		receiverConfigContent, err := bindata.Asset(fmt.Sprintf("%s/%s", alarmconstant.ReceiverTemplateDir, configFile))
		if err != nil {
			logger.WithError(err).Errorf("Read receiver config of %s failed", receiverType)
			continue
		}
		receiverTemplates = append(receiverTemplates, receiver.Template{
			Type:     receiverType,
			Template: string(receiverConfigContent),
		})
	}
	return receiverTemplates, nil
}

func GetReceiverTemplate(receiverType string) (*receiver.Template, error) {
	configFile, found := receiverConfigFiles[receiver.ReceiverType(receiverType)]
	if !found {
		return nil, errors.NewBadRequest("Receiver type is invalid")
	}
	receiverConfigContent, err := bindata.Asset(fmt.Sprintf("%s/%s", alarmconstant.ReceiverTemplateDir, configFile))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternal, "Get receiver templates failed")
	}
	receiverTemplate := &receiver.Template{
		Type:     receiver.ReceiverType(receiverType),
		Template: string(receiverConfigContent),
	}
	return receiverTemplate, nil
}
