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

type EmailConfig struct {
	amconfig.NotifierConfig `yaml:",inline" json:",inline"`

	// Email address to notify.
	To               string            `yaml:"to,omitempty" json:"to,omitempty"`
	From             string            `yaml:"from,omitempty" json:"from,omitempty"`
	Hello            string            `yaml:"hello,omitempty" json:"hello,omitempty"`
	Smarthost        amconfig.HostPort `yaml:"smarthost,omitempty" json:"smarthost,omitempty"`
	AuthUsername     string            `yaml:"auth_username,omitempty" json:"auth_username,omitempty"`
	AuthPassword     string            `yaml:"auth_password,omitempty" json:"auth_password,omitempty"`
	AuthPasswordFile string            `yaml:"auth_password_file,omitempty" json:"auth_password_file,omitempty"`
	AuthSecret       string            `yaml:"auth_secret,omitempty" json:"auth_secret,omitempty"`
	AuthIdentity     string            `yaml:"auth_identity,omitempty" json:"auth_identity,omitempty"`
	Headers          map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	HTML             string            `yaml:"html,omitempty" json:"html,omitempty"`
	Text             string            `yaml:"text,omitempty" json:"text,omitempty"`
	RequireTLS       *bool             `yaml:"require_tls,omitempty" json:"require_tls,omitempty"`
	TLSConfig        TLSConfig         `yaml:"tls_config,omitempty" json:"tls_config,omitempty"`
}
