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

package inspection

import (
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
)

type InspectionScheduleStatus string

const (
	ScheduleEnabled  InspectionScheduleStatus = "enabled"
	ScheduleDisabled InspectionScheduleStatus = "disabled"
)

type InspectionScenario string

const (
	ScenarioBasic       InspectionScenario = "basic"
	ScenarioPerformance InspectionScenario = "performance"
)

// TODO: refactor crontab to scheduleExpr
type InspectionScheduleConfig struct {
	Scenario InspectionScenario `json:"scenario" binding:"required"`
	Schedule string             `json:"schedule" binding:"required"`
}

type PolicyMeta struct {
	OBCluster       *response.OBClusterMetaBasic `json:"obCluster" binding:"required"`
	Status          InspectionScheduleStatus     `json:"status" binding:"required"`
	ScheduleConfigs []InspectionScheduleConfig   `json:"scheduleConfig,omitempty"`
}

type Policy struct {
	PolicyMeta    `json:",inline"`
	LatestReports []ReportBriefInfo `json:"latestReports,omitempty"`
}
