package readiness

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// APIError deliberately omits the response body and credentials.
type APIError struct{ Status int }

func (e *APIError) Error() string { return fmt.Sprintf("Kubernetes API returned HTTP %d", e.Status) }

type APIClient struct {
	BaseURL, TokenFile string
	Client             *http.Client
}

// InClusterClient uses the projected ServiceAccount identity only. It never
// loads a kubeconfig or discovers another cluster, and it rejects redirects so
// Kubernetes credentials cannot be sent to an arbitrary destination.
func InClusterClient() (*APIClient, error) {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if host == "" || port == "" {
		return nil, errors.New("Kubernetes service address is unavailable")
	}
	const dir = "/var/run/secrets/kubernetes.io/serviceaccount/"
	ca, err := os.ReadFile(dir + "ca.crt")
	if err != nil {
		return nil, fmt.Errorf("read Kubernetes CA: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(ca) {
		return nil, errors.New("Kubernetes CA is invalid")
	}
	transport := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}, DialContext: (&net.Dialer{Timeout: 10 * time.Second}).DialContext, ResponseHeaderTimeout: 10 * time.Second}
	return &APIClient{BaseURL: "https://" + net.JoinHostPort(host, port), TokenFile: dir + "token", Client: &http.Client{Transport: transport, Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}
func (c *APIClient) Get(ctx context.Context, path string, out any) error {
	if !strings.HasPrefix(path, "/api/") && !strings.HasPrefix(path, "/apis/") {
		return errors.New("invalid Kubernetes API path")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return errors.New("invalid Kubernetes request")
	}
	// Re-read projected tokens on each request to support rotation during a long
	// cluster bootstrap. Tokens are never copied to another resource or printed.
	if c.TokenFile != "" {
		token, err := os.ReadFile(c.TokenFile)
		if err != nil {
			return errors.New("read Kubernetes ServiceAccount token failed")
		}
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))
	}
	req.Header.Set("Accept", "application/json")
	response, err := c.Client.Do(req)
	if err != nil {
		return errors.New("Kubernetes request failed")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return &APIError{Status: response.StatusCode}
	}
	return DecodeJSON(response.Body, out)
}
