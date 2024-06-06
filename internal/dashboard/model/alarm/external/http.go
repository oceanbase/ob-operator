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
	commoncfg "github.com/prometheus/common/config"
)

type HTTPClientConfig struct {
	BasicAuth       *BasicAuth     `yaml:"basic_auth,omitempty" json:"basic_auth,omitempty"`
	Authorization   *Authorization `yaml:"authorization,omitempty" json:"authorization,omitempty"`
	OAuth2          *OAuth2        `yaml:"oauth2,omitempty" json:"oauth2,omitempty"`
	BearerToken     string         `yaml:"bearer_token,omitempty" json:"bearer_token,omitempty"`
	BearerTokenFile string         `yaml:"bearer_token_file,omitempty" json:"bearer_token_file,omitempty"`
	TLSConfig       TLSConfig      `yaml:"tls_config,omitempty" json:"tls_config,omitempty"`
	FollowRedirects bool           `yaml:"follow_redirects" json:"follow_redirects"`
	EnableHTTP2     bool           `yaml:"enable_http2" json:"enable_http2"`
	ProxyConfig     `yaml:",inline"`
}

type BasicAuth struct {
	Username     string `yaml:"username" json:"username"`
	UsernameFile string `yaml:"username_file,omitempty" json:"username_file,omitempty"`
	Password     string `yaml:"password,omitempty" json:"password,omitempty"`
	PasswordFile string `yaml:"password_file,omitempty" json:"password_file,omitempty"`
}

type Authorization struct {
	Type            string `yaml:"type,omitempty" json:"type,omitempty"`
	Credentials     string `yaml:"credentials,omitempty" json:"credentials,omitempty"`
	CredentialsFile string `yaml:"credentials_file,omitempty" json:"credentials_file,omitempty"`
}

type OAuth2 struct {
	ClientID         string            `yaml:"client_id" json:"client_id"`
	ClientSecret     string            `yaml:"client_secret" json:"client_secret"`
	ClientSecretFile string            `yaml:"client_secret_file" json:"client_secret_file"`
	Scopes           []string          `yaml:"scopes,omitempty" json:"scopes,omitempty"`
	TokenURL         string            `yaml:"token_url" json:"token_url"`
	EndpointParams   map[string]string `yaml:"endpoint_params,omitempty" json:"endpoint_params,omitempty"`
	TLSConfig        TLSConfig         `yaml:"tls_config,omitempty"`
	ProxyConfig      `yaml:",inline"`
}

type TLSConfig struct {
	CA                 string               `yaml:"ca,omitempty" json:"ca,omitempty"`
	Cert               string               `yaml:"cert,omitempty" json:"cert,omitempty"`
	Key                string               `yaml:"key,omitempty" json:"key,omitempty"`
	CAFile             string               `yaml:"ca_file,omitempty" json:"ca_file,omitempty"`
	CertFile           string               `yaml:"cert_file,omitempty" json:"cert_file,omitempty"`
	KeyFile            string               `yaml:"key_file,omitempty" json:"key_file,omitempty"`
	ServerName         string               `yaml:"server_name,omitempty" json:"server_name,omitempty"`
	InsecureSkipVerify bool                 `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
	MinVersion         commoncfg.TLSVersion `yaml:"min_version,omitempty" json:"min_version,omitempty"`
	MaxVersion         commoncfg.TLSVersion `yaml:"max_version,omitempty" json:"max_version,omitempty"`
}

type ProxyConfig struct {
	ProxyURL             amconfig.URL `yaml:"proxy_url,omitempty" json:"proxy_url,omitempty"`
	NoProxy              string       `yaml:"no_proxy,omitempty" json:"no_proxy,omitempty"`
	ProxyFromEnvironment bool         `yaml:"proxy_from_environment,omitempty" json:"proxy_from_environment,omitempty"`
	ProxyConnectHeader   Header       `yaml:"proxy_connect_header,omitempty" json:"proxy_connect_header,omitempty"`
}

type Header map[string][]string
