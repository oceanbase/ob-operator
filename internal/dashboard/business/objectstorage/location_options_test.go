package objectstorage

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestLocationOceanBaseOptions(t *testing.T) {
	const base = "s3://data-bucket/?host=https://s3.example.com&s3_region=cn-wulanchabu"
	for _, options := range []string{
		"&scope=region1&max_iops=10000&max_bandwidth=1GB",
		"&scope=region1", "&max_iops=0", "&max_bandwidth=0B", "&max_bandwidth=1024", "&max_bandwidth=1.5GB",
	} {
		t.Run(options, func(t *testing.T) {
			loc, err := ParseLocation(base + options)
			if err != nil {
				t.Fatal(err)
			}
			if loc.Endpoint != "https://s3.example.com" || loc.Bucket != "data-bucket" || loc.Region != "cn-wulanchabu" || loc.Prefix != "" {
				t.Fatalf("unexpected location: %+v", loc)
			}
			calls := 0
			client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				calls++
				if req.Method != "HEAD" || req.URL.String() != "https://s3.example.com/data-bucket" {
					t.Fatalf("options leaked into bucket check: %s %s", req.Method, req.URL)
				}
				if !strings.Contains(req.Header.Get("Authorization"), "/cn-wulanchabu/s3/aws4_request") {
					t.Fatal("wrong signing region")
				}
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("")), Request: req}, nil
			})}
			result, err := checkBucket(context.Background(), loc, "TEST_ID", "TEST_KEY", client)
			if err != nil || !result.OK || calls != 1 {
				t.Fatalf("result=%+v err=%v calls=%d", result, err, calls)
			}
		})
	}
}

func TestLocationOptionsRejectUnsafeValues(t *testing.T) {
	const base = "s3://data-bucket/?host=https://s3.example.com&s3_region=cn-wulanchabu"
	for _, option := range []string{
		"scope=region1%0A", "max_iops=10000%0A", "max_bandwidth=1GB%0A",
		"access_key=TOPSECRET", "access_id=TOPSECRET", "unknown=x", "scope=region1&scope=region2", "scope=", "scope=region%26access_key%3DTOPSECRET",
		"max_iops=-1", "max_iops=1.2", "max_iops=1&max_iops=2", "max_bandwidth=", "max_bandwidth=fast", "max_bandwidth=1GB&max_bandwidth=2GB",
		"scope=region1&max_iops=10000&max_bandwidth=1GB&access_key=TOPSECRET",
	} {
		t.Run(option, func(t *testing.T) {
			if _, err := ParseLocation(base + "&" + option); err == nil || strings.Contains(err.Error(), "TOPSECRET") {
				t.Fatalf("expected sanitized rejection: %v", err)
			}
		})
	}
}
