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
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
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
	RootPasswordFile       string
	AllowLegacyEmptyRoot   bool
	Timeout                time.Duration
	RetryInterval          time.Duration
}

func main() {
	mode := flag.String("mode", "bootstrap", "bootstrap, check, smoke-write or smoke-read")
	flag.Parse()
	cfg, err := loadConfig(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	password, err := readCredential(cfg.RootPasswordFile)
	if err != nil {
		log.Fatalf("read root password: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	if err := run(ctx, cfg, *mode, password, func(pass string) (*sql.DB, error) {
		return openDatabase(cfg, pass)
	}); err != nil {
		log.Fatal(err)
	}
}

func openDatabase(cfg config, password string) (*sql.DB, error) {
	driverConfig := databaseConfig(cfg, password)
	db, err := sql.Open("mysql", driverConfig.FormatDSN())
	if err == nil {
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
	}
	return db, err
}

func databaseConfig(cfg config, password string) *mysql.Config {
	driverConfig := mysql.NewConfig()
	driverConfig.User = "root"
	driverConfig.Passwd = password
	driverConfig.Net = "tcp"
	driverConfig.Addr = net.JoinHostPort(cfg.ObserverHost, strconv.Itoa(cfg.SQLPort))
	driverConfig.Timeout = 5 * time.Second
	// Initial cluster bootstrap can take minutes. Individual probes have short
	// contexts; the bootstrap statement is bounded by the overall deadline.
	driverConfig.ReadTimeout = cfg.Timeout
	driverConfig.WriteTimeout = 30 * time.Second
	driverConfig.InterpolateParams = true
	return driverConfig
}

func run(ctx context.Context, cfg config, mode, password string, open func(string) (*sql.DB, error)) error {
	if mode != "bootstrap" && mode != "check" && mode != "smoke-write" && mode != "smoke-read" {
		return errors.New("invalid operation mode")
	}
	if password == "" {
		return errors.New("root password must not be empty")
	}
	db, err := open(password)
	if err != nil {
		return err
	}
	defer db.Close()
	err = waitForConnection(ctx, db, cfg.RetryInterval)
	if err != nil {
		if mode != "bootstrap" || !authenticationDenied(err) {
			return err
		}
		initialDB, err := open("")
		if err != nil {
			return err
		}
		defer initialDB.Close()
		if err := waitForConnection(ctx, initialDB, cfg.RetryInterval); err != nil {
			return fmt.Errorf("initial database access failed: %w", err)
		}
		if clusterReady(ctx, initialDB) {
			if !cfg.AllowLegacyEmptyRoot {
				return errors.New("existing database has an empty root password; set ALLOW_LEGACY_EMPTY_ROOT=true to migrate it explicitly")
			}
		} else {
			if err := bootstrapCluster(ctx, cfg, initialDB); err != nil {
				return err
			}
		}
		if _, err := initialDB.ExecContext(ctx, "ALTER USER root IDENTIFIED BY ?", password); err != nil {
			return errors.New("failed to initialize root password")
		}
		if err := waitForConnection(ctx, db, cfg.RetryInterval); err != nil {
			return err
		}
	} else if mode == "bootstrap" && !clusterReady(ctx, db) {
		// Some images accept an initial connection before the cluster is bootstrapped.
		if err := bootstrapCluster(ctx, cfg, db); err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, "ALTER USER root IDENTIFIED BY ?", password); err != nil {
			return errors.New("failed to initialize root password")
		}
	}
	if err := waitForCluster(ctx, db, cfg.RetryInterval); err != nil {
		return err
	}
	switch mode {
	case "smoke-write":
		for _, statement := range []string{"CREATE DATABASE IF NOT EXISTS compose_smoke", "CREATE TABLE IF NOT EXISTS compose_smoke.t (id BIGINT PRIMARY KEY, value VARCHAR(64))", "REPLACE INTO compose_smoke.t VALUES (1, 'lakehouse-compose')"} {
			if _, err := db.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("smoke write: %w", err)
			}
		}
		fallthrough
	case "smoke-read":
		var value string
		if err := db.QueryRowContext(ctx, "SELECT value FROM compose_smoke.t WHERE id=1").Scan(&value); err != nil {
			return fmt.Errorf("smoke read: %w", err)
		}
		if value != "lakehouse-compose" {
			return errors.New("smoke persistence value mismatch")
		}
	}
	log.Printf("OceanBase %s completed with authenticated SQL access", mode)
	return nil
}

func bootstrapCluster(ctx context.Context, cfg config, db *sql.DB) error {
	accessID, err := readCredential(cfg.AccessIDFile)
	if err != nil {
		return fmt.Errorf("read access ID: %w", err)
	}
	accessKey, err := readCredential(cfg.AccessKeyFile)
	if err != nil {
		return fmt.Errorf("read access key: %w", err)
	}
	statement, err := buildBootstrapSQL(cfg, accessID, accessKey)
	if err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, statement); err != nil && !bootstrapInProgress(err) && !clusterReady(ctx, db) {
		return errors.New("OceanBase bootstrap failed; inspect database logs without sharing credentials")
	}
	return waitForCluster(ctx, db, cfg.RetryInterval)
}

func authenticationDenied(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1045
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
		RootPasswordFile:       envOrDefault(getenv, "ROOT_PASSWORD_FILE", "/run/secrets/ob_root_password"),
		Timeout:                defaultTimeout,
		RetryInterval:          defaultRetryInterval,
	}
	var err error
	if raw := getenv("ALLOW_LEGACY_EMPTY_ROOT"); raw != "" {
		cfg.AllowLegacyEmptyRoot, err = strconv.ParseBool(raw)
		if err != nil {
			return config{}, errors.New("invalid ALLOW_LEGACY_EMPTY_ROOT")
		}
	}
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
		if authenticationDenied(err) {
			return fmt.Errorf("database authentication rejected: %w", err)
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
