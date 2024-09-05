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

type RequestStatisticInfo struct {
	Tenant                 string               `json:"tenant" binding:"required"`
	User                   string               `json:"user" binding:"required"`
	Database               string               `json:"database" binding:"required"`
	PlanCategoryStatistics []SqlStatisticMetric `json:"planCategoryStatistics" binding:"required"`
	ExecutionStatistics    []SqlStatisticMetric `json:"executionStatistics" binding:"required"`
	LatencyStatistics      []SqlStatisticMetric `json:"latencyStatistics" binding:"required"`
	ExecutionTrend         response.MetricData  `json:"executionTrend" binding:"required"`
	LatencyTrend           response.MetricData  `json:"latencyTrend" binding:"required"`
}
