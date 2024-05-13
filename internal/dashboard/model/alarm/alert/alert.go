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

package alert

import (
	"errors"
	"time"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/oceanbase"

	alarmconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/alarm/constant"

	apimodels "github.com/prometheus/alertmanager/api/v2/models"
)

type Status struct {
	InhibitedBy []string `json:"inhibitedBy" binding:"required"`
	SilencedBy  []string `json:"silencedBy" binding:"required"`
	State       State    `json:"state" binding:"required"`
}

type Alert struct {
	Fingerprint string               `json:"fingerprint" binding:"required"`
	Rule        string               `json:"rule" binding:"required"`
	Serverity   alarm.Serverity      `json:"serverity" binding:"required"`
	Instance    oceanbase.OBInstance `json:"instance" binding:"required"`
	StartsAt    int64                `json:"startsAt" binding:"required"`
	UpdatedAt   int64                `json:"updatedAt" binding:"required"`
	EndsAt      int64                `json:"endsAt" binding:"required"`
	Status      *Status              `json:"status" binding:"required"`
	Labels      []common.KVPair      `json:"labels,omitempty"`
	Summary     string               `json:"summary,omitempty"`
	Description string               `json:"description,omitempty"`
}

func NewAlert(alert *apimodels.GettableAlert) (*Alert, error) {
	rule, ok := alert.Labels[alarmconstant.LabelRuleName]
	if !ok {
		return nil, errors.New("Convert alert failed")
	}
	serverity, ok := alert.Labels[alarmconstant.LabelServerity]
	if !ok {
		return nil, errors.New("Convert alert failed")
	}
	labels := make([]common.KVPair, 0, len(alert.Labels))
	for k, v := range alert.Labels {
		labels = append(labels, common.KVPair{
			Key:   k,
			Value: v,
		})
	}
	summary, ok := alert.Annotations[alarmconstant.AnnoSummary]
	if !ok {
		return nil, errors.New("No summary info")
	}
	description, ok := alert.Annotations[alarmconstant.AnnoDescription]
	if !ok {
		return nil, errors.New("No description info")
	}
	return &Alert{
		Fingerprint: *alert.Fingerprint,
		Rule:        rule,
		Serverity:   alarm.Serverity(serverity),
		StartsAt:    time.Time(*alert.StartsAt).Unix(),
		UpdatedAt:   time.Time(*alert.UpdatedAt).Unix(),
		EndsAt:      time.Time(*alert.EndsAt).Unix(),
		Status: &Status{
			InhibitedBy: alert.Status.InhibitedBy,
			SilencedBy:  alert.Status.SilencedBy,
			State:       State(*alert.Status.State),
		},
		Labels:      labels,
		Summary:     summary,
		Description: description,
	}, nil
}
