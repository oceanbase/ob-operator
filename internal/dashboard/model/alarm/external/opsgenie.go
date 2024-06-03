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

type OpsGenieConfig struct {
	amconfig.NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	APIKey       string                             `yaml:"api_key,omitempty" json:"api_key,omitempty"`
	APIKeyFile   string                             `yaml:"api_key_file,omitempty" json:"api_key_file,omitempty"`
	APIURL       *amconfig.URL                      `yaml:"api_url,omitempty" json:"api_url,omitempty"`
	Message      string                             `yaml:"message,omitempty" json:"message,omitempty"`
	Description  string                             `yaml:"description,omitempty" json:"description,omitempty"`
	Source       string                             `yaml:"source,omitempty" json:"source,omitempty"`
	Details      map[string]string                  `yaml:"details,omitempty" json:"details,omitempty"`
	Entity       string                             `yaml:"entity,omitempty" json:"entity,omitempty"`
	Responders   []amconfig.OpsGenieConfigResponder `yaml:"responders,omitempty" json:"responders,omitempty"`
	Actions      string                             `yaml:"actions,omitempty" json:"actions,omitempty"`
	Tags         string                             `yaml:"tags,omitempty" json:"tags,omitempty"`
	Note         string                             `yaml:"note,omitempty" json:"note,omitempty"`
	Priority     string                             `yaml:"priority,omitempty" json:"priority,omitempty"`
	UpdateAlerts bool                               `yaml:"update_alerts,omitempty" json:"update_alerts,omitempty"`
}
