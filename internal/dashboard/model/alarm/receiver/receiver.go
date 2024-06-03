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
package receiver

import (
	"fmt"

	externalmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/external"
	"github.com/oceanbase/ob-operator/pkg/errors"

	logger "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ReceiverIdentity struct {
	Name string `json:"name" binding:"required"`
}

type ReceiverTemplateIdentity struct {
	Type string `json:"type" binding:"required"`
}

type Receiver struct {
	Name   string       `json:"name" binding:"required"`
	Type   ReceiverType `json:"type" binding:"required"`
	Config string       `json:"config" binding:"required"`
}

func (r *Receiver) ToAmReceiver() (*externalmodel.Receiver, error) {
	amreceiver := &externalmodel.Receiver{
		Name: r.Name,
	}
	var err error
	switch r.Type {
	case TypeDiscord:
		config := &externalmodel.DiscordConfig{}
		err = yaml.Unmarshal([]byte(r.Config), config)
		if err != nil {
			return nil, err
		}
		amreceiver.DiscordConfigs = []*externalmodel.DiscordConfig{config}
	case TypeEmail:
		config := &externalmodel.EmailConfig{}
		err = yaml.Unmarshal([]byte(r.Config), config)
		if err != nil {
			return nil, err
		}
		amreceiver.EmailConfigs = []*externalmodel.EmailConfig{config}
	case TypePagerduty:
		config := &externalmodel.PagerdutyConfig{}
		err = yaml.Unmarshal([]byte(r.Config), config)
		if err != nil {
			return nil, err
		}
		amreceiver.PagerdutyConfigs = []*externalmodel.PagerdutyConfig{config}
	case TypeSlack:
		config := &externalmodel.SlackConfig{}
		err = yaml.Unmarshal([]byte(r.Config), config)
		if err != nil {
			return nil, err
		}
		amreceiver.SlackConfigs = []*externalmodel.SlackConfig{config}
	case TypeWebhook:
		config := &externalmodel.WebhookConfig{}
		err = yaml.Unmarshal([]byte(r.Config), config)
		if err != nil {
			return nil, err
		}
		amreceiver.WebhookConfigs = []*externalmodel.WebhookConfig{config}
	case TypeOpsGenie:
		config := &externalmodel.OpsGenieConfig{}
		err = yaml.Unmarshal([]byte(r.Config), config)
		if err != nil {
			return nil, err
		}
		amreceiver.OpsGenieConfigs = []*externalmodel.OpsGenieConfig{config}
	case TypeWechat:
		config := &externalmodel.WechatConfig{}
		err = yaml.Unmarshal([]byte(r.Config), config)
		if err != nil {
			return nil, err
		}
		amreceiver.WechatConfigs = []*externalmodel.WechatConfig{config}
	case TypePushover:
		config := &externalmodel.PushoverConfig{}
		err = yaml.Unmarshal([]byte(r.Config), config)
		if err != nil {
			return nil, err
		}
		amreceiver.PushoverConfigs = []*externalmodel.PushoverConfig{config}
	case TypeVictorOps:
		config := &externalmodel.VictorOpsConfig{}
		err = yaml.Unmarshal([]byte(r.Config), config)
		if err != nil {
			return nil, err
		}
		amreceiver.VictorOpsConfigs = []*externalmodel.VictorOpsConfig{config}
	case TypeSNS:
		config := &externalmodel.SNSConfig{}
		err = yaml.Unmarshal([]byte(r.Config), config)
		if err != nil {
			return nil, err
		}
		amreceiver.SNSConfigs = []*externalmodel.SNSConfig{config}
	case TypeTelegram:
		config := &externalmodel.TelegramConfig{}
		err = yaml.Unmarshal([]byte(r.Config), config)
		if err != nil {
			return nil, err
		}
		amreceiver.TelegramConfigs = []*externalmodel.TelegramConfig{config}
	case TypeWebex:
		config := &externalmodel.WebexConfig{}
		err = yaml.Unmarshal([]byte(r.Config), config)
		if err != nil {
			return nil, err
		}
		amreceiver.WebexConfigs = []*externalmodel.WebexConfig{config}
	case TypeMSTeams:
		config := &externalmodel.MSTeamsConfig{}
		err = yaml.Unmarshal([]byte(r.Config), config)
		if err != nil {
			return nil, err
		}
		amreceiver.MSTeamsConfigs = []*externalmodel.MSTeamsConfig{config}
	default:
		return nil, errors.NewBadRequest(fmt.Sprintf("Type %s not found", r.Type))
	}
	return amreceiver, nil
}

func NewReceiver(amreceiver *externalmodel.Receiver) *Receiver {
	foundConfig := false
	receiver := &Receiver{
		Name: amreceiver.Name,
	}
	if len(amreceiver.DiscordConfigs) > 0 {
		receiver.Type = TypeDiscord
		config, err := yaml.Marshal(amreceiver.DiscordConfigs[0])
		if err != nil {
			logger.WithError(err).Errorf("Serialize receiver config error, %v", amreceiver.DiscordConfigs[0])
		} else {
			receiver.Config = string(config)
			foundConfig = true
		}
	}
	if len(amreceiver.EmailConfigs) > 0 {
		receiver.Type = TypeEmail
		config, err := yaml.Marshal(amreceiver.EmailConfigs[0])
		if err != nil {
			logger.WithError(err).Errorf("Serialize receiver config error, %v", amreceiver.EmailConfigs[0])
		} else {
			receiver.Config = string(config)
			foundConfig = true
		}
	}
	if len(amreceiver.PagerdutyConfigs) > 0 {
		receiver.Type = TypePagerduty
		config, err := yaml.Marshal(amreceiver.PagerdutyConfigs[0])
		if err != nil {
			logger.WithError(err).Errorf("Serialize receiver config error, %v", amreceiver.PagerdutyConfigs[0])
		} else {
			receiver.Config = string(config)
			foundConfig = true
		}
	}
	if len(amreceiver.SlackConfigs) > 0 {
		receiver.Type = TypeSlack
		config, err := yaml.Marshal(amreceiver.SlackConfigs[0])
		if err != nil {
			logger.WithError(err).Errorf("Serialize receiver config error, %v", amreceiver.SlackConfigs[0])
		} else {
			receiver.Config = string(config)
			foundConfig = true
		}
	}
	if len(amreceiver.WebhookConfigs) > 0 {
		receiver.Type = TypeWebhook
		config, err := yaml.Marshal(amreceiver.WebhookConfigs[0])
		if err != nil {
			logger.WithError(err).Errorf("Serialize receiver config error, %v", amreceiver.WebhookConfigs[0])
		} else {
			receiver.Config = string(config)
			foundConfig = true
		}
	}
	if len(amreceiver.OpsGenieConfigs) > 0 {
		receiver.Type = TypeOpsGenie
		config, err := yaml.Marshal(amreceiver.OpsGenieConfigs[0])
		if err != nil {
			logger.WithError(err).Errorf("Serialize receiver config error, %v", amreceiver.OpsGenieConfigs[0])
		} else {
			receiver.Config = string(config)
			foundConfig = true
		}
	}
	if len(amreceiver.WechatConfigs) > 0 {
		receiver.Type = TypeWechat
		config, err := yaml.Marshal(amreceiver.WechatConfigs[0])
		if err != nil {
			logger.WithError(err).Errorf("Serialize receiver config error, %v", amreceiver.WechatConfigs[0])
		} else {
			receiver.Config = string(config)
			foundConfig = true
		}
	}
	if len(amreceiver.PushoverConfigs) > 0 {
		receiver.Type = TypePushover
		config, err := yaml.Marshal(amreceiver.PushoverConfigs[0])
		if err != nil {
			logger.WithError(err).Errorf("Serialize receiver config error, %v", amreceiver.PushoverConfigs[0])
		} else {
			receiver.Config = string(config)
			foundConfig = true
		}
	}
	if len(amreceiver.VictorOpsConfigs) > 0 {
		receiver.Type = TypeVictorOps
		config, err := yaml.Marshal(amreceiver.VictorOpsConfigs[0])
		if err != nil {
			logger.WithError(err).Errorf("Serialize receiver config error, %v", amreceiver.VictorOpsConfigs[0])
		} else {
			receiver.Config = string(config)
			foundConfig = true
		}
	}
	if len(amreceiver.SNSConfigs) > 0 {
		receiver.Type = TypeSNS
		config, err := yaml.Marshal(amreceiver.SNSConfigs[0])
		if err != nil {
			logger.WithError(err).Errorf("Serialize receiver config error, %v", amreceiver.SNSConfigs[0])
		} else {
			receiver.Config = string(config)
			foundConfig = true
		}
	}
	if len(amreceiver.TelegramConfigs) > 0 {
		receiver.Type = TypeTelegram
		config, err := yaml.Marshal(amreceiver.TelegramConfigs[0])
		if err != nil {
			logger.WithError(err).Errorf("Serialize receiver config error, %v", amreceiver.TelegramConfigs[0])
		} else {
			receiver.Config = string(config)
			foundConfig = true
		}
	}
	if len(amreceiver.WebexConfigs) > 0 {
		receiver.Type = TypeWebex
		config, err := yaml.Marshal(amreceiver.WebexConfigs[0])
		if err != nil {
			logger.WithError(err).Errorf("Serialize receiver config error, %v", amreceiver.WebexConfigs[0])
		} else {
			receiver.Config = string(config)
			foundConfig = true
		}
	}
	if len(amreceiver.MSTeamsConfigs) > 0 {
		receiver.Type = TypeMSTeams
		config, err := yaml.Marshal(amreceiver.MSTeamsConfigs[0])
		if err != nil {
			logger.WithError(err).Errorf("Serialize receiver config error, %v", amreceiver.MSTeamsConfigs[0])
		} else {
			receiver.Config = string(config)
			foundConfig = true
		}
	}
	if foundConfig {
		return receiver
	}
	return nil
}
