package model

type SqlHistoryRequest struct {
	StartTime      int64    `json:"startTime,omitempty"`
	EndTime        int64    `json:"endTime,omitempty"`
	SqlId          string   `json:"sqlId" binding:"required"`
	Interval       int      `json:"interval" binding:"required"`
	LatencyColumns []string `json:"latencyColumns"`
}

type SqlHistoryResponse struct {
	ExecutionTrend []PlanTypeTrend    `json:"executionTrend"`
	LatencyTrend   []LatencyTrendItem `json:"latencyTrend"`
}
