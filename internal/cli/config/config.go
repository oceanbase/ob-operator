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
package config

import (
	"bytes"
	"fmt"

	"github.com/spf13/viper"

	"github.com/oceanbase/ob-operator/internal/cli/generated/bindata"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
)

// component config for test
var confPath = "internal/assets/cli-templates/component_config.yaml"

func readComponentConf(path string) map[string]string {
	components := make(map[string]string)
	fileobj, err := bindata.Asset(path)
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

// GetAllComponents returns all the components
func GetAllComponents() map[string]string {
	return readComponentConf(confPath)
}

// GetDefaultComponents returns the default components to be installed
func GetDefaultComponents() map[string]string {
	var componentsList []string
	components := GetAllComponents()
	defaultComponents := make(map[string]string) // Initialize the map
	if !utils.CheckIfComponentExists("cert-manager") {
		componentsList = []string{"cert-manager", "ob-operator", "ob-dashboard"}
	} else {
		componentsList = []string{"ob-operator", "ob-dashboard"}
	}
	for _, component := range componentsList {
		defaultComponents[component] = components[component]
	}
	return defaultComponents
}
