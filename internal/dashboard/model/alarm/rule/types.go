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

import (
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/oceanbase"
)

type RuleType string

const (
	TypeBuiltin    RuleType = "builtin"
	TypeCustomized          = "customized"
)

type RuleState string

const (
	StateActive   RuleState = "active"
	StateInactive           = "inactive"
)

type RuleHealth string

const (
	HealthUnknown RuleHealth = "unknown"
	HealthOK                 = "ok"
	HealthError              = "error"
)

type Rule struct {
	Name         string                   `json:"name" binding:"required"`
	InstanceType oceanbase.OBInstanceType `json:"instanceType" binding:"required"`
	Type         RuleType                 `json:"type" binding:"required"`
	Query        string                   `json:"query" binding:"required"`
	Duration     int                      `json:"duration" binding:"required"`
	Labels       common.KVPair            `json:"labels" binding:"required"`
	Serverity    alarm.Serverity          `json:"serverity" binding:"required"`
	Summary      string                   `json:"summary" binding:"required"`
	Description  string                   `json:"description" binding:"required"`
}
