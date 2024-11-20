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
package rule

type RuleType string

const (
	TypeBuiltin    RuleType = "builtin"
	TypeCustomized RuleType = "customized"
)

type RuleState string

const (
	StateActive   RuleState = "active"
	StateInactive RuleState = "inactive"
)

type RuleHealth string

const (
	HealthUnknown RuleHealth = "unknown"
	HealthOK      RuleHealth = "ok"
	HealthError   RuleHealth = "error"
)