package model

type SqlDetailRequest struct {
	StartTime int64  `json:"startTime,omitempty"`
	EndTime   int64  `json:"endTime,omitempty"`
	SqlId     string `json:"sqlId" binding:"required"`
}

type PlanTypeTrend struct {
	Time        int64   `json:"time"`
	Local       float64 `json:"local"`
	Remote      float64 `json:"remote"`
	Distributed float64 `json:"distributed"`
}

type LatencyTrendItem struct {
	Time  int64              `json:"time"`
	Value map[string]float64 `json:"value"`
}

type PlanStats struct {
	TenantID      uint64 `json:"tenantId"`
	SvrIP         string `json:"svrIp"`
	SvrPort       int64  `json:"svrPort"`
	PlanID        int64  `json:"planId"`
	PlanHash      uint64 `json:"planHash"`
	GeneratedTime int64  `json:"generatedTime"`
	IoCost        int64  `json:"ioCost"`
	CpuCost       int64  `json:"cpuCost"`
	Cost          int64  `json:"cost"`
	RealCost      int64  `json:"realCost"`
}

type TableInfo struct {
	DatabaseName string `json:"databaseName"`
	TableName    string `json:"tableName"`
	TableID      int64  `json:"tableId"`
}

type IndexInfo struct {
	TableName  string   `json:"tableName" binding:"required"`
	IndexType  string   `json:"indexType" binding:"required"`
	Uniqueness string   `json:"uniqueness" binding:"required"`
	IndexName  string   `json:"indexName" binding:"required"`
	Columns    []string `json:"columns" binding:"required"`
	Status     string   `json:"status" binding:"required"`
}

type SqlDetailResponse struct {
	QuerySql     string            `json:"querySql,omitempty"`
	Plans        []PlanStats       `json:"plans"`
	Tables       []TableInfo       `json:"tables"`
	Indexes      []IndexInfo       `json:"indexes"`
	DiagnoseInfo []SqlDiagnoseInfo `json:"diagnoseInfo,omitempty"`
}

type SqlDiagnoseInfo struct {
	RuleName   string `json:"ruleName"`
	Level      string `json:"level"`
	Suggestion string `json:"suggestion"`
	Reason     string `json:"reason"`
}
