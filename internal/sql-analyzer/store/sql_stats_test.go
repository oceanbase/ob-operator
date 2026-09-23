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
	"net/http"
	"strings"
	"testing"

	"github.com/oceanbase/ob-operator/internal/sql-analyzer/model"
	httperr "github.com/oceanbase/ob-operator/pkg/errors"
)

func statsOptions(orderBy string) *QueryOptions {
	return &QueryOptions{
		SelectExpressions: []string{"sql_id", "SUM(elapsed_time_sum) / SUM(executions) / 1000 AS elapsed_time", "MAX(query_sql) AS query_sql"},
		GroupByColumns:    []string{"sql_id"}, OrderBy: orderBy, SortOrder: "DESC", Limit: 10,
	}
}

func TestSqlStatsRejectsUntrustedSortColumns(t *testing.T) {
	inputs := []string{
		"(SELECT CASE WHEN substr(user(),1,1)='d' THEN 1 ELSE CAST('a' AS INTEGER) END) -- ",
		"(SELECT CASE WHEN substr(user(),1,1)='x' THEN 1 ELSE CAST('a' AS INTEGER) END) -- ",
		"(SELECT CAST(version() AS INTEGER))", "user()", "1", "elapsed_time DESC",
		"elapsed_time;", "elapsed_time/**/", "elapsed_time --", "elapsed_time, sql_id",
		"\"elapsed_time\"", "elapsed_time\x00", " elapsed_time", "elapsed_time\n", "unknown_column",
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
					rows, err := s.QuerySqlAudits(statsOptions(input))
					e, ok := err.(httperr.ObError)
					if !ok || e.Status() != http.StatusBadRequest || rows != nil {
						t.Fatalf("want BadRequest and no data, got rows=%v, err=%v", rows, err)
					}
					if strings.Contains(err.Error(), input) {
						t.Fatalf("input echoed: %v", err)
					}
				})
			}
		})
	}
}

func TestSqlStatsPreservesSortingPaginationAndFilters(t *testing.T) {
	s := historyStore(t, true)
	if err := s.InsertBatch([][]model.SqlAudit{{
		{SqlId: "slower-sql", Executions: 2, ElapsedTimeSum: 12000, QuerySql: "SELECT 2"},
	}}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ col, direction, first string }{
		{"elapsed_time", "desc", "slower-sql"},
		{"elapsed_time", "ASC", "synthetic-sql"},
		{"sql_id", "ASC", "slower-sql"},
		{"query_sql", "DESC", "slower-sql"},
		{"elapsed_time", "DESC; SELECT 1", "synthetic-sql"},
	} {
		opts := statsOptions(tc.col)
		opts.SortOrder, opts.Limit = tc.direction, 1
		rows, err := s.QuerySqlAudits(opts)
		if err != nil || len(rows) != 1 || rows[0]["sql_id"] != tc.first {
			t.Fatalf("%+v: rows=%v, err=%v", tc, rows, err)
		}
		opts.Offset = 1
		rows, err = s.QuerySqlAudits(opts)
		if err != nil || len(rows) != 1 || rows[0]["sql_id"] == tc.first {
			t.Fatalf("pagination failed: rows=%v, err=%v", rows, err)
		}
	}
	opts := statsOptions("")
	if rows, err := s.QuerySqlAudits(opts); err != nil || len(rows) != 2 {
		t.Fatalf("omitted sort rejected: %v, %v", rows, err)
	}
	opts = statsOptions("elapsed_time")
	opts.Filters = map[string]any{"sql_id =": "synthetic-sql' OR '1'='1"}
	if rows, err := s.QuerySqlAudits(opts); err != nil || len(rows) != 0 {
		t.Fatalf("bound filter not preserved: %v, %v", rows, err)
	}
}
