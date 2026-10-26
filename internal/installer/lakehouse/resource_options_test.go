package lakehouse

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestSizingFlagQuantityValidation(t *testing.T) {
	for _, tc := range []struct{ name, value, want string }{
		{"observer-cpu", "2", "2"},
		{"observer-cpu", "1.5", "1500m"},
		{"logservice-cpu", "0.5", "500m"},
		{"observer-cpu", "0.001", "1m"},
		{"logservice-cpu", "500m", "500m"},
		{"observer-cpu", "1000", "1000"},
		{"observer-cpu", "1000000000", "1000000000"},
		{"observer-memory", "16Gi", "16Gi"},
		{"observer-data-size", "1Ti", "1Ti"},
		{"logservice-log-size", "512Mi", "512Mi"},
		{"minio-size", "8388607Ti", "8388607Ti"},
	} {
		t.Run(tc.name+"/"+tc.value, func(t *testing.T) {
			o, err := parseOptions([]string{"--" + tc.name, tc.value}, io.Discard)
			if err != nil || o.resources[tc.name] != tc.want {
				t.Fatalf("value=%q err=%v; want %q", o.resources[tc.name], err, tc.want)
			}
		})
	}
	for _, name := range []string{"observer-cpu", "logservice-cpu"} {
		for _, value := range []string{"", "0", "0m", "0.0001", "1.0000", "-1", "+1", "1e3", "1Gi", "1.5m", " 2", "2\n", "1000000001", "999999999999999999999999999999"} {
			if _, err := parseOptions([]string{"--" + name + "=" + value}, io.Discard); err == nil {
				t.Errorf("%s accepted invalid quantity %q", name, value)
			}
		}
	}
	for _, name := range []string{"observer-memory", "observer-data-size", "observer-log-size", "logservice-memory", "logservice-store-size", "logservice-log-size", "minio-size"} {
		for _, value := range []string{"", "0Gi", "-1Gi", "1.5Gi", "1G", "100", "1GB", "1gi", "1e3", "1Gi\n", " 1Gi", "8388608Ti", "8589934592Gi", "8796093022208Mi"} {
			if _, err := parseOptions([]string{"--" + name + "=" + value}, io.Discard); err == nil {
				t.Errorf("%s accepted invalid quantity %q", name, value)
			}
		}
	}
}

func TestSizingFlagsHelpAndExplicitness(t *testing.T) {
	var help bytes.Buffer
	if _, err := parseOptions([]string{"--help"}, &help); !errors.Is(err, flag.ErrHelp) {
		t.Fatal("help failed", err)
	}
	for _, option := range resourceOptions {
		if !strings.Contains(help.String(), "-"+option.name) {
			t.Errorf("help omitted %s", option.name)
		}
	}
	o, err := parseOptions([]string{"-observer-cpu", "1", "--observer-cpu=1.5"}, io.Discard)
	if err != nil || !reflect.DeepEqual(o.resources, resourceOverrides{"observer-cpu": "1500m"}) {
		t.Fatal("only explicit options should be stored and the last occurrence should win", o.resources, err)
	}
	values := sizingTestValues()
	before := mergeValues(values, nil)
	result := o.resources.apply(values)
	if pfGet(result, "observer", "resource", "cpu") != "1500m" || pfGet(result, "observer", "resource", "memory") != "10Gi" || !reflect.DeepEqual(values, before) {
		t.Fatal("applying CPU override changed unspecified settings or the original map")
	}
	o, err = parseOptions(nil, io.Discard)
	if err != nil || len(o.resources) != 0 || !reflect.DeepEqual(o.resources.apply(values), values) {
		t.Fatal("omitted sizing flags must preserve configuration", err)
	}
}

func TestSizingFlagsRejectChangingAnyExistingResource(t *testing.T) {
	installed := sizingTestValues()
	setValue(installed, "100Gi", "minio", "size")
	setValue(installed, "bundled", "objectStorage", "mode")
	for _, option := range resourceOptions {
		t.Run(option.name, func(t *testing.T) {
			value := "120Gi"
			if option.cpu {
				value = "3"
			}
			overrides := resourceOverrides{option.name: value}
			values := overrides.apply(installed)
			err := overrides.validateContext(values, installed, true)
			if err == nil || !strings.Contains(err.Error(), "--"+option.name+" changes the sizing of an existing release") {
				t.Fatal("existing-resource changes must be rejected before installation", err)
			}
		})
	}
	overrides := resourceOverrides{"observer-cpu": "2000m", "minio-size": "102400Mi"}
	values := overrides.apply(installed)
	if err := overrides.validateContext(values, installed, true); err != nil {
		t.Fatal("equivalent quantities must support retrying an installation", err)
	}
	reused := overrides.reuseInstalled(values, installed)
	if pfGet(reused, "minio", "size") != "100Gi" || pfGet(reused, "observer", "resource", "cpu") != "2" || pfGet(values, "minio", "size") != "102400Mi" {
		t.Fatal("replaying equivalent sizes must preserve the installed representation without mutating input")
	}
}

func TestSizingFlagsRejectIgnoredMinIOCapacity(t *testing.T) {
	for _, existing := range []bool{false, true} {
		for _, external := range []bool{false, true} {
			values := sizingTestValues()
			setValue(values, "100Gi", "minio", "size")
			setValue(values, "bundled", "objectStorage", "mode")
			want := "existingClaim"
			if external {
				setValue(values, "external", "objectStorage", "mode")
				want = "external storage"
			} else {
				setValue(values, "existing-data", "minio", "existingClaim")
			}
			if err := (resourceOverrides{"minio-size": "100Gi"}).validateContext(values, values, existing); err == nil || !strings.Contains(err.Error(), want) {
				t.Fatal("MinIO capacity flag must not be silently ignored", err)
			}
		}
	}
}
