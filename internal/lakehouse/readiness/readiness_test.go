package readiness

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fakeReader struct {
	values   map[string]string
	requests []string
	err      error
}

func (f *fakeReader) Get(_ context.Context, path string, out any) error {
	f.requests = append(f.requests, path)
	if f.err != nil {
		return f.err
	}
	value, ok := f.values[path]
	if !ok {
		return &APIError{Status: 404}
	}
	return json.Unmarshal([]byte(value), out)
}
func options() Options {
	return Options{Namespace: "qa", Release: "ob", StorageJob: "ob-storage-init-1", Dashboard: "oceanbase-dashboard-ob", ObserverReplicas: 1, LogServiceReplicas: 3, Timeout: time.Second, Interval: time.Millisecond}
}
func readyReader() *fakeReader {
	return &fakeReader{values: map[string]string{
		"/apis/batch/v1/namespaces/qa/jobs/ob-storage-init-1":                                          `{"status":{"conditions":[{"type":"Complete","status":"True"}]}}`,
		"/apis/oceanbase.oceanbase.com/v1alpha1/namespaces/qa/oblogserviceclusters/ob-ls":              `{"status":{"status":"running"}}`,
		"/apis/oceanbase.oceanbase.com/v1alpha1/namespaces/qa/obclusters/ob-ob":                        `{"status":{"status":"running"}}`,
		"/api/v1/namespaces/qa/pods?labelSelector=ref-oblogservicecluster%3Dob-ls%2Coblogservice-node": `{"items":[{"status":{"phase":"Running","conditions":[{"type":"Ready","status":"True"}]}},{"status":{"phase":"Running","conditions":[{"type":"Ready","status":"True"}]}},{"status":{"phase":"Running","conditions":[{"type":"Ready","status":"True"}]}}]}`,
		"/api/v1/namespaces/qa/pods?labelSelector=ref-obcluster%3Dob-ob%2Cref-observer":                `{"items":[{"status":{"phase":"Running","conditions":[{"type":"Ready","status":"True"}]}}]}`,
		"/apis/apps/v1/namespaces/qa/deployments/oceanbase-dashboard-ob":                               `{"metadata":{"generation":2},"spec":{"replicas":1},"status":{"observedGeneration":2,"readyReplicas":1,"updatedReplicas":1,"availableReplicas":1}}`,
	}}
}
func TestWaitReady(t *testing.T) {
	r := readyReader()
	var out bytes.Buffer
	if err := Wait(context.Background(), r, options(), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Dashboard ready") || len(r.requests) != 6 {
		t.Fatalf("output=%s requests=%v", out.String(), r.requests)
	}
	for _, path := range r.requests {
		if strings.Contains(path, "secret") {
			t.Fatal("readiness must not read secrets")
		}
	}
}
func TestSnapshotRequiresRuntimeReadiness(t *testing.T) {
	cases := []struct{ name, path, body, want string }{
		{"storage pending", "/apis/batch/v1/namespaces/qa/jobs/ob-storage-init-1", `{"status":{}}`, "Storage initialization"},
		{"storage failed", "/apis/batch/v1/namespaces/qa/jobs/ob-storage-init-1", `{"status":{"conditions":[{"type":"Failed","status":"True","reason":"BackoffLimitExceeded"}]}}`, "Storage initialization failed"},
		{"ls pending", "/apis/oceanbase.oceanbase.com/v1alpha1/namespaces/qa/oblogserviceclusters/ob-ls", `{"status":{"status":"new"}}`, "LogService: new"},
		{"observer missing", "/api/v1/namespaces/qa/pods?labelSelector=ref-obcluster%3Dob-ob%2Cref-observer", `{"items":[]}`, "OceanBase Pods ready 0/1"},
		{"observer unready", "/api/v1/namespaces/qa/pods?labelSelector=ref-obcluster%3Dob-ob%2Cref-observer", `{"items":[{"status":{"phase":"Pending","containerStatuses":[{"state":{"waiting":{"reason":"ImagePullBackOff"}}}]}}]}`, "ImagePullBackOff"},
		{"dashboard stale", "/apis/apps/v1/namespaces/qa/deployments/oceanbase-dashboard-ob", `{"metadata":{"generation":2},"spec":{"replicas":1},"status":{"observedGeneration":1,"readyReplicas":1,"updatedReplicas":1,"availableReplicas":1}}`, "Dashboard"},
		{"dashboard scaled down", "/apis/apps/v1/namespaces/qa/deployments/oceanbase-dashboard-ob", `{"spec":{"replicas":0}}`, "Dashboard"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := readyReader()
			r.values[tc.path] = tc.body
			detail, ready, err := snapshot(context.Background(), r, options())
			if err != nil || ready || !strings.Contains(detail, tc.want) {
				t.Fatalf("detail=%q ready=%v err=%v", detail, ready, err)
			}
		})
	}
}
func TestWaitFailsActionably(t *testing.T) {
	for _, tc := range []struct {
		name   string
		reader *fakeReader
		want   string
	}{
		{"authorization", &fakeReader{err: &APIError{Status: 403}}, "RoleBinding"},
		{"api unavailable", &fakeReader{err: errors.New("credential=https://private.example")}, "Kubernetes API temporarily unavailable"},
		{"missing resource", &fakeReader{}, "Storage initialization Job has not been created"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := options()
			o.Timeout = 5 * time.Millisecond
			var out bytes.Buffer
			err := Wait(context.Background(), tc.reader, o, &out)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v output=%s", err, out.String())
			}
			if strings.Contains(out.String()+err.Error(), "private.example") {
				t.Fatal("API failure disclosed request details")
			}
		})
	}
}
func TestFailedStorageStopsWithoutWaitingForTimeout(t *testing.T) {
	r := readyReader()
	r.values["/apis/batch/v1/namespaces/qa/jobs/ob-storage-init-1"] = `{"status":{"conditions":[{"type":"Failed","status":"True"}]}}`
	var out bytes.Buffer
	err := Wait(context.Background(), r, options(), &out)
	if err == nil || !strings.Contains(err.Error(), "helm upgrade") || len(r.requests) != 1 {
		t.Fatalf("err=%v requests=%v", err, r.requests)
	}
}
func TestWaitDoesNotRequireDisabledDashboard(t *testing.T) {
	o := options()
	o.Dashboard = ""
	r := readyReader()
	delete(r.values, "/apis/apps/v1/namespaces/qa/deployments/oceanbase-dashboard-ob")
	var out bytes.Buffer
	if err := Wait(context.Background(), r, o, &out); err != nil {
		t.Fatal(err)
	}
}
func TestAPIClientOnlyGetsAndRefreshesToken(t *testing.T) {
	token := filepath.Join(t.TempDir(), "token")
	os.WriteFile(token, []byte("first"), 0600)
	seen := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("unexpected method %s", r.Method)
		}
		seen = append(seen, r.Header.Get("Authorization"))
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer server.Close()
	c := &APIClient{BaseURL: server.URL, TokenFile: token, Client: server.Client()}
	var value map[string]bool
	if err := c.Get(context.Background(), "/api/v1/pods", &value); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(token, []byte("second"), 0600)
	if err := c.Get(context.Background(), "/api/v1/pods", &value); err != nil {
		t.Fatal(err)
	}
	if strings.Join(seen, ",") != "Bearer first,Bearer second" {
		t.Fatal(seen)
	}
}
func TestResponseSizeIsBounded(t *testing.T) {
	var v any
	if err := DecodeJSON(strings.NewReader(strings.Repeat(" ", (8<<20)+1)), &v); err == nil {
		t.Fatal("unbounded response accepted")
	}
}
func TestOptionsRejectInvalidNamesAndCounts(t *testing.T) {
	for _, change := range []func(*Options){func(o *Options) { o.Namespace = "../other" }, func(o *Options) { o.StorageJob = "" }, func(o *Options) { o.ObserverReplicas = 0 }, func(o *Options) { o.Timeout = 0 }} {
		o := options()
		change(&o)
		if o.Validate() == nil {
			t.Fatal("invalid configuration accepted")
		}
	}
}
