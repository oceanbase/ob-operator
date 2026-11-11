package obzone

import (
	"testing"

	apitypes "github.com/oceanbase/ob-operator/api/types"
)

func TestSharedStorageAttribute(t *testing.T) {
	for _, tc := range []struct{ name, iops, bandwidth, want string }{
		{"defaults", "", "", "max_iops=0&max_bandwidth=0B"},
		{"iops only", "1000", "", "max_iops=1000&max_bandwidth=0B"},
		{"bandwidth only", "", "1GB", "max_iops=0&max_bandwidth=1GB"},
		{"numeric bytes", "20", "1024", "max_iops=20&max_bandwidth=1024B"},
		{"zero bytes", "0", "0", "max_iops=0&max_bandwidth=0B"},
		{"whitespace", " ", " ", "max_iops=0&max_bandwidth=0B"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := buildSharedStorageAttribute(tc.iops, tc.bandwidth); got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}

func TestSharedStorageAttributeURLFallback(t *testing.T) {
	s := &apitypes.SharedStorageSpec{BucketURL: "s3://bucket?host=http://minio:9000&max_iops=3000&max_bandwidth=2GB"}
	got, err := sharedStorageAttributeFromSpec(s)
	if err != nil || got != "max_iops=3000&max_bandwidth=2GB" {
		t.Fatalf("%s %v", got, err)
	}
	s.MaxIOPS = "1000"
	got, err = sharedStorageAttributeFromSpec(s)
	if err != nil || got != "max_iops=1000&max_bandwidth=2GB" {
		t.Fatalf("%s %v", got, err)
	}
}
