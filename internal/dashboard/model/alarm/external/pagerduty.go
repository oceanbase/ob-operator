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

type PagerdutyConfig struct {
	amconfig.NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	ServiceKey     string                    `yaml:"service_key,omitempty" json:"service_key,omitempty"`
	ServiceKeyFile string                    `yaml:"service_key_file,omitempty" json:"service_key_file,omitempty"`
	RoutingKey     string                    `yaml:"routing_key,omitempty" json:"routing_key,omitempty"`
	RoutingKeyFile string                    `yaml:"routing_key_file,omitempty" json:"routing_key_file,omitempty"`
	URL            *amconfig.URL             `yaml:"url,omitempty" json:"url,omitempty"`
	Client         string                    `yaml:"client,omitempty" json:"client,omitempty"`
	ClientURL      string                    `yaml:"client_url,omitempty" json:"client_url,omitempty"`
	Description    string                    `yaml:"description,omitempty" json:"description,omitempty"`
	Details        map[string]string         `yaml:"details,omitempty" json:"details,omitempty"`
	Images         []amconfig.PagerdutyImage `yaml:"images,omitempty" json:"images,omitempty"`
	Links          []amconfig.PagerdutyLink  `yaml:"links,omitempty" json:"links,omitempty"`
	Source         string                    `yaml:"source,omitempty" json:"source,omitempty"`
	Severity       string                    `yaml:"severity,omitempty" json:"severity,omitempty"`
	Class          string                    `yaml:"class,omitempty" json:"class,omitempty"`
	Component      string                    `yaml:"component,omitempty" json:"component,omitempty"`
	Group          string                    `yaml:"group,omitempty" json:"group,omitempty"`
}
