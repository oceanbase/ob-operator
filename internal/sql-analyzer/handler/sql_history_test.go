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

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/oceanbase/ob-operator/internal/sql-analyzer/config"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
)

func TestSqlHistoryHTTPRejectsColumnsAndHidesEngineErrors(t *testing.T) {
	dir := t.TempDir()
	if err := store.InitGlobalStores(context.Background(), &config.Config{DataPath: dir, DuckDBMaxOpenConns: 4, DuckDBThreads: 1}, logrus.New()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(store.CloseGlobalStores)
	r := gin.New()
	r.POST("/api/v1/tenants/:tenant_name/sql-history", Wrap(GetSqlHistoryInfo))
	query := func(columns []string) *httptest.ResponseRecorder {
		t.Helper()
		body, err := json.Marshal(map[string]any{"sqlId": "synthetic-sql", "interval": 60, "latencyColumns": columns})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants/synthetic/sql-history", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}
	for _, column := range []string{
		"elapsed_time_sum) AS e1, (SELECT CAST(user() AS INTEGER)) AS pwned -- ",
		"elapsed_time_sum) AS e1, (SELECT CAST(version() AS INTEGER)) AS pwned -- ",
		"elapsed_time;", "(SELECT 1)", "unknown",
	} {
		w := query([]string{"elapsed_time", column})
		if w.Code != http.StatusBadRequest || strings.Contains(w.Body.String(), column) {
			t.Fatalf("unsafe column should receive sanitized 400: %d %s", w.Code, w.Body.String())
		}
	}
	if w := query([]string{"elapsed_time"}); w.Code != http.StatusOK {
		t.Fatalf("valid empty-store query failed: %d %s", w.Code, w.Body.String())
	}
	if err := os.WriteFile(filepath.Join(dir, "sql_audit", "broken.parquet"), []byte("synthetic invalid parquet"), 0600); err != nil {
		t.Fatal(err)
	}
	w := query([]string{"elapsed_time"})
	if w.Code != http.StatusInternalServerError || strings.Contains(w.Body.String(), dir) || strings.Contains(w.Body.String(), "parquet") {
		t.Fatalf("engine details leaked: %d %s", w.Code, w.Body.String())
	}
}

func TestSqlStatsHTTPRejectsSortAndHidesEngineErrors(t *testing.T) {
	dir := t.TempDir()
	if err := store.InitGlobalStores(context.Background(), &config.Config{DataPath: dir, DuckDBMaxOpenConns: 4, DuckDBThreads: 1}, logrus.New()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(store.CloseGlobalStores)
	r := gin.New()
	r.POST("/api/v1/tenants/:tenant_name/sql-stats", Wrap(QuerySqlStats))
	query := func(col string) *httptest.ResponseRecorder {
		t.Helper()
		body, err := json.Marshal(map[string]any{"sortByColumn": col, "outputColumns": []string{"elapsed_time"}, "startTime": 1, "endTime": 7200})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants/synthetic/sql-stats", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}
	for _, col := range []string{
		"(SELECT CASE WHEN substr(user(),1,1)='d' THEN 1 ELSE CAST('a' AS INTEGER) END) -- ",
		"elapsed_time;", "sql_id, user()", "unknown",
	} {
		w := query(col)
		if w.Code != http.StatusBadRequest || strings.Contains(w.Body.String(), col) {
			t.Fatalf("want sanitized 400 from empty store, got %d %s", w.Code, w.Body.String())
		}
	}
	if w := query("elapsed_time"); w.Code != http.StatusOK {
		t.Fatalf("normal empty query failed: %d %s", w.Code, w.Body.String())
	}
	if err := os.WriteFile(filepath.Join(dir, "sql_audit", "broken.parquet"), []byte("synthetic invalid parquet"), 0600); err != nil {
		t.Fatal(err)
	}
	w := query("elapsed_time")
	if w.Code != http.StatusInternalServerError || strings.Contains(w.Body.String(), dir) || strings.Contains(w.Body.String(), "parquet") {
		t.Fatalf("engine details leaked: %d %s", w.Code, w.Body.String())
	}
}
