/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package types

import (
	"fmt"
	"net/url"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

type SharedStorageSpec struct {
	BucketURL    string                      `json:"bucketURL"`
	SecretRef    corev1.LocalObjectReference `json:"secretRef"`
	MaxIOPS      string                      `json:"maxIOPS,omitempty"`
	MaxBandwidth string                      `json:"maxBandwidth,omitempty"`
}

type LogServiceReference struct {
	Name string `json:"name"`
}

// ValidateSharedStorageBucketURL checks the storage address without echoing it:
// URLs may contain sensitive query parameters supplied by existing clients.
func ValidateSharedStorageBucketURL(raw string) error {
	invalid := func() error {
		return fmt.Errorf("bucketURL must contain a supported storage scheme, bucket and one valid host endpoint")
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "s3" && u.Scheme != "oss" && u.Scheme != "cos") || u.Host == "" || u.User != nil || u.Fragment != "" || strings.ContainsAny(u.Host, ": ") {
		return invalid()
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil || len(q["host"]) != 1 || q.Get("host") == "" {
		return invalid()
	}
	endpoint := q.Get("host")
	if !strings.Contains(endpoint, "://") {
		endpoint = "https://" + endpoint
	}
	e, err := url.Parse(endpoint)
	if err != nil || (e.Scheme != "http" && e.Scheme != "https") || e.Hostname() == "" || e.User != nil || e.RawQuery != "" || e.Fragment != "" {
		return invalid()
	}
	return nil
}
