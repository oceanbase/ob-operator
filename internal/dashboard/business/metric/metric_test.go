/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
*/

package metric

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/external"
)

func TestReplaceQueryVariablesUsesOverlappingRateWindow(t *testing.T) {
	tests := []struct {
		name string
		step int64
		want string
	}{
		{name: "reported 28 second step", step: 28, want: "rate(metric[33s])"},
		{name: "short step keeps four scrapes", step: 10, want: "rate(metric[20s])"},
		{name: "long step overlaps one scrape", step: 60, want: "rate(metric[65s])"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceQueryVariables("rate(metric[@INTERVAL])", nil, nil, tt.step)
			if got != tt.want {
				t.Fatalf("replaceQueryVariables() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSqlOtherResponseTimeUsesOtherSqlCount(t *testing.T) {
	expr := metricExprConfig["sql_other_rt"]
	if !strings.Contains(expr, `stat_id="40018"`) {
		t.Fatalf("sql_other_rt must divide by the other SQL count (stat_id 40018): %s", expr)
	}
	if strings.Contains(expr, `stat_id="40000"`) {
		t.Fatalf("sql_other_rt must not divide by the SELECT count (stat_id 40000): %s", expr)
	}
}

func TestExtractMetricDataInterpolatesInfiniteValues(t *testing.T) {
	resp := &external.PrometheusQueryRangeResponse{
		Data: &external.PrometheusMetricData{
			Result: []external.PrometheusMetricResult{
				{
					Metric: map[string]string{"tenant": "test"},
					Values: [][]any{
						{float64(1), "1"},
						{float64(2), "+Inf"},
						{float64(3), "3"},
					},
				},
			},
		},
	}

	data := extractMetricData("sql_other_rt", resp)
	if len(data) != 1 || len(data[0].Values) != 3 {
		t.Fatalf("unexpected metric data: %#v", data)
	}
	if got := data[0].Values[1].Value; got != 2 {
		t.Fatalf("infinite sample should be interpolated between finite neighbors, got %v", got)
	}
	if _, err := json.Marshal(data); err != nil {
		t.Fatalf("metric response must remain JSON serializable: %v", err)
	}
}
