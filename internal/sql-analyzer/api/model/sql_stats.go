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

// QuerySqlStatsRequest defines the request body for querying SQL statistics.
type QuerySqlStatsRequest struct {
	StartTime       int64    `json:"startTime" binding:"required" example:"1731808800"`
	EndTime         int64    `json:"endTime" binding:"required" example:"1731812400"`
	UserName        string   `json:"userName,omitempty" example:"user1"`
	DatabaseName    string   `json:"databaseName,omitempty" example:"db1"`
	FilterInnerSql  bool     `json:"filterInnerSql,omitempty"`
	SuspiciousOnly  bool     `json:"suspiciousOnly,omitempty"`
	QuerySqlKeyword string   `json:"querySqlKeyword,omitempty" example:"SELECT"`
	SortByColumn    string   `json:"sortByColumn,omitempty" example:"request_time"`
	SortOrder       string   `json:"sortOrder,omitempty" example:"DESC"`
	OutputColumns   []string `json:"outputColumns,omitempty" example:"query_sql,request_time,affected_rows"`
	PageNum         int      `json:"pageNum,omitempty"`
	PageSize        int      `json:"pageSize,omitempty"`
}

// StatisticItem holds a single metric's name and value.
type StatisticItem struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// SqlStatsItem contains the dimensions and a list of metrics for a single grouped result.
type SqlStatsItem struct {
	// Dimensions
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

	// Metrics
	Statistics []StatisticItem `json:"statistics"`
}

// SqlStatsResponse defines the overall structure of the SQL statistics API response.
type SqlStatsResponse struct {
	Items      []SqlStatsItem `json:"items"`
	TotalCount int64          `json:"totalCount"`
}
