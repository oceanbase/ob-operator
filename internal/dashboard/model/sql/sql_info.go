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

package sql

import (
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
)

type SqlStatisticMetric struct {
	Name  string  `json:"name" binding:"required"`
	Value float64 `json:"value" binding:"required"`
}

type SqlDiagnoseInfo struct {
	Reason     string `json:"reason" binding:"required"`
	Suggestion string `json:"suggestion,omitempty"`
}

type SqlMetaInfo struct {
	OBServer string `json:"observer" binding:"required"`
	Tenant   string `json:"tenant" binding:"required"`
	User     string `json:"user" binding:"required"`
	Database string `json:"database" binding:"required"`
	SqlType  string `json:"sqlType" binding:"required"`
	SqlID    string `json:"sqlID" binding:"required"`
	IsInner  bool   `json:"isInner" binding:"required"`
	SqlText  string `json:"sqlText" binding:"required"`
}

type SqlInfo struct {
	SqlMetaInfo
	ExecutionStatistics []SqlStatisticMetric `json:"executionStatistics" binding:"required"`
	LatencyStatistics   []SqlStatisticMetric `json:"latencyStatistics" binding:"required"`
	DiagnoseInfo        []SqlDiagnoseInfo    `json:"diagnoseInfo,omitempty"`
}

type SqlDetailedInfo struct {
	ExecutionTrend []response.MetricData `json:"executionTrend" binding:"required"`
	LatencyTrend   []response.MetricData `json:"latencyTrend" binding:"required"`
	DiagnoseInfo   []SqlDiagnoseInfo     `json:"diagnoseInfo,omitempty"`
	Plans          []PlanStatistic       `json:"plans" binding:"required"`
	Indexies       []IndexInfo           `json:"indexies,omitempty"`
}
