package ssmonitor

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestQueryPreservesUnavailableAndNonfinite(t *testing.T) {
	for _, test := range []struct {
		name, body string
		status     int
		available  bool
		points     int
	}{
		{"finite", "1", 200, true, 1}, {"NaN", "NaN", 200, false, 0}, {"positive infinity", "+Inf", 200, false, 0}, {"bad number", "broken", 200, false, 0}, {"upstream failure", "1", 503, false, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Query().Get("query"), "FILTER") {
					t.Error("unsubstituted filter")
				}
				w.WriteHeader(test.status)
				fmt.Fprintf(w, `{"status":"success","data":{"result":[{"values":[[%d,%q]]}]}}`, time.Now().Unix(), test.body)
			}))
			defer server.Close()
			result, err := Query(context.Background(), server.URL, "obcluster", "test", "cluster", 15)
			if err != nil {
				t.Fatal(err)
			}
			for _, s := range result {
				if s.Available != test.available || len(s.Points) != test.points {
					t.Fatalf("%s: %+v", s.Key, s)
				}
				if test.status != 200 && s.Error == "" {
					t.Fatal("upstream failure hidden")
				}
			}
		})
	}
}
func TestQueryRejectsUnboundedRequests(t *testing.T) {
	if _, err := Query(context.Background(), "", "anything", "ns", "name", 15); err == nil {
		t.Fatal("unknown scope accepted")
	}
	if _, err := Query(context.Background(), "", "obcluster", "ns", "name", 100000); err == nil {
		t.Fatal("unbounded range accepted")
	}
}
func TestPrometheusLabelEscaping(t *testing.T) {
	e := &emitter{typed: map[string]bool{}}
	e.emit("ss_test", "gauge", map[string]string{"name": "a\"\nb\\c"}, 1)
	if !strings.Contains(e.String(), `name="a\"\nb\\c"`) {
		t.Fatalf("invalid metric labels: %s", e.String())
	}
}
func TestSSGaugeClassification(t *testing.T) {
	for _, id := range []int{240009, 240010, 240019} {
		if !statGauge(id) {
			t.Fatalf("occupancy %d treated as counter", id)
		}
	}
	for _, id := range []int{240001, 240002, 240022, 240023} {
		if statGauge(id) {
			t.Fatalf("counter %d treated as gauge", id)
		}
	}
}
