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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// This exercises the real handler and error wrapper without a Kubernetes
// connection. Validation must complete before service discovery or proxying.
func TestQuerySqlHistoryRejectsColumnsBeforeServiceDiscovery(t *testing.T) {
	r := gin.New()
	r.POST("/api/v1/sql/querySqlHistoryInfo", Wrap(QuerySqlHistoryInfo))
	for _, column := range []string{
		"elapsed_time_sum) AS e1, (SELECT CAST(user() AS INTEGER)) AS pwned -- ",
		"elapsed_time_sum) AS e1, (SELECT CAST(version() AS INTEGER)) AS pwned -- ",
		"elapsed_time;", "elapsed_time/**/", "(SELECT 1)", "unknown", "",
	} {
		body, err := json.Marshal(map[string]any{
			"namespace": "synthetic", "obtenant": "synthetic", "sqlId": "synthetic-sql", "interval": 60,
			"outputColumns": []string{"elapsed_time", column},
		})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sql/querySqlHistoryInfo", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400 before service lookup, got %d: %s", w.Code, w.Body.String())
		}
		if column != "" && strings.Contains(w.Body.String(), column) {
			t.Fatalf("input echoed: %s", w.Body.String())
		}
	}
}

func TestSqlStatsRejectsSortBeforeServiceDiscovery(t *testing.T) {
	r := gin.New()
	r.POST("/api/v1/sql/stats", Wrap(ListSqlStats))
	for _, col := range []string{
		"(SELECT CASE WHEN substr(user(),1,1)='d' THEN 1 ELSE CAST('a' AS INTEGER) END) -- ",
		"elapsed_time;", "elapsed_time/**/", "sql_id, user()", "unknown",
	} {
		body, err := json.Marshal(map[string]any{
			"namespace": "synthetic", "obtenant": "synthetic", "startTime": 1, "endTime": 7200,
			"sortColumn": col, "outputColumns": []string{"elapsed_time", col},
		})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sql/stats", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest || strings.Contains(w.Body.String(), col) {
			t.Fatalf("want sanitized 400 before discovery, got %d %s", w.Code, w.Body.String())
		}
	}
}
