/*
Copyright (c) 2025 OceanBase
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

type SqlHistoryRequest struct {
	StartTime      int64    `json:"startTime,omitempty"`
	EndTime        int64    `json:"endTime,omitempty"`
	SqlId          string   `json:"sqlId" binding:"required"`
	Interval       int      `json:"interval" binding:"required"`
	LatencyColumns []string `json:"latencyColumns"`
}

type SqlHistoryResponse struct {
	ExecutionTrend []PlanTypeTrend    `json:"executionTrend"`
	LatencyTrend   []LatencyTrendItem `json:"latencyTrend"`
}
