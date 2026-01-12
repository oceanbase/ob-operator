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

type RequestStatisticsRequest struct {
	StartTime      int64  `json:"startTime,omitempty"`
	EndTime        int64  `json:"endTime,omitempty"`
	UserName       string `json:"user,omitempty"`
	DatabaseName   string `json:"database,omitempty"`
	FilterInnerSql bool   `json:"filterInnerSql,omitempty"`
}

type DailyTrend struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

type RequestStatisticsResponse struct {
	TotalExecutions  float64      `json:"totalExecutions"`
	FailedExecutions float64      `json:"failedExecutions"`
	TotalLatency     float64      `json:"totalLatency"`
	ExecutionTrend   []DailyTrend `json:"executionTrend"`
	LatencyTrend     []DailyTrend `json:"latencyTrend"`
}
