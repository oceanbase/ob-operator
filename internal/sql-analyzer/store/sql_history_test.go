/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package store

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	apimodel "github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/common"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/model"
	httperr "github.com/oceanbase/ob-operator/pkg/errors"
)

func historyStore(t *testing.T, populated bool) *SqlAuditStore {
	t.Helper()
	s, err := NewSqlAuditStore(context.Background(), t.TempDir(), 4, 1, logrus.New())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	if populated {
		rows := []model.SqlAudit{
			{SqlId: "synthetic-sql", MaxRequestTime: 3600000000, Executions: 2, ElapsedTimeSum: 8000, ExecuteTimeSum: 4000, QueueTimeSum: 2000, Event0WaitTimeSum: 1000, PlanTypeLocalCount: 2},
			{SqlId: "synthetic-sql", MaxRequestTime: 3610000000, Executions: 3, ElapsedTimeSum: 2000, ExecuteTimeSum: 1000, QueueTimeSum: 3000, Event0WaitTimeSum: 2000, PlanTypeRemoteCount: 3},
		}
		if err := s.InsertBatch([][]model.SqlAudit{rows}); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

func historyRequest(columns []string) apimodel.SqlHistoryRequest {
	return apimodel.SqlHistoryRequest{SqlId: "synthetic-sql", StartTime: 1, EndTime: 7200, Interval: 3600, LatencyColumns: columns}
}

func TestSqlHistoryRejectsUntrustedColumns(t *testing.T) {
	inputs := []string{
		"elapsed_time_sum) AS e1, (SELECT CAST(user() AS INTEGER)) AS pwned -- ",
		"elapsed_time_sum) AS e1, (SELECT CAST(version() AS INTEGER)) AS pwned -- ",
		"elapsed_time;", "elapsed_time/**/", "elapsed_time --", "elapsed_time, execute_time",
		"SUM(elapsed_time_sum)", "(SELECT 1)", "\"elapsed_time\"", "elapsed_time AS other",
		"elapsed_time\x00", " elapsed_time", "elapsed_time\n", "ELAPSED_TIME", "unknown_metric", "",
	}
	for _, populated := range []bool{false, true} {
		name := "empty"
		if populated {
			name = "populated"
		}
		t.Run(name, func(t *testing.T) {
			s := historyStore(t, populated)
			for _, input := range inputs {
				t.Run(input, func(t *testing.T) {
					resp, err := s.QuerySqlHistoryInfo(historyRequest([]string{"elapsed_time", input}))
					oberr, ok := err.(httperr.ObError)
					if !ok || oberr.Status() != http.StatusBadRequest || resp != nil {
						t.Fatalf("want BadRequest and no data, got response=%+v, err=%v", resp, err)
					}
					if strings.Contains(err.Error(), "Conversion Error") || (input != "" && strings.Contains(err.Error(), input)) {
						t.Fatalf("error discloses input or database details: %v", err)
					}
				})
			}
		})
	}
}

func TestSqlHistoryPreservesMetricsAndFilters(t *testing.T) {
	s := historyStore(t, true)
	columns := make([]string, 0, len(common.TimeMetrics))
	for col := range common.TimeMetrics {
		columns = append(columns, col)
	}
	sort.Strings(columns)
	resp, err := s.QuerySqlHistoryInfo(historyRequest(columns))
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.ExecutionTrend) != 1 || resp.ExecutionTrend[0].Local != 2 || resp.ExecutionTrend[0].Remote != 3 || len(resp.LatencyTrend) != 1 {
		t.Fatalf("unexpected history: %+v", resp)
	}
	values := resp.LatencyTrend[0].Value
	if len(values) != len(columns) {
		t.Fatalf("missing metrics: %v", values)
	}
	for key, want := range map[string]float64{"elapsed_time": 2, "execute_time": 1, "queue_time": 1, "event_0_wait_time_sum": 3} {
		if values[key] != want {
			t.Errorf("%s = %v, want %v", key, values[key], want)
		}
	}
	for _, columns := range [][]string{nil, {}} {
		resp, err := s.QuerySqlHistoryInfo(historyRequest(columns))
		if err != nil || len(resp.ExecutionTrend) != 1 || len(resp.LatencyTrend) != 0 {
			t.Fatalf("empty selection should preserve execution trend: %+v, %v", resp, err)
		}
	}
	for _, req := range []apimodel.SqlHistoryRequest{
		{SqlId: "synthetic-sql' OR '1'='1", StartTime: 1, EndTime: 7200, Interval: 3600, LatencyColumns: columns},
		{SqlId: "synthetic-sql", StartTime: 7201, EndTime: 10800, Interval: 3600, LatencyColumns: columns},
	} {
		resp, err := s.QuerySqlHistoryInfo(req)
		if err != nil || len(resp.ExecutionTrend) != 0 || len(resp.LatencyTrend) != 0 {
			t.Fatalf("filters should exclude fixture: %+v, %v", resp, err)
		}
	}
}
