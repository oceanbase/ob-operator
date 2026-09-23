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

package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apimodel "github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	httperr "github.com/oceanbase/ob-operator/pkg/errors"
)

func TestQuerySqlHistoryDoesNotExposeUpstreamErrors(t *testing.T) {
	for _, tc := range []struct {
		name         string
		status, want int
	}{
		{"bad_request", 400, 400}, {"internal_error", 500, 500}, {"failed_envelope", 200, 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				json.NewEncoder(w).Encode(map[string]any{"successful": false, "message": "private-sql-trace: SELECT user(), v1.4.3, /private/data/file"})
			}))
			defer srv.Close()
			resp, err := NewClient(srv.URL).QuerySqlHistory("synthetic", apimodel.SqlHistoryRequest{SqlId: "synthetic-sql", Interval: 60})
			e, ok := err.(httperr.ObError)
			if !ok || e.Status() != tc.want || resp != nil {
				t.Fatalf("want status %d, got %+v, %v", tc.want, resp, err)
			}
			if strings.Contains(err.Error(), "private") || strings.Contains(err.Error(), "SELECT") || strings.Contains(err.Error(), "v1.4.3") {
				t.Fatalf("upstream details exposed: %v", err)
			}
		})
	}
}

func TestQuerySqlHistoryPreservesRequestAndResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/tenants/synthetic/sql-history" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var req apimodel.SqlHistoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
		}
		if len(req.LatencyColumns) != 1 || req.LatencyColumns[0] != "elapsed_time" || req.SqlId != "synthetic-sql" || req.Interval != 60 {
			t.Errorf("request not preserved: %+v", req)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"successful": true, "data": apimodel.SqlHistoryResponse{
			LatencyTrend: []apimodel.LatencyTrendItem{{Time: 3600, Value: map[string]float64{"elapsed_time": 2}}},
		}})
	}))
	defer srv.Close()
	resp, err := NewClient(srv.URL).QuerySqlHistory("synthetic", apimodel.SqlHistoryRequest{
		SqlId: "synthetic-sql", Interval: 60, LatencyColumns: []string{"elapsed_time"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil || len(resp.LatencyTrend) != 1 || resp.LatencyTrend[0].Value["elapsed_time"] != 2 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestQuerySqlStatsDoesNotExposeUpstreamErrors(t *testing.T) {
	for _, tc := range []struct {
		name         string
		status, want int
	}{
		{"bad_request", 400, 400}, {"failed_envelope", 200, 500}, {"legacy_unavailable_fallback", 503, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				json.NewEncoder(w).Encode(map[string]any{"successful": false, "message": "private-sql-trace: SELECT user(), v1.4.3, /private/data/file"})
			}))
			defer srv.Close()
			resp, err := NewClient(srv.URL).QuerySqlStats("synthetic", apimodel.QuerySqlStatsRequest{SortByColumn: "elapsed_time"})
			if tc.want == 0 {
				if err != nil || resp == nil || len(resp.Items) != 0 || resp.TotalCount != 0 {
					t.Fatalf("existing unavailable fallback changed: %+v, %v", resp, err)
				}
				return
			}
			e, ok := err.(httperr.ObError)
			if !ok || e.Status() != tc.want || resp != nil {
				t.Fatalf("want status %d, got %+v, %v", tc.want, resp, err)
			}
			if strings.Contains(err.Error(), "private") || strings.Contains(err.Error(), "SELECT") || strings.Contains(err.Error(), "v1.4.3") {
				t.Fatalf("upstream details exposed: %v", err)
			}
		})
	}
}

func TestQuerySqlStatsPreservesRequestAndResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/tenants/synthetic/sql-stats" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var req apimodel.QuerySqlStatsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
		}
		if req.SortByColumn != "elapsed_time" || req.SortOrder != "desc" || req.PageNum != 2 || req.PageSize != 10 || len(req.OutputColumns) != 1 || req.OutputColumns[0] != "elapsed_time" {
			t.Errorf("request changed: %+v", req)
		}
		json.NewEncoder(w).Encode(map[string]any{"successful": true, "data": apimodel.SqlStatsResponse{Items: []apimodel.SqlStatsItem{{SqlId: "synthetic-sql"}}, TotalCount: 11}})
	}))
	defer srv.Close()
	resp, err := NewClient(srv.URL).QuerySqlStats("synthetic", apimodel.QuerySqlStatsRequest{SortByColumn: "elapsed_time", SortOrder: "desc", PageNum: 2, PageSize: 10, OutputColumns: []string{"elapsed_time"}})
	if err != nil || resp == nil || resp.TotalCount != 11 || len(resp.Items) != 1 || resp.Items[0].SqlId != "synthetic-sql" {
		t.Fatalf("response changed: %+v, %v", resp, err)
	}
}
