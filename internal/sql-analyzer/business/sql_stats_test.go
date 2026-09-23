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

package business

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	apimodel "github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/common"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/config"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
	httperr "github.com/oceanbase/ob-operator/pkg/errors"
)

func TestSqlStatsValidatesSortBeforeCount(t *testing.T) {
	// A nil store proves that invalid sort keys never reach even the count query.
	s := NewSqlStatsService(nil, nil, nil)
	for _, col := range []string{
		"(SELECT CASE WHEN substr(user(),1,1)='d' THEN 1 ELSE CAST('a' AS INTEGER) END) -- ",
		"elapsed_time;", "sql_id, user()", "unknown",
	} {
		resp, err := s.QuerySqlStats(&apimodel.QuerySqlStatsRequest{SortByColumn: col, OutputColumns: []string{col}})
		e, ok := err.(httperr.ObError)
		if !ok || e.Status() != http.StatusBadRequest || resp != nil {
			t.Fatalf("sort %q accepted before counting: %+v, %v", col, resp, err)
		}
	}
}

func TestSqlStatsCatalogSortingRemainsSupported(t *testing.T) {
	l := logrus.New()
	l.SetOutput(io.Discard)
	db, err := store.NewSqlAuditStore(context.Background(), t.TempDir(), 4, 1, l)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	if err := db.InsertBatch([][]model.SqlAudit{{
		{SqlId: "synthetic-sql", TenantName: "synthetic", UserName: "tester", DBName: "synthetic", Executions: 2, ElapsedTimeSum: 8000},
		{SqlId: "synthetic-sql", TenantName: "synthetic", UserName: "tester", DBName: "synthetic", Executions: 3, ElapsedTimeSum: 2000},
	}}); err != nil {
		t.Fatal(err)
	}
	s := NewSqlStatsService(db, &config.Config{}, l)
	columns := []string{""}
	for col := range common.ColumnAggregations {
		columns = append(columns, col)
	}
	for col := range common.Dimensions {
		columns = append(columns, col)
	}
	for col := range common.FixedDimensions {
		columns = append(columns, col)
	}
	for _, col := range columns {
		t.Run(col, func(t *testing.T) {
			output := []string{col}
			if _, fixed := common.FixedDimensions[col]; fixed || col == "" {
				output = nil
			}
			resp, err := s.QuerySqlStats(&apimodel.QuerySqlStatsRequest{SortByColumn: col, SortOrder: "desc", OutputColumns: output, PageNum: 1, PageSize: 10})
			// These pre-existing catalog entries have no backing parquet columns.
			// Preserve their query error; this fix must not silently reinterpret them.
			switch col {
			case "cpu_time", "cpu_time_sum", "cpu_time_min", "cpu_time_max", "sub_plan_count", "last_fail_info", "cause_type":
				if err == nil || !strings.Contains(err.Error(), "Binder Error") {
					t.Fatalf("expected existing schema error for %s: %v", col, err)
				}
				return
			}
			if err != nil || resp == nil || resp.TotalCount != 1 || len(resp.Items) != 1 || resp.Items[0].SqlId != "synthetic-sql" {
				t.Fatalf("normal projected or fixed sort failed: %+v, %v", resp, err)
			}
			if col == "elapsed_time" && (len(resp.Items[0].Statistics) != 1 || resp.Items[0].Statistics[0].Value != 2) {
				t.Fatalf("weighted average changed: %+v", resp.Items[0])
			}
		})
	}
}
