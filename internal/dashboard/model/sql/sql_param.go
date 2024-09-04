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
	Namespace   string `json:"namespace" binding:"required"`
	OBClusterCR string `json:"obcluster" binding:"required"`
	OBTenantCR  string `json:"obtenant,omitempty"`
	OBUser      string `json:"user,omitempty"`
	OBDatabase  string `json:"database,omitempty"`
	StartTime   int64  `json:"startTime,omitempty"`
	EndTime     int64  `json:"endTime,omitempty"`
}

type SqlFilter struct {
	BaseSqlRequestParam `json:",inline"`
	Keyword             string `json:"keyword,omitempty"`
	IncludeInnerSql     bool   `json:"includeInnerSql,omitempty"`
}

type SqlRequestStatisticParam struct {
	BaseSqlRequestParam `json:",inline"`
	StatisticScopes     []string `json:"statisticScopes" binding:"required"`
}

type SqlDetailParam struct {
	BaseSqlRequestParam `json:",inline"`
	Interval            int    `json:"interval" binding:"required"`
	SqlId               string `json:"sqlId" binding:"required"`
}

type PlanDetailParam struct {
	BaseSqlRequestParam `json:",inline"`
	PlanHash            string `json:"planHash" binding:"required"`
}
