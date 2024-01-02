package response

import "github.com/oceanbase/oceanbase-dashboard/internal/model/common"

type MetricClass struct {
	Name         string        `json:"name" yaml:"name"`
	Description  string        `json:"description" yaml:"description"`
	MetricGroups []MetricGroup `json:"metricGroups" yaml:"metricGroups"`
}

type MetricGroup struct {
	Name        string       `json:"name" yaml:"name"`
	Description string       `json:"description" yaml:"description"`
	Metrics     []MetricMeta `json:"metrics" yaml:"metrics"`
}

type MetricMeta struct {
	Name        string `json:"name" yaml:"name"`
	Unit        string `json:"unit" yaml:"unit"`
	Description string `json:"description" yaml:"description"`
	Key         string `json:"key" yaml:"key"`
}

type Metric struct {
	Name   string          `json:"name" yaml:"name"`
	Labels []common.KVPair `json:"labels" yaml:"labels"`
}

type MetricValue struct {
	Value     float64 `json:"value" yaml:"value"`
	Timestamp float64 `json:"timestamp" yaml:"timestamp"`
}

type MetricData struct {
	Metric Metric        `json:"metric" yaml:"metric"`
	Values []MetricValue `json:"values" yaml:"values"`
}
