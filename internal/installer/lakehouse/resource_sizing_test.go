package lakehouse

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
)

func sizingTestValues() map[string]interface{} {
	return map[string]interface{}{
		"observer": map[string]interface{}{
			"resource":   map[string]interface{}{"cpu": "2", "memory": "10Gi"},
			"dataSize":   "50Gi",
			"logSize":    "20Gi",
			"parameters": []interface{}{map[string]interface{}{"name": "system_memory", "value": "1G"}},
		},
		"logService": map[string]interface{}{
			"resource":   map[string]interface{}{"cpu": "1", "memory": "4Gi"},
			"storeSize":  "1Mi",
			"logSize":    "1Mi",
			"parameters": []interface{}{map[string]interface{}{"name": "log_level", "value": "INFO"}},
		},
	}
}

func TestQuickstartSizingMemoryAndDisk(t *testing.T) {
	for _, tc := range []struct {
		name, memory, data, log, errorPart string
	}{
		{"defaults", "10Gi", "50Gi", "20Gi", ""},
		{"sixteen GiB", "16Gi", "50Gi", "20Gi", ""},
		{"twenty GiB default disk", "20Gi", "50Gi", "20Gi", "increase --observer-data-size or reduce --observer-memory"},
		{"twenty GiB larger disk", "20Gi", "100Gi", "20Gi", ""},
		{"minimums", "8Gi", "30Gi", "10Gi", ""},
		{"memory below minimum", "8191Mi", "50Gi", "20Gi", "observer.resource.memory must be at least 8Gi"},
		{"data below minimum", "8Gi", "29Gi", "20Gi", "observer.dataSize must be at least 30Gi"},
		{"log below minimum", "8Gi", "50Gi", "9Gi", "observer.logSize must be at least 10Gi"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := sizingTestValues()
			setValue(values, tc.memory, "observer", "resource", "memory")
			setValue(values, tc.data, "observer", "dataSize")
			setValue(values, tc.log, "observer", "logSize")
			err := validateQuickstartSizing(values)
			if tc.errorPart == "" && err != nil || tc.errorPart != "" && (err == nil || !strings.Contains(err.Error(), tc.errorPart)) {
				t.Fatalf("unexpected validation result: %v (want %q)", err, tc.errorPart)
			}
		})
	}
}

func TestQuickstartSizingExplicitMemoryLimit(t *testing.T) {
	for _, tc := range []struct {
		name, component, memory, data string
		limits                        []interface{}
		errorPart                     string
	}{
		{"smaller explicit limit", "observer", "20Gi", "50Gi", []interface{}{"10Gi"}, ""},
		{"exact three times", "observer", "20Gi", "60Gi", []interface{}{"20Gi"}, ""},
		{"below three times", "observer", "20Gi", "59Gi", []interface{}{"20Gi"}, "three times"},
		{"last wins", "observer", "20Gi", "50Gi", []interface{}{"99Gi", "10Gi"}, ""},
		{"invalid overridden value", "observer", "20Gi", "50Gi", []interface{}{42, "10Gi"}, ""},
		{"last exceeds memory", "observer", "20Gi", "100Gi", []interface{}{"10Gi", "21Gi"}, "exceeds observer.resource.memory"},
		{"logservice limit", "logService", "4Gi", "50Gi", []interface{}{"4Gi"}, ""},
		{"logservice excessive limit", "logService", "4Gi", "50Gi", []interface{}{"5Gi"}, "exceeds logService.resource.memory"},
		{"invalid unit", "observer", "20Gi", "100Gi", []interface{}{"10GB"}, "positive Kubernetes quantity"},
		{"empty", "observer", "20Gi", "100Gi", []interface{}{""}, "positive Kubernetes quantity"},
		{"zero", "observer", "20Gi", "100Gi", []interface{}{"0"}, "positive Kubernetes quantity"},
		{"negative", "logService", "4Gi", "50Gi", []interface{}{"-1Gi"}, "positive Kubernetes quantity"},
		{"non string", "logService", "4Gi", "50Gi", []interface{}{42}, "positive Kubernetes quantity"},
		{"huge binary limit", "observer", "20Gi", "100Gi", []interface{}{"999999999999999999999999Ti"}, "exceeds observer.resource.memory"},
		{"huge decimal limit", "observer", "20Gi", "100Gi", []interface{}{"999999999999999999999999G"}, "exceeds observer.resource.memory"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := sizingTestValues()
			setValue(values, tc.memory, tc.component, "resource", "memory")
			setValue(values, tc.data, "observer", "dataSize")
			parameters := pfGet(values, tc.component, "parameters").([]interface{})
			for _, limit := range tc.limits {
				parameters = append(parameters, map[string]interface{}{"name": "memory_limit", "value": limit})
			}
			setValue(values, parameters, tc.component, "parameters")
			before, err := json.Marshal(values)
			if err != nil {
				t.Fatal(err)
			}
			err = validateQuickstartSizing(values)
			if tc.errorPart == "" && err != nil || tc.errorPart != "" && (err == nil || !strings.Contains(err.Error(), tc.errorPart)) {
				t.Fatalf("unexpected validation result: %v (want %q)", err, tc.errorPart)
			}
			after, marshalErr := json.Marshal(values)
			if marshalErr != nil || !bytes.Equal(before, after) {
				t.Fatal("sizing validation changed supplied values or parameters")
			}
		})
	}
}

func TestQuickstartSizingInvalidCompleteValues(t *testing.T) {
	for _, tc := range []struct {
		name      string
		path      []string
		value     interface{}
		errorPart string
	}{
		{"missing observer", []string{"observer"}, nil, "observer.resource.memory"},
		{"missing memory", []string{"observer", "resource", "memory"}, nil, "observer.resource.memory"},
		{"missing data", []string{"observer", "dataSize"}, nil, "observer.dataSize"},
		{"numeric size", []string{"observer", "logSize"}, 10240, "observer.logSize"},
		{"unsupported size unit", []string{"observer", "dataSize"}, "100G", "observer.dataSize"},
		{"fractional size", []string{"observer", "resource", "memory"}, "10.5Gi", "observer.resource.memory"},
		{"memory overflow", []string{"observer", "resource", "memory"}, "8388608Ti", "int64 byte capacity"},
		{"data overflow", []string{"observer", "dataSize"}, "999999999999999999999999Ti", "int64 byte capacity"},
		{"logservice memory overflow", []string{"logService", "resource", "memory"}, "8796093022208Mi", "int64 byte capacity"},
		{"missing logservice", []string{"logService"}, nil, "logService.resource.memory"},
		{"parameters not list", []string{"observer", "parameters"}, "bad", "observer.parameters must be a parameter list"},
		{"parameter not object", []string{"observer", "parameters"}, []interface{}{"bad"}, "parameters[0] must be a parameter object"},
		{"parameter missing name", []string{"observer", "parameters"}, []interface{}{map[string]interface{}{"value": "10Gi"}}, "parameters[0].name"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := sizingTestValues()
			setValue(values, tc.value, tc.path...)
			if err := validateQuickstartSizing(values); err == nil || !strings.Contains(err.Error(), tc.errorPart) {
				t.Fatalf("unexpected validation result: %v (want %q)", err, tc.errorPart)
			}
		})
	}
	if err := validateQuickstartSizing(nil); err == nil {
		t.Fatal("nil values must return a clear error")
	}
}

func TestQuickstartSizingDefaultRoundingAndOverflow(t *testing.T) {
	// Multiplication by 90 exceeds int64 for the large cases; the default and
	// three-times comparison must still be exact and positive.
	for _, tc := range []struct{ memory, limit string }{
		{"10Gi", "9216M"},
		{"8193Mi", "7373M"},
		{"100000Ti", "94371840000M"},
		{"8388607Ti", "7916482776268M"},
	} {
		t.Run(tc.memory, func(t *testing.T) {
			values := sizingTestValues()
			setValue(values, tc.memory, "observer", "resource", "memory")
			memory, err := quickstartSize(values, "observer", "resource", "memory")
			if err != nil {
				t.Fatal(err)
			}
			limit, err := quickstartMemoryLimit(values, "observer", memory, true)
			if err != nil || limit.Cmp(resource.MustParse(tc.limit)) != 0 {
				t.Fatalf("default limit=%s err=%v, want %s", limit.String(), err, tc.limit)
			}
		})
	}
	values := sizingTestValues()
	setValue(values, "100000Ti", "observer", "resource", "memory")
	setValue(values, "300000Ti", "observer", "dataSize")
	if err := validateQuickstartSizing(values); err != nil {
		t.Fatalf("large, non-overflowing capacities should remain comparable: %v", err)
	}
	setValue(values, "8388607Ti", "observer", "resource", "memory")
	setValue(values, "8388607Ti", "observer", "dataSize")
	setValue(values, []interface{}{map[string]interface{}{"name": "memory_limit", "value": "8388607Ti"}}, "observer", "parameters")
	if err := validateQuickstartSizing(values); err == nil || !strings.Contains(err.Error(), "three times") {
		t.Fatalf("three-times arithmetic must not overflow into accepting an undersized disk: %v", err)
	}
}
