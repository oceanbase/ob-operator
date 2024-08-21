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
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type ComponentVersions struct {
	Components map[string]string `yaml:"components"`
}

// filePath for test
var filePath = "internal/cli/LATEST_VERSION.yaml"

func GetComponentsConf() map[string]string {
	var components ComponentVersions
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(fmt.Errorf("Error reading LATEST_VERSION file: %v", err))
	}

	err = yaml.Unmarshal(data, &components)
	// panic if file not exists
	if err != nil {
		panic(fmt.Errorf("Error decoding LATEST_VERSION file: %v", err))
	}
	return components.Components
}
