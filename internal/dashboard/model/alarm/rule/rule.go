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
	alarmconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/alarm/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/oceanbase"

	promv1 "github.com/prometheus/prometheus/web/api/v1"
)

type Rule struct {
	Name         string                   `json:"name" binding:"required"`
	InstanceType oceanbase.OBInstanceType `json:"instanceType" binding:"required"`
	Type         RuleType                 `json:"type" binding:"required"`
	Query        string                   `json:"query" binding:"required"`
	Duration     int                      `json:"duration" binding:"required"`
	Labels       []common.KVPair          `json:"labels" binding:"required"`
	Serverity    alarm.Serverity          `json:"serverity" binding:"required"`
	Summary      string                   `json:"summary" binding:"required"`
	Description  string                   `json:"description" binding:"required"`
}

type RuleResponse struct {
	State          RuleState  `json:"state" binding:"required"`
	KeepFiringFor  int        `json:"keepFiringFor" binding:"required"`
	Health         RuleHealth `json:"health" binding:"required"`
	LastEvaluation int64      `json:"lastEvaluation" binding:"required"`
	EvaluationTime float64    `json:"evaluationTime" binding:"required"`
	LastError      string     `json:"lastError,omitempty"`
	Rule
}

type RuleIdentity struct {
	Name string `json:"name" binding:"required"`
}

func NewRuleResponse(promRule *promv1.AlertingRule) *RuleResponse {
	var instanceType oceanbase.OBInstanceType
	serverity := alarm.ServerityInfo
	summary := ""
	description := ""
	labels := make([]common.KVPair, 0, len(promRule.Labels))
	for _, label := range promRule.Labels {
		labels = append(labels, common.KVPair{
			Key:   label.Name,
			Value: label.Value,
		})
		if label.Name == alarmconstant.LabelServerity {
			serverity = alarm.Serverity(label.Value)
		}
		if label.Name == alarmconstant.LabelInstanceType {
			instanceType = oceanbase.OBInstanceType(label.Value)
		}
	}
	for _, annotation := range promRule.Annotations {
		if annotation.Name == alarmconstant.AnnoSummary {
			summary = annotation.Value
		}
		if annotation.Name == alarmconstant.AnnoDescription {
			description = annotation.Value
		}
	}
	rule := &Rule{
		Name:         promRule.Name,
		InstanceType: instanceType,
		Type:         TypeBuiltin,
		Query:        promRule.Query,
		Duration:     int(promRule.Duration),
		Labels:       labels,
		Serverity:    serverity,
		Summary:      summary,
		Description:  description,
	}
	return &RuleResponse{
		State:          RuleState(promRule.State),
		KeepFiringFor:  int(promRule.KeepFiringFor),
		Health:         RuleHealth(promRule.Health),
		LastEvaluation: promRule.LastEvaluation.Unix(),
		EvaluationTime: promRule.EvaluationTime,
		LastError:      promRule.LastError,
		Rule:           *rule,
	}
}
