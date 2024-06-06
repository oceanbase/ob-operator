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

type VictorOpsConfig struct {
	amconfig.NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	APIKey            string            `yaml:"api_key,omitempty" json:"api_key,omitempty"`
	APIKeyFile        string            `yaml:"api_key_file,omitempty" json:"api_key_file,omitempty"`
	APIURL            *amconfig.URL     `yaml:"api_url" json:"api_url"`
	RoutingKey        string            `yaml:"routing_key" json:"routing_key"`
	MessageType       string            `yaml:"message_type" json:"message_type"`
	StateMessage      string            `yaml:"state_message" json:"state_message"`
	EntityDisplayName string            `yaml:"entity_display_name" json:"entity_display_name"`
	MonitoringTool    string            `yaml:"monitoring_tool" json:"monitoring_tool"`
	CustomFields      map[string]string `yaml:"custom_fields,omitempty" json:"custom_fields,omitempty"`
}
