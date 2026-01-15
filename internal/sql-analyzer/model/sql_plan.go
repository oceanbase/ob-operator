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

// PlanIdentifier holds the identifiers for a plan.
type SqlPlanIdentifier struct {
	TenantID uint64 `json:"tenantID" binding:"required"`
	SvrIP    string `json:"svrIP" binding:"required"`
	SvrPort  int64  `json:"svrPort" binding:"required"`
	PlanID   int64  `json:"planID" binding:"required"`
}

type SqlPlan struct {
	TenantID          uint64 `db:"TENANT_ID"`
	SvrIP             string `db:"SVR_IP"`
	SvrPort           int64  `db:"SVR_PORT"`
	PlanID            int64  `db:"PLAN_ID"`
	SqlID             string `db:"SQL_ID"`
	DbID              int64  `db:"DB_ID"`
	PlanHash          uint64 `db:"PLAN_HASH"`
	GmtCreate         string `db:"GMT_CREATE"`
	Operator          string `db:"OPERATOR"`
	ObjectNode        string `db:"OBJECT_NODE"`
	ObjectID          int64  `db:"OBJECT_ID"`
	ObjectOwner       string `db:"OBJECT_OWNER"`
	ObjectName        string `db:"OBJECT_NAME"`
	ObjectAlias       string `db:"OBJECT_ALIAS"`
	ObjectType        string `db:"OBJECT_TYPE"`
	Optimizer         string `db:"OPTIMIZER"`
	ID                int64  `db:"ID"`
	ParentID          int64  `db:"PARENT_ID"`
	Depth             int64  `db:"DEPTH"`
	Position          int64  `db:"POSITION"`
	Cost              int64  `db:"COST"`
	RealCost          int64  `db:"REAL_COST"`
	Cardinality       int64  `db:"CARDINALITY"`
	RealCardinality   int64  `db:"REAL_CARDINALITY"`
	IoCost            int64  `db:"IO_COST"`
	CpuCost           int64  `db:"CPU_COST"`
	Bytes             int64  `db:"BYTES"`
	Rowset            int64  `db:"ROWSET"`
	OtherTag          string `db:"OTHER_TAG"`
	PartitionStart    string `db:"PARTITION_START"`
	Other             string `db:"OTHER"`
	AccessPredicates  string `db:"ACCESS_PREDICATES"`
	FilterPredicates  string `db:"FILTER_PREDICATES"`
	StartupPredicates string `db:"STARTUP_PREDICATES"`
	Projection        string `db:"PROJECTION"`
	SpecialPredicates string `db:"SPECIAL_PREDICATES"`
	QblockName        string `db:"QBLOCK_NAME"`
	Remarks           string `db:"REMARKS"`
	OtherXML          string `db:"OTHER_XML"`
}

type PlanStatistic struct {
	TenantID      uint64 `db:"TENANT_ID"`
	SvrIP         string `db:"SVR_IP"`
	SvrPort       int64  `db:"SVR_PORT"`
	PlanID        int64  `db:"PLAN_ID"`
	PlanHash      uint64 `db:"PLAN_HASH"`
	GeneratedTime string `db:"GMT_CREATE"`
	IoCost        int64  `db:"IO_COST"`
	CpuCost       int64  `db:"CPU_COST"`
	Cost          int64  `db:"COST"`
	RealCost      int64  `db:"REAL_COST"`
}

type TableInfo struct {
	DatabaseName string `db:"OBJECT_OWNER"`
	TableName    string `db:"OBJECT_NAME"`
	TableID      int64  `db:"OBJECT_ID"`
}
