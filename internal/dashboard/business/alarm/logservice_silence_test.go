package alarm

import (
	"context"
	"encoding/json"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/silence"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/oceanbase"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestLogServiceSilenceIdentity(t *testing.T) {
	for _, v := range []string{"ls", "/ls", "ns/", "ns/a|b", "ns/ls/other"} {
		if _, _, err := logServiceSilenceIdentity(v); err == nil {
			t.Fatalf("invalid identity %q accepted", v)
		}
	}
	ns, name, err := logServiceSilenceIdentity("test/ls.one")
	if err != nil || ns != "test" || name != "ls.one" {
		t.Fatal("valid identity rejected")
	}
}

type silenceTransport func(*http.Request) (*http.Response, error)

func (f silenceTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestLogServiceSilenceIsNamespaceScoped(t *testing.T) {
	client := getClient()
	old := client.GetClient().Transport
	defer client.SetTransport(old)
	calls := 0
	client.SetTransport(silenceTransport(func(req *http.Request) (*http.Response, error) {
		calls++
		var body struct {
			Matchers []struct {
				Name, Value string
				IsRegex     bool `json:"isRegex"`
			} `json:"matchers"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		matches := map[string]string{}
		for _, m := range body.Matchers {
			matches[m.Name] = m.Value
		}
		if matches["namespace"] != "test" || matches["logservice"] != "ls\\.one" {
			t.Fatalf("bad matchers: %v", matches)
		}
		if _, ok := matches["ob_cluster_name"]; ok {
			t.Fatal("LS silence accidentally requires an OBCluster")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"silenceID":"test-id"}`)), Request: req}, nil
	}))
	p := &silence.SilencerParam{Instances: []oceanbase.OBInstance{{Type: oceanbase.TypeLogService, LogService: "test/ls.one"}}, Rules: []string{"ss_ls_replica_shortage"}}
	result, err := CreateOrUpdateSilencer(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Instances) != 1 || result.Instances[0].LogService != "test/ls.one" {
		t.Fatalf("invalid response: %+v", result)
	}
	p.Instances = append(p.Instances, oceanbase.OBInstance{Type: oceanbase.TypeLogService, LogService: "other/ls-two"})
	if _, err := CreateOrUpdateSilencer(context.Background(), p); err == nil {
		t.Fatal("cross-namespace silence accepted")
	}
	if calls != 1 {
		t.Fatal("invalid silence reached Alertmanager")
	}
}
