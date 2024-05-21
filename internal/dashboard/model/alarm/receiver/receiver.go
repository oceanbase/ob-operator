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
	amconfig "github.com/prometheus/alertmanager/config"
	logger "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ReceiverIdentity struct {
	Name string `json:"name" binding:"required"`
}

type Receiver struct {
	Name   string       `json:"name" binding:"required"`
	Type   ReceiverType `json:"type" binding:"required"`
	Config string       `json:"config" binding:"required"`
}

func NewReceiver(amreceiver *amconfig.Receiver) *Receiver {
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
