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

package response

import "github.com/oceanbase/ob-operator/internal/dashboard/model/common"

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
