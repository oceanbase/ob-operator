package lakehouse

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestDashboardHostPriorityAndValidation(t *testing.T) {
	for _, tt := range []struct {
		name, explicit, saved, external, internal, discovered, want, source string
		calls                                                               int
		wantError                                                           bool
	}{
		{name: "explicit", explicit: "dashboard.example.com", saved: "8.8.8.8", external: "1.1.1.1", internal: "1.0.0.1", discovered: "8.8.4.4", want: "dashboard.example.com", source: "--dashboard-host"},
		{name: "explicit-repairs-invalid-saved", explicit: "8.8.8.8", saved: "bad host", want: "8.8.8.8", source: "--dashboard-host"},
		{name: "saved", saved: "dashboard.example.com\n", external: "1.1.1.1", internal: "1.0.0.1", discovered: "8.8.4.4", want: "dashboard.example.com", source: "saved --dashboard-host"},
		{name: "external", external: "1.1.1.1", internal: "1.0.0.1", discovered: "8.8.4.4", want: "1.1.1.1", source: "Kubernetes Node ExternalIP"},
		{name: "public-internal", external: "10.0.0.2", internal: "1.0.0.1", discovered: "8.8.4.4", want: "1.0.0.1", source: "public Kubernetes node address"},
		{name: "metadata", external: "192.168.0.2", internal: "172.16.177.227", discovered: "8.8.4.4", want: "8.8.4.4", source: "test metadata", calls: 1},
		{name: "unavailable", internal: "172.16.177.227", calls: 1},
		{name: "private-discovery", internal: "172.16.177.227", discovered: "10.0.0.2", calls: 1},
		{name: "invalid-discovery", internal: "172.16.177.227", discovered: "not-an-ip", calls: 1},
		{name: "reserved-external", external: "192.0.2.1", internal: "172.16.177.227", discovered: "8.8.4.4", want: "8.8.4.4", source: "test metadata", calls: 1},
		{name: "invalid-saved", saved: "http://bad.example/path", wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			state := t.TempDir()
			if tt.saved != "" {
				if err := os.WriteFile(filepath.Join(state, "dashboard-host"), []byte(tt.saved), 0600); err != nil {
					t.Fatal(err)
				}
			}
			result := PreflightResult{NodeName: "selected-worker", NodeSystemUUID: "selected-uuid", NodeInternalIP: tt.internal, NodeExternalIP: tt.external}
			calls := 0
			host, source, err := dashboardHost(context.Background(), options{host: tt.explicit}, result, state, func(_ context.Context, got PreflightResult) (string, string) {
				calls++
				if got.NodeName != result.NodeName || got.NodeSystemUUID != result.NodeSystemUUID || got.NodeInternalIP != result.NodeInternalIP {
					t.Fatal("discovery must receive the selected node identity", got)
				}
				return tt.discovered, "test metadata"
			})
			if (err != nil) != tt.wantError || host != tt.want || source != tt.source || calls != tt.calls {
				t.Fatalf("host=%q source=%q err=%v discovery calls=%d; want host=%q source=%q error=%v calls=%d", host, source, err, calls, tt.want, tt.source, tt.wantError, tt.calls)
			}
		})
	}
}

type publicAccessFixture struct {
	server      *httptest.Server
	run         RunCommand
	port        int
	requests    []string
	requestMu   sync.Mutex
	nodeReads   int
	secretReads int
}

func newPublicAccessFixture(t *testing.T, serviceType string) *publicAccessFixture {
	t.Helper()
	f := &publicAccessFixture{}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.requestMu.Lock()
		f.requests = append(f.requests, r.URL.Path)
		f.requestMu.Unlock()
		fmt.Fprint(w, "ready")
	}))
	t.Cleanup(f.server.Close)
	_, port, err := net.SplitHostPort(strings.TrimPrefix(f.server.URL, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	f.port, _ = strconv.Atoi(port)
	f.run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("Kubernetes reads must remain bounded")
		}
		command := name + " " + strings.Join(args, " ")
		if name != "kubectl" || !strings.Contains(command, " get ") {
			t.Fatalf("unexpected mutation: %s", command)
		}
		switch {
		case strings.Contains(command, " get service "):
			return json.Marshal(map[string]interface{}{"spec": map[string]interface{}{"type": serviceType, "clusterIP": "127.0.0.1", "ports": []interface{}{map[string]interface{}{"port": f.port, "nodePort": f.port}}}})
		case strings.Contains(command, " get pods "):
			return []byte(`{"items":[{"spec":{"nodeName":"dashboard-worker"}}]}`), nil
		case strings.Contains(command, " get node dashboard-worker "):
			f.nodeReads++
			return []byte(`{"status":{"addresses":[{"Type":"InternalIP","Address":"127.0.0.1"}],"nodeInfo":{"systemUUID":"dashboard-worker-uuid"}}}`), nil
		case strings.Contains(command, " get secret demo-dashboard-bootstrap "):
			f.secretReads++
			return json.Marshal(map[string]interface{}{"data": map[string]string{"password": base64.StdEncoding.EncodeToString([]byte("PUBLIC_ACCESS_SECRET_SENTINEL"))}})
		default:
			t.Fatalf("unexpected read: %s", command)
			return nil, nil
		}
	}
	return f
}

func (f *publicAccessFixture) assertPrivateAccess(t *testing.T, state, output string) {
	t.Helper()
	f.requestMu.Lock()
	paths := append([]string(nil), f.requests...)
	f.requestMu.Unlock()
	if strings.Join(paths, ",") != "/,/api/v1/info" {
		t.Fatal("readiness must use the actual service, not the advertised address", paths)
	}
	if f.secretReads != 1 || strings.Contains(output, "PUBLIC_ACCESS_SECRET_SENTINEL") {
		t.Fatal("credential must be fetched once and never printed", output)
	}
	for _, name := range []string{"dashboard-initial-password", "ACCESS.txt"} {
		info, err := os.Stat(filepath.Join(state, name))
		if err != nil || info.Mode().Perm() != 0600 {
			t.Fatalf("%s must be private: %v", name, err)
		}
	}
	access, err := os.ReadFile(filepath.Join(state, "ACCESS.txt"))
	if err != nil || string(access) != output {
		t.Fatal("saved access instructions must match the terminal output")
	}
}

func TestDashboardPublicAddressSummary(t *testing.T) {
	for _, tt := range []struct {
		name, serviceType, discovered string
		readNode                      bool
	}{
		{"discovered", "NodePort", "8.8.4.4", false},
		{"show-access-reads-node", "NodePort", "8.8.4.4", true},
		{"unavailable", "NodePort", "", false},
		{"invalid-private-discovery", "NodePort", "10.1.2.3", false},
		{"clusterip", "ClusterIP", "", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := newPublicAccessFixture(t, tt.serviceType)
			state := filepath.Join(t.TempDir(), "state")
			opts := options{release: "demo", namespace: "demo-ns", showAccess: tt.readNode}
			if tt.serviceType == "ClusterIP" {
				opts.host = "override.example.com"
			}
			result := PreflightResult{NodeInternalIP: "127.0.0.1"}
			if tt.readNode {
				result = PreflightResult{}
			}
			discoveryCalls := 0
			var out bytes.Buffer
			err := showAccessWithDiscovery(context.Background(), f.run, opts, result, state, &out, func(_ context.Context, got PreflightResult) (string, string) {
				discoveryCalls++
				if tt.readNode && (got.NodeName != "dashboard-worker" || got.NodeSystemUUID != "dashboard-worker-uuid" || got.NodeInternalIP != "127.0.0.1") {
					t.Fatal("show-access discovery did not get the selected Dashboard node identity", got)
				}
				return tt.discovered, "test cloud metadata"
			})
			if err != nil {
				t.Fatal(err)
			}
			output := out.String()
			f.assertPrivateAccess(t, state, output)
			if strings.Contains(output, "HTTP verified from this node: http") {
				t.Fatal("do not print a second internal HTTP URL as if it were another browser entry", output)
			}
			if tt.readNode && (f.nodeReads != 1 || !strings.Contains(output, "Dashboard access") || strings.Contains(output, "Installation complete")) {
				t.Fatal("show-access must query the Dashboard node and avoid claiming installation", output)
			}
			if tt.serviceType == "ClusterIP" {
				if discoveryCalls != 0 || strings.Contains(output, opts.host) || !strings.Contains(output, "Dashboard (cluster network only): ") || !strings.Contains(output, "kubectl -n demo-ns port-forward service/oceanbase-dashboard-demo 18080:") {
					t.Fatal("ClusterIP must ignore public host and explain port-forward", output)
				}
			} else if tt.discovered == "8.8.4.4" {
				want := fmt.Sprintf("Dashboard: http://8.8.4.4:%d/#/login", f.port)
				if discoveryCalls != 1 || !strings.Contains(output, want) || !strings.Contains(output, "Address source: test cloud metadata") || strings.Contains(output, f.server.URL) {
					t.Fatal("public address must be the single advertised Dashboard entry", output)
				}
			} else {
				if discoveryCalls != 1 || !strings.Contains(output, "External Dashboard address: unavailable") || !strings.Contains(output, "Dashboard (private network only): "+f.server.URL+"/#/login") || strings.Contains(output, "\nDashboard: ") || !strings.Contains(output, "--release demo --namespace demo-ns --dashboard-host") {
					t.Fatal("unknown public address must not advertise an internal IP as public", output)
				}
			}
			if _, err := os.Stat(filepath.Join(state, "dashboard-host")); !os.IsNotExist(err) {
				t.Fatal("automatic discovery and ClusterIP must not persist a public host", err)
			}
		})
	}
}

func TestDashboardExplicitHostPersistsAndCanBeReplaced(t *testing.T) {
	state := filepath.Join(t.TempDir(), "state")
	for _, tt := range []struct{ name, explicit, want, source string }{
		{"initial", "dashboard.example.com", "dashboard.example.com", "--dashboard-host"},
		{"reuse", "", "dashboard.example.com", "saved --dashboard-host"},
		{"replace", "1.1.1.1", "1.1.1.1", "--dashboard-host"},
		{"reuse-replacement", "", "1.1.1.1", "saved --dashboard-host"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := newPublicAccessFixture(t, "NodePort")
			var out bytes.Buffer
			err := showAccessWithDiscovery(context.Background(), f.run, options{release: "demo", namespace: "demo-ns", host: tt.explicit, showAccess: true}, PreflightResult{NodeInternalIP: "127.0.0.1", NodeExternalIP: "8.8.8.8"}, state, &out, func(context.Context, PreflightResult) (string, string) {
				t.Fatal("explicit or saved address must avoid cloud discovery")
				return "", ""
			})
			if err != nil {
				t.Fatal(err)
			}
			f.assertPrivateAccess(t, state, out.String())
			if !strings.Contains(out.String(), fmt.Sprintf("Dashboard: http://%s:%d/#/login", tt.want, f.port)) || !strings.Contains(out.String(), "Address source: "+tt.source+"\n") {
				t.Fatal("explicit/saved Dashboard host was not used", out.String())
			}
			path := filepath.Join(state, "dashboard-host")
			info, err := os.Stat(path)
			if err != nil || info.Mode().Perm() != 0600 {
				t.Fatal("saved host must be private", err)
			}
			saved, err := os.ReadFile(path)
			if err != nil || strings.TrimSpace(string(saved)) != tt.want {
				t.Fatal("saved explicit host did not survive or was not replaced", err)
			}
		})
	}
}
