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
	"time"

	amconfig "github.com/prometheus/alertmanager/config"
)

type PushoverConfig struct {
	amconfig.NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	UserKey     string        `yaml:"user_key,omitempty" json:"user_key,omitempty"`
	UserKeyFile string        `yaml:"user_key_file,omitempty" json:"user_key_file,omitempty"`
	Token       string        `yaml:"token,omitempty" json:"token,omitempty"`
	TokenFile   string        `yaml:"token_file,omitempty" json:"token_file,omitempty"`
	Title       string        `yaml:"title,omitempty" json:"title,omitempty"`
	Message     string        `yaml:"message,omitempty" json:"message,omitempty"`
	URL         string        `yaml:"url,omitempty" json:"url,omitempty"`
	URLTitle    string        `yaml:"url_title,omitempty" json:"url_title,omitempty"`
	Device      string        `yaml:"device,omitempty" json:"device,omitempty"`
	Sound       string        `yaml:"sound,omitempty" json:"sound,omitempty"`
	Priority    string        `yaml:"priority,omitempty" json:"priority,omitempty"`
	Retry       time.Duration `yaml:"retry,omitempty" json:"retry,omitempty"`
	Expire      time.Duration `yaml:"expire,omitempty" json:"expire,omitempty"`
	TTL         time.Duration `yaml:"ttl,omitempty" json:"ttl,omitempty"`
	HTML        bool          `yaml:"html" json:"html,omitempty"`
}
