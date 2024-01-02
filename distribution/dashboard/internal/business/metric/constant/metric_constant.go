package constant

const (
	MetricConfigFile     = "assets/metric.yaml"
	MetricConfigFileEnUS = "assets/metric_en_US.yaml"
	MetricConfigFileZhCN = "assets/metric_zh_CN.yaml"
	MetricExprConfigFile = "assets/metric_expr.yaml"
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
)

const (
	DefaultMetricQueryTimeout = 5
)
