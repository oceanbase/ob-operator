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
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	apimodel "github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	sqlconst "github.com/oceanbase/ob-operator/internal/sql-analyzer/const/sql"
)

type schemaAuditRow struct {
	id         string
	query      any
	executions int64
}

func newSchemaTestStore(t *testing.T) *SqlAuditStore {
	t.Helper()
	l := logrus.New()
	l.SetOutput(io.Discard)
	s, err := NewSqlAuditStore(context.Background(), t.TempDir(), 4, 1, l)
	require.NoError(t, err)
	t.Cleanup(s.Close)
	_, err = s.db.Exec(strings.Replace(fmt.Sprintf(sqlconst.CreateSqlAuditTempTableTemplate, "fixture"), "CREATE TEMP TABLE", "CREATE TABLE", 1))
	require.NoError(t, err)
	return s
}

// Create real Parquet files with the current schema or a historical schema
// missing text columns. Each fixture lives only in the test's temporary directory.
func writeSchemaParquet(t *testing.T, s *SqlAuditStore, name string, legacy bool, rows ...schemaAuditRow) string {
	t.Helper()
	_, err := s.db.Exec("DELETE FROM fixture")
	require.NoError(t, err)
	for _, row := range rows {
		_, err = s.db.Exec(`INSERT INTO fixture
			(sql_id, query_sql, client_ip, executions, elapsed_time_sum,
			 fail_count_sum, min_request_time, max_request_time,
			 plan_type_local_count, plan_type_remote_count, plan_type_distributed_count,
			 user_name, db_name, inner_sql_count)
			VALUES (?, ?, '127.0.0.1', ?, ?, 0, 100000000, 100000000, ?, 0, 0, 'reader', 'test', 0)`,
			row.id, row.query, row.executions, row.executions*10, row.executions)
		require.NoError(t, err)
	}
	projection := "*"
	if legacy {
		projection = "* EXCLUDE (query_sql, client_ip)"
	}
	path := filepath.Join(s.path, name)
	_, err = s.db.Exec(fmt.Sprintf("COPY (SELECT %s FROM fixture) TO '%s' (FORMAT PARQUET)", projection, strings.ReplaceAll(path, "'", "''")))
	require.NoError(t, err)
	return path
}

func TestSqlAuditQueriesMixedParquetSchemas(t *testing.T) {
	for _, legacyName := range []string{
		"compacted-2026-09-22-09-26-31.parquet",
		"2026-09-21-09-26-31-legacy.parquet",
	} {
		t.Run(legacyName, func(t *testing.T) {
			s := newSchemaTestStore(t)
			writeSchemaParquet(t, s, legacyName, true,
				schemaAuditRow{"shared", nil, 2}, schemaAuditRow{"legacy_only", nil, 4})
			writeSchemaParquet(t, s, "2026-09-23-00-29-31-current.parquet", false,
				schemaAuditRow{"shared", "SELECT 42", 3})

			t.Run("list and pagination", func(t *testing.T) {
				opts := &QueryOptions{
					SelectExpressions: []string{"sql_id", "MAX(query_sql) AS query_sql", "CAST(SUM(executions) AS BIGINT) AS executions"},
					GroupByColumns:    []string{"sql_id"}, OrderBy: "sql_id", Limit: 1,
				}
				count, err := s.CountSqlAudits(opts)
				require.NoError(t, err)
				require.EqualValues(t, 2, count)
				result, err := s.QuerySqlAudits(opts)
				require.NoError(t, err)
				require.Len(t, result, 1)
				require.Equal(t, "legacy_only", result[0]["sql_id"])
				require.Nil(t, result[0]["query_sql"])
				require.EqualValues(t, 4, result[0]["executions"])

				opts.Offset = 1
				result, err = s.QuerySqlAudits(opts)
				require.NoError(t, err)
				require.Len(t, result, 1)
				require.Equal(t, "shared", result[0]["sql_id"])
				require.Equal(t, "SELECT 42", result[0]["query_sql"])
				require.EqualValues(t, 5, result[0]["executions"])
			})

			t.Run("keyword filter", func(t *testing.T) {
				opts := &QueryOptions{
					SelectExpressions: []string{"query_sql"}, Filters: map[string]any{"query_sql ILIKE": "%42%"},
					Limit: 10,
				}
				count, err := s.CountSqlAudits(opts)
				require.NoError(t, err)
				require.EqualValues(t, 1, count)
				result, err := s.QuerySqlAudits(opts)
				require.NoError(t, err)
				require.Len(t, result, 1)
				require.Equal(t, "SELECT 42", result[0]["query_sql"])
			})

			t.Run("request statistics", func(t *testing.T) {
				result, err := s.QueryRequestStatistics(apimodel.RequestStatisticsRequest{
					StartTime: 99000, EndTime: 101000, UserName: "reader", DatabaseName: "test", FilterInnerSql: true,
				})
				require.NoError(t, err)
				require.Equal(t, float64(9), result.TotalExecutions)
				require.Zero(t, result.FailedExecutions)
				require.Equal(t, float64(10), result.TotalLatency)
				require.Len(t, result.ExecutionTrend, 1)
				require.Equal(t, float64(9), result.ExecutionTrend[0].Value)
			})

			t.Run("history", func(t *testing.T) {
				result, err := s.QuerySqlHistoryInfo(apimodel.SqlHistoryRequest{
					StartTime: 99, EndTime: 101, SqlId: "shared", Interval: 60, LatencyColumns: []string{"elapsed_time"},
				})
				require.NoError(t, err)
				require.Len(t, result.ExecutionTrend, 1)
				require.Equal(t, float64(5), result.ExecutionTrend[0].Local)
				require.Len(t, result.LatencyTrend, 1)
				require.Equal(t, 0.01, result.LatencyTrend[0].Value["elapsed_time"]) // microseconds to milliseconds
			})

			t.Run("detail", func(t *testing.T) {
				plans, err := NewPlanStore(context.Background(), t.TempDir(), 2, 1, s.Logger)
				require.NoError(t, err)
				t.Cleanup(plans.Close)
				require.NoError(t, plans.InitSqlPlanTable())
				for _, tc := range []struct{ id, want string }{
					{"shared", "SELECT 42"}, {"legacy_only", ""}, {"absent", ""},
				} {
					t.Run(tc.id, func(t *testing.T) {
						result, err := s.QuerySqlDetailInfo(plans, apimodel.SqlDetailRequest{
							StartTime: 99, EndTime: 101, SqlId: tc.id,
						})
						require.NoError(t, err)
						require.Equal(t, tc.want, result.QuerySql)
						require.Empty(t, result.Plans)
					})
				}
			})
		})
	}
}

func TestCompactMixedParquetSchemas(t *testing.T) {
	for _, legacyFirst := range []bool{true, false} {
		t.Run(fmt.Sprintf("legacyFirst=%t", legacyFirst), func(t *testing.T) {
			s := newSchemaTestStore(t)
			first := writeSchemaParquet(t, s, "2026-09-21-00-00-00-first.parquet", legacyFirst,
				schemaAuditRow{"shared", "SELECT 42", 2})
			second := writeSchemaParquet(t, s, "2026-09-22-00-00-00-second.parquet", !legacyFirst,
				schemaAuditRow{"shared", "SELECT 42", 3})
			latest := writeSchemaParquet(t, s, "2026-09-23-00-00-00-latest.parquet", false,
				schemaAuditRow{"latest", "SELECT 43", 4})
			require.NoError(t, s.Compact())
			require.NoFileExists(t, first)
			require.NoFileExists(t, second)
			require.FileExists(t, latest)
			compacted := filepath.Join(s.path, "compacted-2026-09-22-00-00-00.parquet")
			require.FileExists(t, compacted)
			var count, executions, textCount int64
			var query string
			err := s.db.QueryRow(fmt.Sprintf("SELECT COUNT(*), SUM(executions), COUNT(query_sql), MAX(query_sql) FROM read_parquet('%s')", compacted)).
				Scan(&count, &executions, &textCount, &query)
			require.NoError(t, err)
			require.EqualValues(t, 2, count)
			require.EqualValues(t, 5, executions)
			require.EqualValues(t, 1, textCount)
			require.Equal(t, "SELECT 42", query)

			result, err := s.QuerySqlAudits(&QueryOptions{
				SelectExpressions: []string{"CAST(SUM(executions) AS BIGINT) AS executions"}, Limit: 1,
			})
			require.NoError(t, err)
			require.Len(t, result, 1)
			require.EqualValues(t, 9, result[0]["executions"])
			require.NoError(t, s.Compact()) // The latest file is still retained.
			require.FileExists(t, latest)
		})
	}
}

func TestSqlAuditCurrentSchemaControl(t *testing.T) {
	s := newSchemaTestStore(t)
	opts := &QueryOptions{SelectExpressions: []string{"query_sql"}, Limit: 10}
	count, err := s.CountSqlAudits(opts)
	require.NoError(t, err)
	require.Zero(t, count)
	result, err := s.QuerySqlAudits(opts)
	require.NoError(t, err)
	require.Empty(t, result)

	writeSchemaParquet(t, s, "2026-09-23-00-00-00-current.parquet", false,
		schemaAuditRow{"current", "SELECT 42", 1})
	result, err = s.QuerySqlAudits(opts)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "SELECT 42", result[0]["query_sql"])
}
