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

package obcluster

// ParameterItem defines the parameter item returned by 'show parameters' command
type ParameterItem struct {
	Zone         string `json:"zone"`
	SvrType      string `json:"svrType" db:"svr_type"`
	SvrIP        string `json:"svrIP" db:"svr_ip"`
	SvrPort      string `json:"svrPort" db:"svr_port"`
	Name         string `json:"name"`
	DataType     string `json:"dataType" db:"data_type"`
	Value        string `json:"value"`
	Info         string `json:"info"`
	Section      string `json:"section"`
	Scope        string `json:"scope"`
	Source       string `json:"source"`
	EditLevel    string `json:"editLevel" db:"edit_level"`
	DefaultValue string `json:"defaultValue" db:"default_value"`
	IsDefault    bool   `json:"isDefault" db:"isdefault"`
}
