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

package operator

import (
	"flag"
	"strings"
	"sync"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	cfgOnce sync.Once
	cfg     *Config
)

func newConfig() *Config {
	v := viper.New()
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/admin/oceanbase")
	v.SetConfigName(".ob-operator")
	v.SetConfigType("yaml")

	setDefaultConfigs(v)

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	v.BindPFlags(pflag.CommandLine)

	v.AutomaticEnv()
	v.SetEnvPrefix("OB_OPERATOR")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", ""))

	config := &Config{}
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(err)
		}
	}
	if err := v.Unmarshal(config); err != nil {
		panic(err)
	}
	config.v = v
	return config
}

func GetConfig() *Config {
	if cfg == nil {
		cfgOnce.Do(func() {
			cfg = newConfig()
		})
	}
	return cfg
}

func (c *Config) Write() error {
	return c.v.WriteConfigAs(".ob-operator.yaml")
}
