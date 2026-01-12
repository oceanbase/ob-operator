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

type PlanIdentity struct {
	TenantID uint64 `json:"tenantID" binding:"required"`
	SvrIP    string `json:"svrIP" binding:"required"`
	SvrPort  int64  `json:"svrPort" binding:"required"`
	PlanID   int64  `json:"planID" binding:"required"`
}

type PlanMeta struct {
	PlanIdentity  `json:",inline"`
	PlanHash      uint64 `json:"planHash" binding:"required"`
	GeneratedTime int64  `json:"generatedTime" binding:"required"`
}

type PlanStatistic struct {
	PlanMeta `json:",inline"`
	IoCost   int64 `json:"ioCost"`
	CpuCost  int64 `json:"cpuCost"`
	Cost     int64 `json:"cost"`
	RealCost int64 `json:"realCost"`
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
	PlanMeta   `json:",inline"`
	PlanDetail *PlanOperator `json:"planDetail" binding:"required"`
}
