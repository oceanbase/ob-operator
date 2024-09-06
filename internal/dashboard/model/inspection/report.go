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
	"github.com/oceanbase/ob-operator/internal/dashboard/model/job"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
)

type ResultStatistics struct {
	FailedCount     int `json:"failedCount"`
	CriticalCount   int `json:"criticalCount"`
	ModerateCount   int `json:"moderateCount"`
	NegligibleCount int `json:"negligibleCount"`
}

type ReportBriefInfo struct {
	Id               string                 `json:"id" binding:"required"`
	OBCluster        response.OBClusterMeta `json:"obCluster" binding:"required"`
	Scenario         InspectionScenario     `json:"scenario" binding:"required"`
	ResultStatistics ResultStatistics       `json:"resultStatistics" binding:"required"`
	Status           job.JobStatus          `json:"status" binding:"required"`
	StartTime        int64                  `json:"startTime,omitempty"`
	FinishTime       int64                  `json:"finishTime,omitempty"`
}

type InspectionItem struct {
	Name    string   `json:"name" binding:"required"`
	Results []string `json:"results,omitempty"`
}

type ResultDetail struct {
	CriticalItems   []InspectionItem `json:"criticalItems,omitempty"`
	ModerateItems   []InspectionItem `json:"moderateItems,omitempty"`
	NegligibleItems []InspectionItem `json:"negligibleItems,omitempty"`
	FailedItems     []InspectionItem `json:"failedItems,omitempty"`
}

type Report struct {
	ReportBriefInfo `json:",inline"`
	ResultDetail    ResultDetail `json:"resultDetail,omitempty"`
}
