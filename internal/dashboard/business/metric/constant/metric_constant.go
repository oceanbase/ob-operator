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

package constant

const (
	MetricConfigFile     = "internal/assets/dashboard/metric.yaml"
	MetricConfigFileEnUS = "internal/assets/dashboard/metric_en_US.yaml"
	MetricConfigFileZhCN = "internal/assets/dashboard/metric_zh_CN.yaml"
	MetricExprConfigFile = "internal/assets/dashboard/metric_expr.yaml"
)

const (
	KeyInterval    = "@INTERVAL"
	KeyLabels      = "@LABELS"
	KeyGroupLabels = "@GBLABELS"
)

const (
	PrometheusAddress   = "http://127.0.0.1:9090"
	MetricRangeQueryUrl = "/api/v1/query_range"
	MetricQueryUrl      = "/api/v1/query"
)

const (
	ScopeCluster         = "OBCLUSTER"
	ScopeClusterOverview = "OBCLUSTER_OVERVIEW"
	ScopeTenant          = "OBTENANT"
	ScopeOBProxy         = "OBPROXY"
)

const (
	DefaultMetricQueryTimeout = 5
)
