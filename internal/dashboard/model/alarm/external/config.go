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
package external

import (
	amconfig "github.com/prometheus/alertmanager/config"
)

type Config struct {
	Global       *amconfig.GlobalConfig `yaml:"global,omitempty" json:"global,omitempty"`
	Route        *amconfig.Route        `yaml:"route,omitempty" json:"route,omitempty"`
	InhibitRules []amconfig.InhibitRule `yaml:"inhibit_rules,omitempty" json:"inhibit_rules,omitempty"`
	Receivers    []Receiver             `yaml:"receivers,omitempty" json:"receivers,omitempty"`
	Templates    []string               `yaml:"templates" json:"templates"`
	// Deprecated. Remove before v1.0 release.
	MuteTimeIntervals []amconfig.MuteTimeInterval `yaml:"mute_time_intervals,omitempty" json:"mute_time_intervals,omitempty"`
	TimeIntervals     []amconfig.TimeInterval     `yaml:"time_intervals,omitempty" json:"time_intervals,omitempty"`
}

type Receiver struct {
	// A unique identifier for this receiver.
	Name string `yaml:"name" json:"name"`

	DiscordConfigs   []*DiscordConfig   `yaml:"discord_configs,omitempty" json:"discord_configs,omitempty"`
	EmailConfigs     []*EmailConfig     `yaml:"email_configs,omitempty" json:"email_configs,omitempty"`
	PagerdutyConfigs []*PagerdutyConfig `yaml:"pagerduty_configs,omitempty" json:"pagerduty_configs,omitempty"`
	SlackConfigs     []*SlackConfig     `yaml:"slack_configs,omitempty" json:"slack_configs,omitempty"`
	WebhookConfigs   []*WebhookConfig   `yaml:"webhook_configs,omitempty" json:"webhook_configs,omitempty"`
	OpsGenieConfigs  []*OpsGenieConfig  `yaml:"opsgenie_configs,omitempty" json:"opsgenie_configs,omitempty"`
	WechatConfigs    []*WechatConfig    `yaml:"wechat_configs,omitempty" json:"wechat_configs,omitempty"`
	PushoverConfigs  []*PushoverConfig  `yaml:"pushover_configs,omitempty" json:"pushover_configs,omitempty"`
	VictorOpsConfigs []*VictorOpsConfig `yaml:"victorops_configs,omitempty" json:"victorops_configs,omitempty"`
	SNSConfigs       []*SNSConfig       `yaml:"sns_configs,omitempty" json:"sns_configs,omitempty"`
	TelegramConfigs  []*TelegramConfig  `yaml:"telegram_configs,omitempty" json:"telegram_configs,omitempty"`
	WebexConfigs     []*WebexConfig     `yaml:"webex_configs,omitempty" json:"webex_configs,omitempty"`
	MSTeamsConfigs   []*MSTeamsConfig   `yaml:"msteams_configs,omitempty" json:"msteams_configs,omitempty"`
}
