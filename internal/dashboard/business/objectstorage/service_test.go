package objectstorage

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLocationValidation(t *testing.T) {
	good := "s3://sharedstorage/cluster-a?host=http://192.0.2.10:9000&s3_region=us-east-1"
	loc, err := ParseLocation(good)
	if err != nil || loc.Endpoint != "http://192.0.2.10:9000" || loc.Prefix != "cluster-a" {
		t.Fatalf("%+v %v", loc, err)
	}
	for _, raw := range []string{
		strings.Replace(good, "s3://", "oss://", 1), good + "&access_key=TOPSECRET", good + "&host=http://elsewhere",
		"s3://user:TOPSECRET@bucket?host=http://s3.example.com&s3_region=r", "s3://bucket?host=http://id:TOPSECRET@s3.example.com&s3_region=r",
		"s3://bucket?host=http://127.0.0.1:9000&s3_region=r", "s3://bucket?host=http://169.254.169.254&s3_region=r", "s3://bucket?host=http://100.100.100.200&s3_region=r",
		"s3://bucket?host=http://[::1]&s3_region=r", "s3://bucket?host=http://localhost&s3_region=r", "s3://bucket?host=http://s3.example.com/api&s3_region=r",
		"s3://bucket/../other?host=http://s3.example.com&s3_region=r", "s3://bucket/%2e%2e/other?host=http://s3.example.com&s3_region=r",
		"s3://bucket?host=http://s3.example.com:99999&s3_region=r", "s3://bucket?host=http://s3.example.com", "s3://bucket?host=http://s3.example.com&s3_region=r;bad",
	} {
		t.Run(raw, func(t *testing.T) {
			if _, err := ParseLocation(raw); err == nil || strings.Contains(err.Error(), "TOPSECRET") {
				t.Fatalf("expected sanitized validation failure, got %v", err)
			}
		})
	}
}

func TestCredentialCreateListAndNoOverwrite(t *testing.T) {
	ctx := context.Background()
	core := fake.NewSimpleClientset(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test"}}, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "unrelated", Namespace: "test"}, Data: map[string][]byte{"password": []byte("OTHERSECRET")}})
	s := &Service{Core: core}
	ref, err := s.Create(ctx, "test", "s3-credentials", "ACCESS_ID", "TOPSECRET")
	if err != nil || ref.Name != "s3-credentials" {
		t.Fatal(ref, err)
	}
	if _, err := s.Create(ctx, "test", "s3-credentials", "NEW_ID", "NEW_SECRET"); err == nil {
		t.Fatal("overwriting an existing secret must fail")
	}
	secret, _ := core.CoreV1().Secrets("test").Get(ctx, "s3-credentials", metav1.GetOptions{})
	if string(secret.Data["access_key"]) != "TOPSECRET" || secret.Immutable == nil || !*secret.Immutable {
		t.Fatal("existing credentials changed or not immutable")
	}
	refs, err := s.List(ctx, "test")
	if err != nil || len(refs) != 1 {
		t.Fatal(refs, err)
	}
	b, _ := json.Marshal(refs)
	for _, sensitive := range []string{"TOPSECRET", "ACCESS_ID", "access_key", "OTHERSECRET"} {
		if strings.Contains(string(b), sensitive) {
			t.Fatal("credential leaked")
		}
	}
	if _, err := s.Create(ctx, "absent", "s3", "id", "key"); err == nil {
		t.Fatal("must use existing namespace")
	}
	if _, err := s.List(ctx, ""); err == nil {
		t.Fatal("must not list secrets across namespaces")
	}
	if refs, err := s.List(ctx, "other"); err != nil || len(refs) != 0 {
		t.Fatal("cross-namespace credential leak")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestBucketCheckReadOnlySignedAndSanitized(t *testing.T) {
	loc := &Location{Endpoint: "https://s3.example.com", Bucket: "bucket", Prefix: "private-prefix", Region: "us-east-1"}
	for _, status := range []int{200, 204, 301, 400, 401, 403, 404, 500} {
		client := safeHTTPClient()
		calls := 0
		client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			calls++
			if req.Method != "HEAD" || req.URL.Path != "/bucket" || req.Body != nil || !strings.HasPrefix(req.Header.Get("Authorization"), "AWS4-HMAC-SHA256 Credential=TEST_ID/") {
				t.Fatal("check must be a signed HEAD without object operations", req.Method, req.URL.Path)
			}
			return &http.Response{StatusCode: status, Header: http.Header{"Location": []string{"https://another.example.com/TOPSECRET"}}, Body: io.NopCloser(strings.NewReader("TOPSECRET")), Request: req}, nil
		})
		r, err := checkBucket(context.Background(), loc, "TEST_ID", "TOPSECRET", client)
		if err != nil || r.OK != (status == 200) || calls != 1 || r.HTTPStatus != status {
			t.Fatal(r, err, calls)
		}
		body, _ := json.Marshal(r)
		if strings.Contains(string(body), "TOPSECRET") || strings.Contains(string(body), "TEST_ID") {
			t.Fatal("credentials/provider response leaked")
		}
	}
}

func TestBucketCheckUsesProviderAddressingStyle(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		bucket   string
		wantURL  string
	}{
		{name: "aliyun oss", endpoint: "https://oss-cn-wulanchabu.aliyuncs.com", bucket: "data-bucket", wantURL: "https://data-bucket.oss-cn-wulanchabu.aliyuncs.com/"},
		{name: "huawei obs", endpoint: "https://obs.cn-north-4.myhuaweicloud.com", bucket: "data-bucket", wantURL: "https://data-bucket.obs.cn-north-4.myhuaweicloud.com/"},
		{name: "huawei obs eu", endpoint: "https://obs.eu-west-101.myhuaweicloud.eu", bucket: "data-bucket", wantURL: "https://data-bucket.obs.eu-west-101.myhuaweicloud.eu/"},
		{name: "huawei dotted bucket over http", endpoint: "http://obs.cn-north-4.myhuaweicloud.com", bucket: "logs.prod", wantURL: "http://logs.prod.obs.cn-north-4.myhuaweicloud.com/"},
		{name: "tencent cos", endpoint: "https://cos.ap-shanghai.myqcloud.com", bucket: "data-bucket-123456", wantURL: "https://data-bucket-123456.cos.ap-shanghai.myqcloud.com/"},
		{name: "aws s3", endpoint: "https://s3.us-east-1.amazonaws.com", bucket: "data-bucket", wantURL: "https://data-bucket.s3.us-east-1.amazonaws.com/"},
		{name: "aws dotted bucket keeps path style", endpoint: "https://s3.us-east-1.amazonaws.com", bucket: "logs.prod", wantURL: "https://s3.us-east-1.amazonaws.com/logs.prod"},
		{name: "baidu bos", endpoint: "https://s3.bj.bcebos.com", bucket: "data-bucket", wantURL: "https://data-bucket.s3.bj.bcebos.com/"},
		{name: "gcp path style remains compatible", endpoint: "https://storage.googleapis.com", bucket: "data-bucket", wantURL: "https://storage.googleapis.com/data-bucket"},
		{name: "custom s3 path style remains compatible", endpoint: "https://s3.example.com", bucket: "data-bucket", wantURL: "https://s3.example.com/data-bucket"},
		{name: "already virtual hosted", endpoint: "https://data-bucket.oss-cn-wulanchabu.aliyuncs.com", bucket: "data-bucket", wantURL: "https://data-bucket.oss-cn-wulanchabu.aliyuncs.com/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := &Location{Endpoint: tt.endpoint, Bucket: tt.bucket, Region: "test-region"}
			calls := 0
			client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				calls++
				if req.Method != http.MethodHead || req.URL.String() != tt.wantURL {
					t.Fatalf("unexpected bucket check request: %s %s, want HEAD %s", req.Method, req.URL, tt.wantURL)
				}
				if !strings.HasPrefix(req.Header.Get("Authorization"), "AWS4-HMAC-SHA256 Credential=TEST_ID/") {
					t.Fatal("bucket check must remain signed")
				}
				return &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("")), Request: req}, nil
			})}
			result, err := checkBucket(context.Background(), loc, "TEST_ID", "TEST_KEY", client)
			if err != nil || result == nil || !result.OK || calls != 1 {
				t.Fatalf("result=%+v err=%v calls=%d", result, err, calls)
			}
		})
	}
}

func TestBucketCheckRejectsHuaweiDottedBucketOverHTTPS(t *testing.T) {
	for _, endpoint := range []string{
		"https://obs.cn-north-4.myhuaweicloud.com",
		"https://logs.prod.obs.cn-north-4.myhuaweicloud.com",
	} {
		t.Run(endpoint, func(t *testing.T) {
			called := false
			client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				called = true
				return &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("")), Request: req}, nil
			})}
			_, err := checkBucket(context.Background(), &Location{Endpoint: endpoint, Bucket: "logs.prod", Region: "cn-north-4"}, "TEST_ID", "TEST_KEY", client)
			if err == nil || called {
				t.Fatalf("Huawei dotted bucket over HTTPS must be rejected before sending a request: err=%v called=%v", err, called)
			}
		})
	}
}

func TestEndpointAddressRestrictions(t *testing.T) {
	for _, address := range []string{"127.0.0.1", "::1", "::ffff:127.0.0.1", "169.254.169.254", "100.100.100.200", "0.0.0.0", "0.0.0.1", "224.0.0.1", "fe80::1", "::"} {
		if allowedIP(net.ParseIP(address)) {
			t.Error("unsafe address allowed:", address)
		}
	}
	for _, address := range []string{"192.0.2.10", "10.42.0.2", "192.168.1.10", "172.20.0.1", "fd00::12"} {
		if !allowedIP(net.ParseIP(address)) {
			t.Error("private MinIO address disallowed:", address)
		}
	}
	client := safeHTTPClient()
	if _, err := client.Get("http://localhost:9000"); err == nil {
		t.Fatal("DNS-resolved loopback must be blocked")
	}
}

func TestCheckRejectsMissingCredentialsBeforeNetwork(t *testing.T) {
	s := &Service{Core: fake.NewSimpleClientset()}
	if _, err := s.Check(context.Background(), "test", CheckRequest{BucketURL: "s3://bucket?host=http://192.0.2.10:9000&s3_region=r", SecretName: "missing"}); err == nil {
		t.Fatal("expected missing credential error")
	}
}
