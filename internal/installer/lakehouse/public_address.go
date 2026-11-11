package lakehouse

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"
)

const alibabaMetadataURL = "http://100.100.100.200"

// discoverDashboardHost only advertises addresses assigned to the target worker.
// An external IP echo service could instead return a shared NAT gateway address.
func discoverDashboardHost(ctx context.Context, result PreflightResult) (host, source string) {
	if checkLocalWorker(result) != nil {
		return "", ""
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	client := newDashboardMetadataClient()
	defer client.CloseIdleConnections()
	host = discoverAlibabaDashboardHost(ctx, client, alibabaMetadataURL, result.NodeInternalIP)
	if host == "" {
		return "", ""
	}
	return host, "Alibaba Cloud instance metadata"
}

func newDashboardMetadataClient() *http.Client {
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
		ResponseHeaderTimeout: time.Second,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// The injectable client and base URL keep tests independent of cloud metadata.
// Production always uses the fixed instance endpoint above and never falls back
// to unauthenticated metadata or queries credential endpoints.
// Protocol: https://help.aliyun.com/zh/ecs/user-guide/view-instance-metadata/
func discoverAlibabaDashboardHost(ctx context.Context, client *http.Client, baseURL, internalIP string) string {
	wantPrivate, err := netip.ParseAddr(internalIP)
	if err != nil || !wantPrivate.Is4() {
		return ""
	}
	token, ok := dashboardMetadataValue(ctx, client, http.MethodPut, baseURL+"/latest/api/token", "", true)
	if !ok || token == "" || strings.ContainsAny(token, "\r\n") {
		return ""
	}
	privateIP, ok := dashboardMetadataValue(ctx, client, http.MethodGet, baseURL+"/latest/meta-data/private-ipv4", token, false)
	gotPrivate, err := netip.ParseAddr(privateIP)
	if !ok || err != nil || gotPrivate != wantPrivate {
		return ""
	}
	for _, key := range []string{"eipv4", "public-ipv4"} {
		value, ok := dashboardMetadataValue(ctx, client, http.MethodGet, baseURL+"/latest/meta-data/"+key, token, false)
		ip, err := netip.ParseAddr(value)
		if ok && err == nil && ip.Is4() && isPublicDashboardIP(value) {
			return value
		}
	}
	return ""
}

func dashboardMetadataValue(ctx context.Context, client *http.Client, method, url, token string, tokenRequest bool) (string, bool) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return "", false
	}
	if tokenRequest {
		req.Header.Set("X-aliyun-ecs-metadata-token-ttl-seconds", "60")
	} else {
		req.Header.Set("X-aliyun-ecs-metadata-token", token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", false
	}
	const maxResponse = 4 << 10
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse+1))
	if err != nil || len(body) > maxResponse {
		return "", false
	}
	return strings.TrimSpace(string(body)), true
}

// Reject private, shared, documentation and reserved address ranges. A failed
// classification leaves the installer with its explicitly labelled private URL.
func isPublicDashboardIP(value string) bool {
	ip, err := netip.ParseAddr(strings.TrimSpace(value))
	if err != nil || ip.Is4In6() || !ip.IsGlobalUnicast() || ip.IsPrivate() {
		return false
	}
	if ip.Is6() && !dashboardGlobalIPv6.Contains(ip) {
		return false
	}
	for _, prefix := range nonPublicDashboardIPs {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}

var dashboardGlobalIPv6 = netip.MustParsePrefix("2000::/3")

var nonPublicDashboardIPs = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	// Special protocol allocations are not ordinary server address ranges.
	netip.MustParsePrefix("2001::/23"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
	netip.MustParsePrefix("3ffe::/16"),
	netip.MustParsePrefix("3fff::/20"),
}
