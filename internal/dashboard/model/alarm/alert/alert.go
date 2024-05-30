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

	alarmconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/alarm/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/oceanbase"

	ammodels "github.com/prometheus/alertmanager/api/v2/models"
	logger "github.com/sirupsen/logrus"
)

type Status struct {
	InhibitedBy []string `json:"inhibitedBy" binding:"required"`
	SilencedBy  []string `json:"silencedBy" binding:"required"`
	State       State    `json:"state" binding:"required"`
}

type Alert struct {
	Fingerprint string                `json:"fingerprint" binding:"required"`
	Rule        string                `json:"rule" binding:"required"`
	Serverity   alarm.Serverity       `json:"serverity" binding:"required"`
	Instance    *oceanbase.OBInstance `json:"instance" binding:"required"`
	StartsAt    int64                 `json:"startsAt" binding:"required"`
	UpdatedAt   int64                 `json:"updatedAt" binding:"required"`
	EndsAt      int64                 `json:"endsAt" binding:"required"`
	Status      *Status               `json:"status" binding:"required"`
	Labels      []common.KVPair       `json:"labels,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
}

func NewAlert(alert *ammodels.GettableAlert) (*Alert, error) {
	rule := "default-rule"
	rule, ok := alert.Labels[alarmconstant.LabelRuleName]
	if !ok {
		// TODO: return error
		// return nil, errors.New("Convert alert failed, no rule")
		logger.Error("Convert alert failed, no rule label")
	}
	serverity, ok := alert.Labels[alarmconstant.LabelServerity]
	if !ok {
		return nil, errors.New("Convert alert failed, no serverity")
	}
	labels := make([]common.KVPair, 0, len(alert.Labels))
	for k, v := range alert.Labels {
		labels = append(labels, common.KVPair{
			Key:   k,
			Value: v,
		})
	}
	instance := &oceanbase.OBInstance{}
	obcluster, exists := alert.Labels[alarmconstant.LabelOBCluster]
	if exists {
		instance.OBCluster = obcluster
		instance.Type = oceanbase.TypeOBCluster
	}
	observer, exists := alert.Labels[alarmconstant.LabelOBServer]
	if exists {
		instance.OBServer = observer
		instance.Type = oceanbase.TypeOBServer
	}
	obtenant, exists := alert.Labels[alarmconstant.LabelOBTenant]
	if exists {
		instance.OBTenant = obtenant
		instance.Type = oceanbase.TypeOBTenant
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
		Instance:    instance,
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
