/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package common

import (
	"net"
	"net/url"
	"regexp"
	"strings"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

// ObjectReference deliberately contains no Secret data.
type ObjectReference struct {
	Name string `json:"name" binding:"required"`
}

type SharedStorageSpec struct {
	BucketURL    string          `json:"bucketURL" binding:"required"`
	SecretRef    ObjectReference `json:"secretRef" binding:"required"`
	MaxIOPS      string          `json:"maxIOPS,omitempty"`
	MaxBandwidth string          `json:"maxBandwidth,omitempty"`
}

func (s *SharedStorageSpec) ToAPI() *apitypes.SharedStorageSpec {
	if s == nil {
		return nil
	}
	return &apitypes.SharedStorageSpec{BucketURL: s.BucketURL, SecretRef: corev1.LocalObjectReference{Name: s.SecretRef.Name}, MaxIOPS: s.MaxIOPS, MaxBandwidth: s.MaxBandwidth}
}

func SharedStorageFromAPI(s *apitypes.SharedStorageSpec) *SharedStorageSpec {
	if s == nil {
		return nil
	}
	bucketURL, _ := PublicBucketURL(s.BucketURL)
	return &SharedStorageSpec{BucketURL: bucketURL, SecretRef: ObjectReference{Name: s.SecretRef.Name}, MaxIOPS: s.MaxIOPS, MaxBandwidth: s.MaxBandwidth}
}

// PublicBucketURL returns the storage location that a Dashboard reader may see.
// The second result reports whether the input was safe to accept on creation.
// Preserve the original encoding of safe values because OceanBase also accepts
// unescaped endpoint schemes in this query string.
func PublicBucketURL(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil || len(validation.IsDNS1123Subdomain(strings.ToLower(u.Host))) != 0 || (u.Scheme != "s3" && u.Scheme != "oss" && u.Scheme != "cos") {
		return "<invalid URL>", false
	}
	safe := u.User == nil && u.Fragment == ""
	u.User, u.Fragment, u.RawFragment = nil, "", ""
	var parts []string
	if u.RawQuery != "" {
		for _, part := range strings.Split(u.RawQuery, "&") {
			key, value, found := strings.Cut(part, "=")
			decoded, err := url.QueryUnescape(value)
			if !found || err != nil || !publicBucketOption(key, decoded) {
				safe = false
				continue
			}
			parts = append(parts, part)
		}
	}
	u.RawQuery = strings.Join(parts, "&")
	return u.String(), safe
}

var storageOptionValue = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

func publicBucketOption(key, value string) bool {
	switch key {
	case "host", "endpoint":
		if !strings.Contains(value, "://") {
			value = "https://" + value
		}
		endpoint, err := url.Parse(value)
		return err == nil && (endpoint.Scheme == "https" || endpoint.Scheme == "http") &&
			(net.ParseIP(endpoint.Hostname()) != nil || len(validation.IsDNS1123Subdomain(strings.TrimSuffix(strings.ToLower(endpoint.Hostname()), "."))) == 0) &&
			endpoint.User == nil && endpoint.RawQuery == "" &&
			endpoint.Fragment == "" && (endpoint.Path == "" || endpoint.Path == "/")
	case "s3_region", "scope", "max_iops", "max_bandwidth":
		return storageOptionValue.MatchString(value)
	default:
		return false
	}
}

func (r *ObjectReference) ToLogServiceRef() *apitypes.LogServiceReference {
	if r == nil {
		return nil
	}
	return &apitypes.LogServiceReference{Name: r.Name}
}

func LogServiceRefFromAPI(r *apitypes.LogServiceReference) *ObjectReference {
	if r == nil {
		return nil
	}
	return &ObjectReference{Name: r.Name}
}
