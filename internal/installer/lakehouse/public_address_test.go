package lakehouse

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPublicDashboardIPClassification(t *testing.T) {
	for _, value := range []string{"8.145.45.249", "1.1.1.1", "223.255.255.254", "2606:4700:4700::1111", "2408:8000::1"} {
		if !isPublicDashboardIP(value) {
			t.Errorf("public address rejected: %s", value)
		}
	}
	for _, value := range []string{
		"", "host.example", "http://8.145.45.249", "8.145.45.249:80", "8.145.45.249\n1.1.1.1",
		"0.0.0.0", "0.1.2.3", "10.1.2.3", "100.64.0.1", "100.127.255.254", "127.0.0.1",
		"169.254.1.2", "172.16.1.2", "172.31.255.254", "192.0.0.1", "192.0.2.1", "192.88.99.1",
		"192.168.1.1", "198.18.1.1", "198.19.1.1", "198.51.100.1", "203.0.113.1", "224.0.0.1", "240.0.0.1", "255.255.255.255",
		"::", "::1", "::ffff:8.145.45.249", "::ffff:10.0.0.1", "fc00::1", "fe80::1", "ff02::1",
		"100::1", "64:ff9b::808:808", "2001:2::1", "2001:10::1", "2001:20::1", "2001:db8::1", "2002::1", "3ffe::1", "3fff::1",
	} {
		if isPublicDashboardIP(value) {
			t.Errorf("non-public or invalid address accepted: %q", value)
		}
	}
}

func TestAlibabaDashboardHostMetadataProtocol(t *testing.T) {
	for _, fallback := range []bool{false, true} {
		t.Run(fmt.Sprintf("fallback=%t", fallback), func(t *testing.T) {
			var mu sync.Mutex
			var seen []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				seen = append(seen, r.Method+" "+r.URL.Path)
				mu.Unlock()
				if r.URL.Path == "/latest/api/token" {
					if r.Method != http.MethodPut || r.Header.Get("X-aliyun-ecs-metadata-token-ttl-seconds") != "60" || r.Header.Get("X-aliyun-ecs-metadata-token") != "" {
						t.Error("token request did not use the expected IMDSv2 protocol")
					}
					fmt.Fprint(w, "TEST_METADATA_TOKEN")
					return
				}
				if r.Method != http.MethodGet || r.Header.Get("X-aliyun-ecs-metadata-token") != "TEST_METADATA_TOKEN" {
					t.Error("metadata request was unauthenticated or used an incorrect method")
				}
				switch r.URL.Path {
				case "/latest/meta-data/private-ipv4":
					fmt.Fprint(w, "172.16.177.227\n")
				case "/latest/meta-data/eipv4":
					if fallback {
						w.WriteHeader(http.StatusNotFound)
					} else {
						fmt.Fprint(w, "8.145.45.249\n")
					}
				case "/latest/meta-data/public-ipv4":
					fmt.Fprint(w, "8.145.45.249")
				default:
					t.Error("unexpected metadata path", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			client := newDashboardMetadataClient()
			defer client.CloseIdleConnections()
			if got := discoverAlibabaDashboardHost(context.Background(), client, server.URL, "172.16.177.227"); got != "8.145.45.249" {
				t.Fatal("metadata discovery returned", got)
			}
			want := "PUT /latest/api/token,GET /latest/meta-data/private-ipv4,GET /latest/meta-data/eipv4"
			if fallback {
				want += ",GET /latest/meta-data/public-ipv4"
			}
			mu.Lock()
			got := strings.Join(seen, ",")
			mu.Unlock()
			if got != want {
				t.Errorf("request sequence = %q, want %q", got, want)
			}
		})
	}
}

func TestAlibabaDashboardHostRejectsUntrustedMetadata(t *testing.T) {
	for _, tc := range []struct {
		name, token, privateIP, eip, publicIP string
		tokenStatus, privateStatus            int
		maxRequests                           int32
	}{
		{name: "token denied", tokenStatus: 401, maxRequests: 1},
		{name: "no token endpoint", tokenStatus: 404, maxRequests: 1},
		{name: "empty token", maxRequests: 1},
		{name: "oversize token", token: strings.Repeat("a", 4097), maxRequests: 1},
		{name: "header injection", token: "token\r\nUnexpected: value", maxRequests: 1},
		{name: "private mismatch", token: "token", privateIP: "172.16.177.228", maxRequests: 2},
		{name: "invalid private", token: "token", privateIP: "not-an-address", maxRequests: 2},
		{name: "private denied", token: "token", privateStatus: 403, maxRequests: 2},
		{name: "oversize private", token: "token", privateIP: strings.Repeat(" ", 4097) + "172.16.177.227", maxRequests: 2},
		{name: "private public values", token: "token", privateIP: "172.16.177.227", eip: "192.168.1.1", publicIP: "100.64.0.1", maxRequests: 4},
		{name: "metadata IPv6 not IPv4", token: "token", privateIP: "172.16.177.227", eip: "2606:4700:4700::1111", publicIP: "::ffff:8.145.45.249", maxRequests: 4},
		{name: "oversize IP", token: "token", privateIP: "172.16.177.227", eip: strings.Repeat(" ", 4097) + "8.145.45.249", publicIP: "invalid", maxRequests: 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				switch r.URL.Path {
				case "/latest/api/token":
					if tc.tokenStatus != 0 {
						w.WriteHeader(tc.tokenStatus)
					}
					fmt.Fprint(w, tc.token)
				case "/latest/meta-data/private-ipv4":
					if tc.privateStatus != 0 {
						w.WriteHeader(tc.privateStatus)
					}
					fmt.Fprint(w, tc.privateIP)
				case "/latest/meta-data/eipv4":
					fmt.Fprint(w, tc.eip)
				case "/latest/meta-data/public-ipv4":
					fmt.Fprint(w, tc.publicIP)
				default:
					t.Error("unexpected metadata path", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			client := newDashboardMetadataClient()
			defer client.CloseIdleConnections()
			if got := discoverAlibabaDashboardHost(context.Background(), client, server.URL, "172.16.177.227"); got != "" {
				t.Error("invalid metadata advertised an address", got)
			}
			if requests.Load() != tc.maxRequests {
				t.Errorf("request count = %d, want %d", requests.Load(), tc.maxRequests)
			}
		})
	}
}

func TestAlibabaDashboardHostDoesNotFollowRedirectsOrProxy(t *testing.T) {
	var redirected atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirected.Add(1)
		fmt.Fprint(w, "token")
	}))
	defer target.Close()
	t.Setenv("HTTP_PROXY", target.URL)
	t.Setenv("http_proxy", target.URL)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	client := newDashboardMetadataClient()
	defer client.CloseIdleConnections()
	transport := client.Transport.(*http.Transport)
	if transport.Proxy != nil {
		t.Fatal("metadata requests must not use environment proxies")
	}
	if got := discoverAlibabaDashboardHost(context.Background(), client, server.URL, "172.16.177.227"); got != "" || redirected.Load() != 0 {
		t.Fatal("metadata redirect followed or accepted")
	}
}

func TestAlibabaDashboardHostHonorsTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()
	for _, mode := range []string{"client", "context"} {
		t.Run(mode, func(t *testing.T) {
			client := newDashboardMetadataClient()
			defer client.CloseIdleConnections()
			if client.Timeout <= 0 || client.Timeout > time.Second {
				t.Fatal("production metadata client timeout exceeds one second")
			}
			ctx := context.Background()
			if mode == "client" {
				client.Timeout = 20 * time.Millisecond
			} else {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 20*time.Millisecond)
				defer cancel()
			}
			start := time.Now()
			if got := discoverAlibabaDashboardHost(ctx, client, server.URL, "172.16.177.227"); got != "" {
				t.Fatal("timed out request returned an address", got)
			}
			if time.Since(start) > time.Second {
				t.Fatal("metadata request ignored timeout")
			}
		})
	}
}

func TestDiscoverDashboardHostRequiresLocalWorker(t *testing.T) {
	if host, source := discoverDashboardHost(context.Background(), PreflightResult{NodeName: "remote-worker", NodeInternalIP: "not-an-IP"}); host != "" || source != "" {
		t.Fatal("a non-local worker must not use the installer's instance metadata")
	}
}
