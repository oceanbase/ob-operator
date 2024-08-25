/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:

	http://license.coscl.org.cn/MulanPSL2

THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/
package utils

import (
	"bytes"
	"fmt"

	"github.com/spf13/viper"

	"github.com/oceanbase/ob-operator/internal/cli/generated/bindata"
)

// component config for test
var component_conf = "internal/assets/cli-templates/component_config.yaml"

func GetComponentsConf() map[string]string {
	components := make(map[string]string)
	fileobj, err := bindata.Asset(component_conf)
	// panic if file not exists
	if err != nil {
		panic(fmt.Errorf("Error reading component config file: %v", err))
	}
	viper.SetConfigType("yaml")
	err = viper.ReadConfig(bytes.NewBuffer(fileobj))
	if err != nil {
		panic(fmt.Errorf("Read Config err:%v", err))
	}
	if err := viper.UnmarshalKey("components", &components); err != nil {
		panic(fmt.Errorf("Error decoding component config file: %v", err))
	}
	return components
}
