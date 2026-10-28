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

package main

import (
	"errors"
	"strings"
	"testing"
)

func TestBuildBootstrapSQL(t *testing.T) {
	cfg := config{
		ObserverHost:           "172.30.52.12",
		RPCPort:                2882,
		Zone:                   "zone1",
		LogServiceBucketURL:    "s3://lakehouse-logservice?host=http://minio:9000&s3_region=us-east-1",
		SharedStorageBucketURL: "s3://lakehouse-data?host=http://minio:9000&s3_region=us-east-1",
		LogServiceClusterID:    1001,
	}
	sql, err := buildBootstrapSQL(cfg, "access-id", "access-key")
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"ZONE 'zone1' SERVER '172.30.52.12:2882'",
		"s3://lakehouse-logservice?host=http://minio:9000&s3_region=us-east-1&access_id=access-id&access_key=access-key&logservice_cluster_id=1001",
		"s3://lakehouse-data?host=http://minio:9000&s3_region=us-east-1&access_id=access-id&access_key=access-key",
	} {
		if !strings.Contains(sql, expected) {
			t.Fatalf("bootstrap SQL %q does not contain %q", sql, expected)
		}
	}
}

func TestBuildBootstrapSQLRejectsUnsafeValues(t *testing.T) {
	cfg := config{
		ObserverHost:           "observer",
		RPCPort:                2882,
		Zone:                   "zone1' OR 1=1",
		LogServiceBucketURL:    "s3://logs?host=http://minio:9000",
		SharedStorageBucketURL: "s3://data?host=http://minio:9000",
		LogServiceClusterID:    1001,
	}
	if _, err := buildBootstrapSQL(cfg, "access", "secret"); err == nil {
		t.Fatal("expected unsafe zone to be rejected")
	}
}

func TestLoadConfigRequiresStorageAndClusterID(t *testing.T) {
	env := map[string]string{
		"LOGSERVICE_CLUSTER_ID":     "1001",
		"LOGSERVICE_BUCKET_URL":     "s3://logs?host=http://minio:9000",
		"SHARED_STORAGE_BUCKET_URL": "s3://data?host=http://minio:9000",
	}
	cfg, err := loadConfig(func(name string) string { return env[name] })
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ObserverHost != "observer" || cfg.SQLPort != 2881 || cfg.RPCPort != 2882 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestBootstrapInProgress(t *testing.T) {
	if !bootstrapInProgress(errors.New("Access denied; you need the BOOTSTRAP privilege")) {
		t.Fatal("expected bootstrap privilege error to be treated as in progress")
	}
	if bootstrapInProgress(errors.New("connection refused")) {
		t.Fatal("unexpected transient connection error classification")
	}
}
