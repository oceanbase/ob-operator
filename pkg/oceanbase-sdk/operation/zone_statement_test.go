package operation

import (
	"reflect"
	"strings"
	"testing"
)

func TestSharedStorageDestRequiredAttributeDefault(t *testing.T) {
	for _, attribute := range []string{"", "  ", "max_iops=1000", "max_iops=1000&max_bandwidth=1GB"} {
		t.Run(attribute, func(t *testing.T) {
			query, args := sharedStorageDestStatement("s3://bucket", "test-only-credential", attribute, "zone4")
			if !strings.Contains(query, " attribute = ?") {
				t.Fatalf("unexpected optional clause: %s", query)
			}
			if strings.TrimSpace(attribute) == "" {
				attribute = "max_iops=0&max_bandwidth=0B"
			}
			want := []any{"s3://bucket", "test-only-credential", attribute, "zone4"}
			if !reflect.DeepEqual(args, want) || strings.Count(query, "?") != len(args) {
				t.Fatal("incorrect parameter binding")
			}
			if strings.Contains(query, "test-only-credential") {
				t.Fatal("credential must remain a bound parameter")
			}
		})
	}
}
