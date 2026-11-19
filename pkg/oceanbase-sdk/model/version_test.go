package model

import "testing"

func TestObserverBuildVersion(t *testing.T) {
	v, err := ParseOBVersion("4.2.1.3_103000032023122818-8fe69c20")
	if err != nil || v.Major != 4 || v.Minor != 2 || v.Patch != 1 || v.Incr != 3 || v.String() != "4.2.1.3-103000032023122818" {
		t.Fatalf("unexpected observer version: %+v, %v", v, err)
	}
	for _, raw := range []string{"4.3.1.3", "4.2.2.3", "4.2.1.7"} {
		newer, err := ParseOBVersion(raw)
		if err != nil || v.Compare(newer) >= 0 {
			t.Fatalf("%s should precede %s: %v", v, raw, err)
		}
	}
	for _, raw := range []string{"", "4.2", "4.x.1", "4.2.x", "4.2.1.x"} {
		if v, err := ParseOBVersion(raw); err == nil || v != nil {
			t.Fatalf("invalid observer version %q produced %#v, %v", raw, v, err)
		}
	}
}
