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

type BaseSqlRequestParam struct {
	Namespace       string `json:"namespace" binding:"required"`
	OBTenant        string `json:"obtenant" binding:"required"`
	User            string `json:"user,omitempty"`
	Database        string `json:"database,omitempty"`
	IncludeInnerSql bool   `json:"includeInnerSql,omitempty"`
	StartTime       int64  `json:"startTime,omitempty"`
	EndTime         int64  `json:"endTime,omitempty"`
}

type Pagination struct {
	SortByColumn string `json:"sortColumn"`
	SortOrder    string `json:"sortOrder"`
	PageNum      int    `json:"pageNum,omitempty"`
	PageSize     int    `json:"pageSize,omitempty"`
}

type SqlFilter struct {
	BaseSqlRequestParam `json:",inline"`
	Pagination          `json:",inline"`
	OutputColumns       []string `json:"outputColumns"`
	Keyword             string   `json:"keyword,omitempty"`
	SuspiciousOnly      bool     `json:"suspiciousOnly,omitempty"`
}

type SqlRequestStatisticParam struct {
	BaseSqlRequestParam `json:",inline"`
	Pagination          `json:",inline"`
	StatisticScopes     []string `json:"statisticScopes" binding:"required"`
}

type SqlDetailParam struct {
	BaseSqlRequestParam `json:",inline"`
	SqlId               string `json:"sqlId" binding:"required"`
}

type SqlHistoryParam struct {
	BaseSqlRequestParam `json:",inline"`
	Interval            int      `json:"interval" binding:"required"`
	SqlId               string   `json:"sqlId" binding:"required"`
	LatencyColumns      []string `json:"outputColumns"`
}
