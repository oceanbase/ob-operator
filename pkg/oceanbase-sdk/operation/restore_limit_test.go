package operation

import (
	"context"
	"testing"
)

func TestRestoreUntilArgument(t *testing.T) {
	for _, tc := range []struct {
		name string
		key  string
		in   any
		want any
	}{
		{"SCN from CR", "SCN", "1795125885051733109", uint64(1795125885051733109)},
		{"numeric SDK input", "SCN", int64(1795125885051733109), uint64(1795125885051733109)},
		{"time", "TIME", "2026-11-20 06:04:45", "2026-11-20 06:04:45"},
		{"negative", "SCN", -1, nil},
		{"fractional", "SCN", "1.5", nil},
		{"overflow", "SCN", "18446744073709551616", nil},
		{"empty", "SCN", "", nil},
		{"unrecognized limit", "OTHER", "42", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := RestoreUntilArgument(tc.key, tc.in)
			if tc.want == nil {
				if err == nil {
					t.Fatalf("expected invalid restore limit to be rejected, got %v", got)
				}
			} else if err != nil || got != tc.want {
				t.Fatalf("restore argument = %T(%v), %v; want %T(%v)", got, got, err, tc.want, tc.want)
			}
		})
	}
}

func TestLimitedRestoreRejectsMissingKey(t *testing.T) {
	// An invalid request must be rejected before any connection is used.
	var db *OceanbaseOperationManager
	err := db.StartRestoreWithLimit(context.Background(), "target", "uri", "pool_list=p", "", nil)
	if err == nil {
		t.Fatal("limited restore must not submit an unlimited restore")
	}
}
