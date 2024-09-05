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

type PlanCategory string

const (
	PlanCategoryLocal       PlanCategory = "local"
	PlanCategoryRemote      PlanCategory = "remote"
	PlanCategoryDistributed PlanCategory = "distributed"
)

type PlanMeta struct {
	PlanHash      string       `json:"planHash" binding:"required"`
	Category      PlanCategory `json:"category" binding:"required"`
	MergedVersion int          `json:"mergedVersion" binding:"required"`
	GeneratedTime int64        `json:"generatedTime" binding:"required"`
}

type PlanStatistic struct {
	PlanMeta `json:",inline"`
	CpuTime  int64 `json:"cpuTime" binding:"required"`
	Cost     int64 `json:"cost" binding:"required"`
}

type PlanStatisticByServer struct {
	PlanStatistic `json:",inline"`
	Server        string `json:"server" binding:"required"`
	PlanId        int64  `json:"planId" binding:"required"`
}

type PlanOperator struct {
	Operator       string          `json:"operator" binding:"required"`
	Name           string          `json:"name,omitempty"`
	EstimatedRows  int             `json:"estimatedRows" binding:"required"`
	Cost           int64           `json:"cost" binding:"required"`
	OutputOrFilter string          `json:"outputOrFilter,omitempty"`
	ChildOperators []*PlanOperator `json:"childOperators,omitempty"`
}

type PlanDetail struct {
	PlanMeta       `json:",inline"`
	PlanStatistics []PlanStatisticByServer `json:"planStatistics" binding:"required"`
	PlanDetail     *PlanOperator           `json:"planDetail" binding:"required"`
}
