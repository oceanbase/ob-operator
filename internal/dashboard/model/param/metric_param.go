package param

import "github.com/oceanbase/ob-operator/internal/dashboard/model/common"

type QueryRange struct {
	StartTimestamp float64 `json:"startTimestamp"`
	EndTimestamp   float64 `json:"endTimestamp"`
	Step           int64   `json:"step"`
}

type MetricQuery struct {
	Metrics     []string        `json:"metrics"`
	Labels      []common.KVPair `json:"labels"`
	GroupLabels []string        `json:"groupLabels"`
	QueryRange  QueryRange      `json:"queryRange"`
}
