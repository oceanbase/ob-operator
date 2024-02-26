package external

type PrometheusMetricResult struct {
	Metric map[string]string `json:"metric"`
	Values [][]any           `json:"values"`
}

type PrometheusQueryRangeResponse struct {
	Status string                `json:"status"`
	Data   *PrometheusMetricData `json:"data"`
}

type PrometheusMetricData struct {
	ResultType string                   `json:"resultType"`
	Result     []PrometheusMetricResult `json:"result"`
}
