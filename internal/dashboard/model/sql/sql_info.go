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
	RuleName   string `json:"ruleName"`
	Level      string `json:"level"`
	Reason     string `json:"reason"`
	Suggestion string `json:"suggestion"`
}

type SqlMetaInfo struct {
	SvrIP      string `json:"svrIp"`
	SvrPort    int64  `json:"svrPort"`
	TenantId   uint64 `json:"tenantId"`
	TenantName string `json:"tenantName"`
	UserId     int64  `json:"userId"`
	UserName   string `json:"userName"`
	DBId       uint64 `json:"dbId"`
	DBName     string `json:"dbName"`
	SqlId      string `json:"sqlId"`
	PlanId     int64  `json:"planId"`

	QuerySql          string `json:"querySql"`
	ClientIp          string `json:"clientIp"`
	Event             string `json:"event"`
	EffectiveTenantId uint64 `json:"effectiveTenantId"`
	TraceId           string `json:"traceId"`
	Sid               uint64 `json:"sid"`
	UserClientIp      string `json:"userClientIp"`
	TxId              string `json:"txId"`
	SubPlanCount      int64  `json:"subPlanCount"`
	LastFailInfo      int64  `json:"lastFailInfo"`
	CauseType         int64  `json:"causeType"`
}

type SqlInfo struct {
	SqlMetaInfo         `json:",inline"`
	ExecutionStatistics []SqlStatisticMetric `json:"executionStatistics" binding:"required"`
	LatencyStatistics   []SqlStatisticMetric `json:"latencyStatistics" binding:"required"`
	DiagnoseInfo        []SqlDiagnoseInfo    `json:"diagnoseInfo,omitempty"`
}

type SqlDetailedInfo struct {
	DiagnoseInfo []SqlDiagnoseInfo `json:"diagnoseInfo,omitempty"`
	Plans        []PlanStatistic   `json:"plans" binding:"required"`
	Indexies     []IndexInfo       `json:"indexies,omitempty"`
}

type SqlHistoryInfo struct {
	ExecutionTrend []response.MetricData `json:"executionTrend" binding:"required"`
	LatencyTrend   []response.MetricData `json:"latencyTrend" binding:"required"`
}

type SqlStatsList struct {
	Items      []SqlInfo `json:"items"`
	TotalCount int64     `json:"totalCount"`
}
