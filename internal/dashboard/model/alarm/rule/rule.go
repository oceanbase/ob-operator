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
	"time"

	alarmconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/alarm/constant"
	bizcommon "github.com/oceanbase/ob-operator/internal/dashboard/business/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/oceanbase"

	prommodel "github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/rulefmt"
	promv1 "github.com/prometheus/prometheus/web/api/v1"
)

type PromRuleResponse struct {
	Status string                `json:"status" binding:"required"`
	Data   *promv1.RuleDiscovery `json:"data" binding:"required"`
}

type Rule struct {
	Name         string                   `json:"name" binding:"required"`
	InstanceType oceanbase.OBInstanceType `json:"instanceType" binding:"required"`
	Type         RuleType                 `json:"type" default:"customized"`
	Query        string                   `json:"query" binding:"required"`
	Duration     int                      `json:"duration" binding:"required"`
	Labels       []common.KVPair          `json:"labels" binding:"required"`
	Severity     alarm.Severity           `json:"severity" binding:"required"`
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

type ConfigRuleGroups struct {
	Groups []ConfigRuleGroup `json:"groups"`
}

type ConfigRuleGroup struct {
	Name  string         `json:"name"`
	Rules []rulefmt.Rule `json:"rules"`
}

func (r *Rule) ToPromRule() *rulefmt.Rule {
	annotations := make(map[string]string)
	annotations[alarmconstant.AnnoSummary] = r.Summary
	annotations[alarmconstant.AnnoDescription] = r.Description
	labels := r.Labels
	labels = append(labels, common.KVPair{
		Key:   alarmconstant.LabelRuleType,
		Value: string(r.Type),
	})
	labels = append(labels, common.KVPair{
		Key:   alarmconstant.LabelSeverity,
		Value: string(r.Severity),
	})
	labels = append(labels, common.KVPair{
		Key:   alarmconstant.LabelRuleName,
		Value: r.Name,
	})
	labels = append(labels, common.KVPair{
		Key:   alarmconstant.LabelInstanceType,
		Value: string(r.InstanceType),
	})
	promRule := &rulefmt.Rule{
		Alert:       r.Name,
		Expr:        r.Query,
		For:         prommodel.Duration(r.Duration * int(time.Second)),
		Labels:      bizcommon.KVsToMap(labels),
		Annotations: annotations,
	}
	return promRule
}

func NewRuleResponse(promRule *promv1.AlertingRule) *RuleResponse {
	var instanceType oceanbase.OBInstanceType
	var ruleType RuleType
	severity := alarm.SeverityInfo
	summary := ""
	description := ""
	labels := make([]common.KVPair, 0, len(promRule.Labels))
	for _, label := range promRule.Labels {
		labels = append(labels, common.KVPair{
			Key:   label.Name,
			Value: label.Value,
		})
		if label.Name == alarmconstant.LabelSeverity {
			severity = alarm.Severity(label.Value)
		}
		if label.Name == alarmconstant.LabelInstanceType {
			instanceType = oceanbase.OBInstanceType(label.Value)
		}
		if label.Name == alarmconstant.LabelRuleType {
			ruleType = RuleType(label.Value)
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
		Type:         ruleType,
		Query:        promRule.Query,
		Duration:     int(promRule.Duration),
		Labels:       labels,
		Severity:     severity,
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
