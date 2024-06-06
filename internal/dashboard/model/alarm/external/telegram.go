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

type TelegramConfig struct {
	amconfig.NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	APIUrl               *amconfig.URL `yaml:"api_url" json:"api_url,omitempty"`
	BotToken             string        `yaml:"bot_token,omitempty" json:"token,omitempty"`
	BotTokenFile         string        `yaml:"bot_token_file,omitempty" json:"token_file,omitempty"`
	ChatID               int64         `yaml:"chat_id,omitempty" json:"chat,omitempty"`
	Message              string        `yaml:"message,omitempty" json:"message,omitempty"`
	DisableNotifications bool          `yaml:"disable_notifications,omitempty" json:"disable_notifications,omitempty"`
	ParseMode            string        `yaml:"parse_mode,omitempty" json:"parse_mode,omitempty"`
}
