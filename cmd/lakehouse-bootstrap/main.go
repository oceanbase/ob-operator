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
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	defaultTimeout       = 30 * time.Minute
	defaultRetryInterval = 5 * time.Second
)

type config struct {
	ObserverHost           string
	SQLPort                int
	RPCPort                int
	Zone                   string
	LogServiceBucketURL    string
	SharedStorageBucketURL string
	LogServiceClusterID    uint64
	AccessIDFile           string
	AccessKeyFile          string
	Timeout                time.Duration
	RetryInterval          time.Duration
}

func main() {
	cfg, err := loadConfig(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	accessID, err := readCredential(cfg.AccessIDFile)
	if err != nil {
		log.Fatalf("read access ID: %v", err)
	}
	accessKey, err := readCredential(cfg.AccessKeyFile)
	if err != nil {
		log.Fatalf("read access key: %v", err)
	}

	bootstrapSQL, err := buildBootstrapSQL(cfg, accessID, accessKey)
	if err != nil {
		log.Fatal(err)
	}

	dsn := fmt.Sprintf("root@tcp(%s)/?timeout=5s&writeTimeout=30s&interpolateParams=true", net.JoinHostPort(cfg.ObserverHost, strconv.Itoa(cfg.SQLPort)))
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open OceanBase connection: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	if err := waitForConnection(ctx, db, cfg.RetryInterval); err != nil {
		log.Fatal(err)
	}
	if clusterReady(ctx, db) {
		log.Println("OceanBase is already bootstrapped")
		return
	}

	log.Printf("bootstrapping OceanBase at %s:%d", cfg.ObserverHost, cfg.SQLPort)
	if _, err := db.ExecContext(ctx, bootstrapSQL); err != nil {
		if clusterReady(ctx, db) {
			log.Println("OceanBase became ready while bootstrap was running")
			return
		}
		if bootstrapInProgress(err) {
			log.Println("another OceanBase bootstrap is already running; waiting for it to finish")
			if err := waitForCluster(ctx, db, cfg.RetryInterval); err != nil {
				log.Fatal(err)
			}
			log.Println("OceanBase bootstrap completed")
			return
		}
		log.Fatalf("bootstrap OceanBase: %v", err)
	}
	if err := waitForCluster(ctx, db, cfg.RetryInterval); err != nil {
		log.Fatal(err)
	}
	log.Println("OceanBase bootstrap completed")
}

func bootstrapInProgress(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "bootstrap privilege") || strings.Contains(message, "already bootstrap")
}

func loadConfig(getenv func(string) string) (config, error) {
	cfg := config{
		ObserverHost:           envOrDefault(getenv, "OBSERVER_HOST", "observer"),
		Zone:                   envOrDefault(getenv, "ZONE_NAME", "zone1"),
		LogServiceBucketURL:    getenv("LOGSERVICE_BUCKET_URL"),
		SharedStorageBucketURL: getenv("SHARED_STORAGE_BUCKET_URL"),
		AccessIDFile:           envOrDefault(getenv, "ACCESS_ID_FILE", "/run/secrets/minio_access_id"),
		AccessKeyFile:          envOrDefault(getenv, "ACCESS_KEY_FILE", "/run/secrets/minio_access_key"),
		Timeout:                defaultTimeout,
		RetryInterval:          defaultRetryInterval,
	}
	var err error
	if cfg.SQLPort, err = positiveInt(getenv, "OBSERVER_SQL_PORT", 2881); err != nil {
		return config{}, err
	}
	if cfg.RPCPort, err = positiveInt(getenv, "OBSERVER_RPC_PORT", 2882); err != nil {
		return config{}, err
	}
	clusterID, err := positiveUint(getenv, "LOGSERVICE_CLUSTER_ID")
	if err != nil {
		return config{}, err
	}
	cfg.LogServiceClusterID = clusterID
	if raw := getenv("BOOTSTRAP_TIMEOUT"); raw != "" {
		cfg.Timeout, err = time.ParseDuration(raw)
		if err != nil || cfg.Timeout <= 0 {
			return config{}, fmt.Errorf("invalid BOOTSTRAP_TIMEOUT %q", raw)
		}
	}
	if raw := getenv("BOOTSTRAP_RETRY_INTERVAL"); raw != "" {
		cfg.RetryInterval, err = time.ParseDuration(raw)
		if err != nil || cfg.RetryInterval <= 0 {
			return config{}, fmt.Errorf("invalid BOOTSTRAP_RETRY_INTERVAL %q", raw)
		}
	}
	if cfg.LogServiceBucketURL == "" || cfg.SharedStorageBucketURL == "" {
		return config{}, errors.New("LOGSERVICE_BUCKET_URL and SHARED_STORAGE_BUCKET_URL are required")
	}
	return cfg, nil
}

func buildBootstrapSQL(cfg config, accessID, accessKey string) (string, error) {
	values := map[string]string{
		"observer host":             cfg.ObserverHost,
		"zone":                      cfg.Zone,
		"logservice bucket URL":     cfg.LogServiceBucketURL,
		"shared storage bucket URL": cfg.SharedStorageBucketURL,
		"access ID":                 accessID,
		"access key":                accessKey,
	}
	for name, value := range values {
		if err := validateSQLValue(name, value); err != nil {
			return "", err
		}
	}
	logServiceAccessPoint := fmt.Sprintf("%s&access_id=%s&access_key=%s&logservice_cluster_id=%d", cfg.LogServiceBucketURL, accessID, accessKey, cfg.LogServiceClusterID)
	sharedStorageInfo := fmt.Sprintf("%s&access_id=%s&access_key=%s", cfg.SharedStorageBucketURL, accessID, accessKey)
	return fmt.Sprintf(
		"ALTER SYSTEM BOOTSTRAP ZONE '%s' SERVER '%s:%d', LOGSERVICE_ACCESS_POINT \"%s\" SHARED_STORAGE_INFO \"%s\"",
		cfg.Zone, cfg.ObserverHost, cfg.RPCPort, logServiceAccessPoint, sharedStorageInfo,
	), nil
}

func validateSQLValue(name, value string) error {
	if value == "" {
		return fmt.Errorf("%s is empty", name)
	}
	if strings.ContainsAny(value, "'\"\\\r\n\x00") {
		return fmt.Errorf("%s contains a character that cannot be used in bootstrap SQL", name)
	}
	return nil
}

func waitForConnection(ctx context.Context, db *sql.DB, interval time.Duration) error {
	for {
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := db.PingContext(pingCtx)
		cancel()
		if err == nil {
			return nil
		}
		if err := wait(ctx, interval); err != nil {
			return fmt.Errorf("wait for OceanBase SQL endpoint: %w", err)
		}
	}
}

func waitForCluster(ctx context.Context, db *sql.DB, interval time.Duration) error {
	for {
		if clusterReady(ctx, db) {
			return nil
		}
		if err := wait(ctx, interval); err != nil {
			return fmt.Errorf("wait for bootstrapped OceanBase cluster: %w", err)
		}
	}
}

func clusterReady(ctx context.Context, db *sql.DB) bool {
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var serverCount int
	err := db.QueryRowContext(queryCtx, "SELECT COUNT(*) FROM oceanbase.DBA_OB_SERVERS").Scan(&serverCount)
	return err == nil && serverCount > 0
}

func wait(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func readCredential(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	credential := strings.TrimSpace(string(content))
	if credential == "" {
		return "", errors.New("credential file is empty")
	}
	return credential, nil
}

func envOrDefault(getenv func(string) string, name, fallback string) string {
	if value := getenv(name); value != "" {
		return value
	}
	return fallback
}

func positiveInt(getenv func(string) string, name string, fallback int) (int, error) {
	raw := getenv(name)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 || value > 65535 {
		return 0, fmt.Errorf("invalid %s %q", name, raw)
	}
	return value, nil
}

func positiveUint(getenv func(string) string, name string) (uint64, error) {
	raw := getenv(name)
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || value == 0 {
		return 0, fmt.Errorf("invalid %s %q", name, raw)
	}
	return value, nil
}
