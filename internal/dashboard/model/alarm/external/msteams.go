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

type MSTeamsConfig struct {
	amconfig.NotifierConfig `yaml:",inline" json:",inline"`
	HTTPConfig              *HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`
	WebhookURL              *amconfig.URL     `yaml:"webhook_url,omitempty" json:"webhook_url,omitempty"`
	WebhookURLFile          string            `yaml:"webhook_url_file,omitempty" json:"webhook_url_file,omitempty"`

	Title   string `yaml:"title,omitempty" json:"title,omitempty"`
	Summary string `yaml:"summary,omitempty" json:"summary,omitempty"`
	Text    string `yaml:"text,omitempty" json:"text,omitempty"`
}
