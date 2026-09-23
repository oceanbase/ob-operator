// Package objectstorage manages references to S3 credentials, never credential
// reads/updates, and performs bounded, read-only bucket preflight checks.
package objectstorage

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"

	oberr "github.com/oceanbase/ob-operator/pkg/errors"
	k8s "github.com/oceanbase/ob-operator/pkg/k8s/client"
)

type Service struct{ Core kubernetes.Interface }

func Default() *Service { return &Service{Core: k8s.GetClient().ClientSet} }

type CredentialReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
type CheckRequest struct {
	BucketURL  string `json:"bucketURL"`
	SecretName string `json:"secretName"`
}
type CheckResult struct {
	OK         bool      `json:"ok"`
	Code       string    `json:"code"`
	CheckedAt  time.Time `json:"checkedAt"`
	HTTPStatus int       `json:"httpStatus,omitempty"`
	Scope      string    `json:"scope"`
}
type Location struct{ Endpoint, Bucket, Prefix, Region string }

var bucketName = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`)
var regionName = regexp.MustCompile(`^[a-zA-Z0-9-]{1,64}$`)
var prefixName = regexp.MustCompile(`^[a-zA-Z0-9_./-]*$`)
var locationOptions = map[string]*regexp.Regexp{
	"scope":         regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`),
	"max_iops":      regexp.MustCompile(`^[0-9]{1,20}$`),
	"max_bandwidth": regexp.MustCompile(`(?i)^[0-9]+(?:\.[0-9]+)?(?:[KMGTPE]?B)?$`),
}

// Keep OceanBase's raw host=http://... grammar. This deliberately accepts only
// the S3 subset surfaced by the form; providers/STS are not inferred from names.
func ParseLocation(raw string) (*Location, error) {
	fail := func() (*Location, error) {
		return nil, oberr.NewBadRequest("Use s3://bucket[/prefix]?host=http(s)://endpoint&s3_region=region; only scope, max_iops and max_bandwidth options are supported; inline credentials are not supported")
	}
	u, err := url.Parse(raw)
	if err != nil || len(raw) > 2048 || u.Scheme != "s3" || u.User != nil || u.Fragment != "" || !bucketName.MatchString(u.Host) || strings.Contains(u.Host, "..") || net.ParseIP(u.Host) != nil {
		return fail()
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil || len(q["host"]) != 1 || len(q["s3_region"]) != 1 || !regionName.MatchString(q.Get("s3_region")) {
		return fail()
	}
	// Extra options describe OceanBase placement/limits, not the S3 request.
	// Reject unknown keys rather than silently accepting inline credentials.
	for key, values := range q {
		if key == "host" || key == "s3_region" {
			continue
		}
		pattern, ok := locationOptions[key]
		if !ok || len(values) != 1 || len(values[0]) > 64 || !pattern.MatchString(values[0]) {
			return fail()
		}
	}
	e, err := url.Parse(q.Get("host"))
	if err != nil || (e.Scheme != "http" && e.Scheme != "https") || e.Hostname() == "" || e.User != nil || e.RawQuery != "" || e.Fragment != "" || (e.Path != "" && e.Path != "/") || strings.ContainsAny(e.Host, "\\%") {
		return fail()
	}
	if e.Port() != "" {
		p, err := strconv.Atoi(e.Port())
		if err != nil || p < 1 || p > 65535 {
			return fail()
		}
	}
	if e.Hostname() == "localhost" || strings.HasSuffix(strings.ToLower(e.Hostname()), ".localhost") {
		return fail()
	}
	if ip := net.ParseIP(e.Hostname()); ip != nil && !allowedIP(ip) {
		return fail()
	}
	if !prefixName.MatchString(u.Path) {
		return fail()
	}
	for _, part := range strings.Split(u.Path, "/") {
		if part == "." || part == ".." {
			return fail()
		}
	}
	return &Location{Endpoint: strings.TrimSuffix(e.String(), "/"), Bucket: u.Host, Prefix: strings.TrimPrefix(u.Path, "/"), Region: q.Get("s3_region")}, nil
}

func validIdentity(ns, name string) bool {
	return len(validation.IsDNS1123Label(ns)) == 0 && len(validation.IsDNS1123Subdomain(name)) == 0
}

func (s *Service) List(ctx context.Context, ns string) ([]CredentialReference, error) {
	if !validIdentity(ns, "credentials") {
		return nil, oberr.NewBadRequest("A valid namespace is required")
	}
	items, err := s.Core.CoreV1().Secrets(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, oberr.NewInternal("Cannot list object storage credential references")
	}
	refs := []CredentialReference{}
	for _, v := range items.Items {
		if v.DeletionTimestamp == nil && v.Type == corev1.SecretTypeOpaque && len(v.Data["access_id"]) > 0 && len(v.Data["access_key"]) > 0 {
			refs = append(refs, CredentialReference{Name: v.Name, Namespace: ns})
		}
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].Name < refs[j].Name })
	return refs, nil
}

// Create only. Existing Secrets are never replaced, patched or attached to a
// cluster owner. A later failed/cancelled wizard must not delete shared secrets.
func (s *Service) Create(ctx context.Context, ns, name, id, key string) (*CredentialReference, error) {
	if !validIdentity(ns, name) || strings.TrimSpace(id) == "" || strings.TrimSpace(key) == "" || len(id) > 200 || len(key) > 200 {
		return nil, oberr.NewBadRequest("Valid namespace/name and non-empty credentials (max 200 bytes each) are required")
	}
	namespace, err := s.Core.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
	if err != nil || namespace.DeletionTimestamp != nil {
		return nil, oberr.NewBadRequest("Choose an existing, non-deleting namespace")
	}
	immutable := true
	_, err = s.Core.CoreV1().Secrets(ns).Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns, Labels: map[string]string{"app.kubernetes.io/managed-by": "oceanbase-dashboard", "oceanbase.com/credential-purpose": "object-storage"}},
		Type:       corev1.SecretTypeOpaque, Immutable: &immutable,
		Data: map[string][]byte{"access_id": []byte(id), "access_key": []byte(key)},
	}, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		return nil, oberr.New(oberr.ErrConflict, "Secret already exists; select it or choose another name. Existing credentials were not changed")
	}
	if err != nil {
		return nil, oberr.NewInternal("Cannot create object storage credentials")
	}
	return &CredentialReference{Name: name, Namespace: ns}, nil
}

// Internal MinIO/private networks are intentional. Block loopback, metadata,
// link-local, unspecified and multicast addresses, including DNS rebinding.
func allowedIP(ip net.IP) bool {
	if !ip.IsGlobalUnicast() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return false
	}
	if v := ip.To4(); v != nil {
		return v[0] != 0 && v[0] < 224 && !ip.Equal(net.ParseIP("100.100.100.200"))
	}
	return true
}
func safeHTTPClient() *http.Client {
	transport := &http.Transport{Proxy: nil, DisableKeepAlives: true, TLSHandshakeTimeout: 4 * time.Second, ResponseHeaderTimeout: 5 * time.Second,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil || len(ips) == 0 {
				return nil, fmt.Errorf("endpoint DNS resolution failed")
			}
			for _, ip := range ips {
				if !allowedIP(ip.IP) {
					return nil, fmt.Errorf("endpoint address is not permitted")
				}
			}
			// Dial the validated IP, never resolve a hostname again after checking.
			return (&net.Dialer{Timeout: 4 * time.Second}).DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
		},
	}
	return &http.Client{Transport: transport, Timeout: 7 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
}

var checkSlots = make(chan struct{}, 4)

func (s *Service) Check(ctx context.Context, ns string, p CheckRequest) (*CheckResult, error) {
	if !validIdentity(ns, p.SecretName) {
		return nil, oberr.NewBadRequest("Valid namespace and Secret reference are required")
	}
	loc, err := ParseLocation(p.BucketURL)
	if err != nil {
		return nil, err
	}
	select {
	case checkSlots <- struct{}{}:
		defer func() { <-checkSlots }()
	default:
		return nil, oberr.New(oberr.ErrConflict, "Object storage checks are busy; retry shortly")
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	secret, err := s.Core.CoreV1().Secrets(ns).Get(ctx, p.SecretName, metav1.GetOptions{})
	if err != nil || secret.DeletionTimestamp != nil || len(secret.Data["access_id"]) == 0 || len(secret.Data["access_key"]) == 0 {
		return nil, oberr.NewBadRequest("Cannot use Secret: it must exist in this namespace and contain access_id and access_key")
	}
	return checkBucket(ctx, loc, string(secret.Data["access_id"]), string(secret.Data["access_key"]), safeHTTPClient())
}

func checkBucket(ctx context.Context, loc *Location, id, key string, client *http.Client) (*CheckResult, error) {
	r := &CheckResult{CheckedAt: time.Now().UTC(), Scope: "HeadBucket only; no object read/write/delete or prefix permission validation"}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, loc.Endpoint+"/"+loc.Bucket, nil)
	if err != nil {
		return nil, oberr.NewBadRequest("Invalid endpoint")
	}
	_, err = v4.NewSigner(credentials.NewStaticCredentials(id, key, "")).Sign(req, nil, "s3", loc.Region, time.Now())
	if err != nil {
		return nil, oberr.NewInternal("Cannot sign bucket check")
	}
	resp, err := client.Do(req)
	if err != nil {
		r.Code = "connection_failed"
		return r, nil
	} // never return signed URLs, headers or provider responses
	defer resp.Body.Close()
	r.HTTPStatus = resp.StatusCode
	switch {
	case resp.StatusCode == http.StatusOK:
		r.OK = true
		r.Code = "bucket_accessible"
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		r.Code = "redirect_refused"
	case resp.StatusCode == 401 || resp.StatusCode == 403:
		r.Code = "access_denied"
	case resp.StatusCode == 404:
		r.Code = "bucket_not_found_or_denied"
	default:
		r.Code = "endpoint_error"
	}
	return r, nil
}
