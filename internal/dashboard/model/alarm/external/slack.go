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

type SlackConfig struct {
	amconfig.NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	APIURL     *amconfig.URL `yaml:"api_url,omitempty" json:"api_url,omitempty"`
	APIURLFile string        `yaml:"api_url_file,omitempty" json:"api_url_file,omitempty"`

	// Slack channel override, (like #other-channel or @username).
	Channel  string `yaml:"channel,omitempty" json:"channel,omitempty"`
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	Color    string `yaml:"color,omitempty" json:"color,omitempty"`

	Title       string                  `yaml:"title,omitempty" json:"title,omitempty"`
	TitleLink   string                  `yaml:"title_link,omitempty" json:"title_link,omitempty"`
	Pretext     string                  `yaml:"pretext,omitempty" json:"pretext,omitempty"`
	Text        string                  `yaml:"text,omitempty" json:"text,omitempty"`
	Fields      []*amconfig.SlackField  `yaml:"fields,omitempty" json:"fields,omitempty"`
	ShortFields bool                    `yaml:"short_fields" json:"short_fields,omitempty"`
	Footer      string                  `yaml:"footer,omitempty" json:"footer,omitempty"`
	Fallback    string                  `yaml:"fallback,omitempty" json:"fallback,omitempty"`
	CallbackID  string                  `yaml:"callback_id,omitempty" json:"callback_id,omitempty"`
	IconEmoji   string                  `yaml:"icon_emoji,omitempty" json:"icon_emoji,omitempty"`
	IconURL     string                  `yaml:"icon_url,omitempty" json:"icon_url,omitempty"`
	ImageURL    string                  `yaml:"image_url,omitempty" json:"image_url,omitempty"`
	ThumbURL    string                  `yaml:"thumb_url,omitempty" json:"thumb_url,omitempty"`
	LinkNames   bool                    `yaml:"link_names" json:"link_names,omitempty"`
	MrkdwnIn    []string                `yaml:"mrkdwn_in,omitempty" json:"mrkdwn_in,omitempty"`
	Actions     []*amconfig.SlackAction `yaml:"actions,omitempty" json:"actions,omitempty"`
}
