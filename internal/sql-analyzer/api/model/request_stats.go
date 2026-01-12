package model

type RequestStatisticsRequest struct {
	StartTime      int64  `json:"startTime,omitempty"`
	EndTime        int64  `json:"endTime,omitempty"`
	UserName       string `json:"user,omitempty"`
	DatabaseName   string `json:"database,omitempty"`
	FilterInnerSql bool   `json:"filterInnerSql,omitempty"`
}

type DailyTrend struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

type RequestStatisticsResponse struct {
	TotalExecutions  float64      `json:"totalExecutions"`
	FailedExecutions float64      `json:"failedExecutions"`
	TotalLatency     float64      `json:"totalLatency"`
	ExecutionTrend   []DailyTrend `json:"executionTrend"`
	LatencyTrend     []DailyTrend `json:"latencyTrend"`
}
