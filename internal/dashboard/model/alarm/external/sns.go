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

type SNSConfig struct {
	amconfig.NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	APIUrl      string            `yaml:"api_url,omitempty" json:"api_url,omitempty"`
	Sigv4       SigV4Config       `yaml:"sigv4" json:"sigv4"`
	TopicARN    string            `yaml:"topic_arn,omitempty" json:"topic_arn,omitempty"`
	PhoneNumber string            `yaml:"phone_number,omitempty" json:"phone_number,omitempty"`
	TargetARN   string            `yaml:"target_arn,omitempty" json:"target_arn,omitempty"`
	Subject     string            `yaml:"subject,omitempty" json:"subject,omitempty"`
	Message     string            `yaml:"message,omitempty" json:"message,omitempty"`
	Attributes  map[string]string `yaml:"attributes,omitempty" json:"attributes,omitempty"`
}

type SigV4Config struct {
	Region    string `yaml:"region,omitempty"`
	AccessKey string `yaml:"access_key,omitempty"`
	SecretKey string `yaml:"secret_key,omitempty"`
	Profile   string `yaml:"profile,omitempty"`
	RoleARN   string `yaml:"role_arn,omitempty"`
}
