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

package model

type OBConfig struct {
	Name    string      `json:"name"`
	Value   interface{} `json:"value"`
	Comment string      `json:"comment"`
}

type ScopedOBConfigs struct {
	Cluster []OBConfig `json:"cluster"`
	Tenant  []OBConfig `json:"tenant"`
}

type OptimizedParameters struct {
	Scenario   string          `json:"scenario"`
	Comment    string          `json:"comment"`
	Parameters ScopedOBConfigs `json:"parameters"`
}

type OptimizedVariables struct {
	Scenario  string          `json:"scenario"`
	Comment   string          `json:"comment"`
	Variables ScopedOBConfigs `json:"variables"`
}

type OptimizationResponse struct {
	Parameters []OBConfig `json:"parameters"`
	Variables  []OBConfig `json:"variables"`
}
