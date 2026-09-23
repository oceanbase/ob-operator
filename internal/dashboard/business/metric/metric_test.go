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
