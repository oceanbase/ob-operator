package lakehouse

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The public alias contract is deliberately independent of resourceOptions so
// an accidental alias-to-field swap in registration cannot change the oracle.
var resourceAliasContract = []struct {
	long, short, input, normalized string
	path                           []string
}{
	{"observer-cpu", "oc", "1.5", "1500m", []string{"observer", "resource", "cpu"}},
	{"observer-memory", "om", "16Gi", "16Gi", []string{"observer", "resource", "memory"}},
	{"observer-data-size", "ods", "80Gi", "80Gi", []string{"observer", "dataSize"}},
	{"observer-log-size", "ols", "30Gi", "30Gi", []string{"observer", "logSize"}},
	{"logservice-cpu", "lc", "0.5", "500m", []string{"logService", "resource", "cpu"}},
	{"logservice-memory", "lm", "6Gi", "6Gi", []string{"logService", "resource", "memory"}},
	{"logservice-store-size", "lss", "35Gi", "35Gi", []string{"logService", "storeSize"}},
	{"logservice-log-size", "lls", "15Gi", "15Gi", []string{"logService", "logSize"}},
	{"minio-size", "ms", "120Gi", "120Gi", []string{"minio", "size"}},
}

func TestResourceAliasesMapToCanonicalOptionsAndChartFields(t *testing.T) {
	for _, alias := range resourceAliasContract {
		for _, prefix := range []string{"-", "--"} {
			for _, equal := range []bool{false, true} {
				name := prefix + alias.short
				args := []string{name, alias.input}
				if equal {
					name += "="
					args = []string{prefix + alias.short + "=" + alias.input}
				}
				t.Run(name, func(t *testing.T) {
					o, err := parseOptions(args, io.Discard)
					if err != nil {
						t.Fatal(err)
					}
					if len(o.resources) != 1 || o.resources[alias.long] != alias.normalized {
						t.Fatalf("alias must set exactly its canonical option: got=%v want=%s:%s", o.resources, alias.long, alias.normalized)
					}
					if _, exists := o.resources[alias.short]; exists {
						t.Fatal("alias must not create a separate precedence slot")
					}
					values := o.resources.apply(sizingTestValues())
					if got := pfString(pfGet(values, alias.path...)); got != alias.normalized {
						t.Fatal("alias changed the wrong chart field", alias.short, alias.path, got)
					}
				})
			}
		}
	}
}

func TestResourceAliasesAndLongNamesShareLastOccurrencePrecedence(t *testing.T) {
	for _, alias := range resourceAliasContract {
		for _, shortLast := range []bool{false, true} {
			name := alias.short + "-long-last"
			first, last := "-"+alias.short, "--"+alias.long
			if shortLast {
				name = alias.short + "-short-last"
				first, last = "--"+alias.long, "--"+alias.short
			}
			t.Run(name, func(t *testing.T) {
				old := "8Gi"
				if strings.HasSuffix(alias.long, "-cpu") {
					old = "2"
				}
				o, err := parseOptions([]string{first, old, last + "=" + alias.input}, io.Discard)
				if err != nil || len(o.resources) != 1 || o.resources[alias.long] != alias.normalized {
					t.Fatal("long and short names must share one setter; the last occurrence wins", o.resources, err)
				}
			})
		}
	}
}

func TestResourceAliasesDoNotBypassValidationOrShowAccessRestriction(t *testing.T) {
	for _, alias := range resourceAliasContract {
		t.Run(alias.short, func(t *testing.T) {
			if _, err := parseOptions([]string{"-" + alias.short, "0"}, io.Discard); err == nil || !strings.Contains(err.Error(), "--"+alias.long) {
				t.Fatal("short alias bypassed the canonical quantity validator", err)
			}
			if _, err := parseOptions([]string{"--show-access", "--" + alias.short, alias.input}, io.Discard); err == nil || !strings.Contains(err.Error(), "sizing options cannot be used with --show-access") {
				t.Fatal("short alias bypassed the read-only access-mode restriction", err)
			}
		})
	}
}

func TestResourceAliasesAppearAlongsideLongOptionsInHelp(t *testing.T) {
	var output bytes.Buffer
	if _, err := parseOptions([]string{"--help"}, &output); !errors.Is(err, flag.ErrHelp) {
		t.Fatal("help was not available", err)
	}
	for _, alias := range resourceAliasContract {
		for _, name := range []string{alias.long, alias.short} {
			pattern := regexp.MustCompile(`(?m)^\s+-` + regexp.QuoteMeta(name) + `(?:\s|$)`)
			if !pattern.MatchString(output.String()) {
				t.Errorf("help does not list -%s as an available option", name)
			}
		}
	}
}

func TestResourceAliasesKeepExistingReleaseAndMinIOContextGuards(t *testing.T) {
	installed := sizingTestValues()
	setValue(installed, "100Gi", "minio", "size")
	setValue(installed, "bundled", "objectStorage", "mode")
	for _, alias := range resourceAliasContract {
		t.Run("resize-"+alias.short, func(t *testing.T) {
			changed := "120Gi"
			if strings.HasSuffix(alias.long, "-cpu") {
				changed = "3"
			}
			o, err := parseOptions([]string{"-" + alias.short, changed}, io.Discard)
			if err != nil {
				t.Fatal(err)
			}
			if err = o.resources.validateContext(o.resources.apply(installed), installed, true); err == nil || !strings.Contains(err.Error(), "--"+alias.long+" changes the sizing of an existing release") {
				t.Fatal("alias bypassed the existing-resource resize restriction", err)
			}
		})
	}
	o, err := parseOptions([]string{"-ms", "100Gi"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"external", "existingClaim"} {
		t.Run("minio-"+mode, func(t *testing.T) {
			values := o.resources.apply(installed)
			if mode == "external" {
				setValue(values, "external", "objectStorage", "mode")
			} else {
				setValue(values, "retained-minio", "minio", "existingClaim")
			}
			if err := o.resources.validateContext(values, installed, false); err == nil || !strings.Contains(err.Error(), "--minio-size") {
				t.Fatal("MinIO alias bypassed unsupported-storage protection", err)
			}
		})
	}
}

func TestResourceAliasesOverrideValuesFilesInTheRunPath(t *testing.T) {
	for _, aliasFirst := range []bool{false, true} {
		name := "values-first"
		if aliasFirst {
			name = "aliases-first"
		}
		t.Run(name, func(t *testing.T) {
			f := newResourceRunFixture(t)
			path := filepath.Join(t.TempDir(), "values.yaml")
			resourceTestWrite(t, path, "observer:\n  resource: {cpu: '3', memory: 12Gi}\n  dataSize: 60Gi\n  logSize: 21Gi\nlogService:\n  resource: {cpu: '1', memory: 5Gi}\n  storeSize: 22Gi\n  logSize: 11Gi\nminio: {size: 101Gi}\n")
			var aliases []string
			for _, alias := range resourceAliasContract {
				aliases = append(aliases, "-"+alias.short, alias.input)
			}
			args := append([]string{"-f", path}, aliases...)
			if aliasFirst {
				args = append(aliases, "-f", path)
			}
			output, err := f.run(t, args...)
			if err != nil {
				t.Fatalf("alias installation failed: %v\n%s", err, output)
			}
			values := f.persisted(t)
			for _, alias := range resourceAliasContract {
				if got := pfString(pfGet(values, alias.path...)); got != alias.normalized {
					t.Errorf("-%s lost to the values file: got %q want %q", alias.short, got, alias.normalized)
				}
			}
		})
	}
}
