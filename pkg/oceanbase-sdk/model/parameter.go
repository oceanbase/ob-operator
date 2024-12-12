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

type Parameter struct {
	Zone      string `json:"zone" db:"zone"`
	SvrIp     string `json:"svr_ip" db:"svr_ip"`
	SvrPort   int64  `json:"svr_port" db:"svr_port"`
	Name      string `json:"name" db:"name"`
	Value     string `json:"value" db:"value"`
	Scope     string `json:"scope" db:"scope"`
	EditLevel string `json:"edit_level" db:"edit_level"`
	TenantID  int64  `json:"tenant_id" db:"tenant_id"`

	DataType     string `json:"dataType" db:"data_type"`
	Info         string `json:"info"`
	Section      string `json:"section"`
	DefaultValue string `json:"defaultValue" db:"default_value"`
	IsDefault    string `json:"isDefault" db:"isdefault"`
}
