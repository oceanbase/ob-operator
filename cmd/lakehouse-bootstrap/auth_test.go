/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND.
*/

package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

type testDatabase struct {
	password                                string
	ready                                   bool
	acceptInitialPassword                   bool
	failBootstrap                           bool
	bootstraps, passwordChanges, emptyOpens int
}
type testDriver struct {
	state    *testDatabase
	password string
}
type testConnection struct{ *testDriver }
type testRows struct {
	done  bool
	value driver.Value
}

func (r *testRows) Columns() []string { return []string{"value"} }
func (r *testRows) Close() error      { return nil }
func (r *testRows) Next(values []driver.Value) error {
	if r.done {
		return io.EOF
	}
	values[0] = r.value
	r.done = true
	return nil
}
func (d *testDriver) Open(string) (driver.Conn, error) { return &testConnection{d}, nil }
func (c *testConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (c *testConnection) Close() error              { return nil }
func (c *testConnection) Begin() (driver.Tx, error) { return nil, errors.New("unexpected transaction") }
func (c *testConnection) Ping(context.Context) error {
	if c.password != c.state.password && !(c.state.acceptInitialPassword && !c.state.ready) {
		return &mysql.MySQLError{Number: 1045, Message: "access denied"}
	}
	return nil
}
func (c *testConnection) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if !c.state.ready {
		return nil, errors.New("not initialized")
	}
	if strings.Contains(query, "DBA_OB_SERVERS") {
		return &testRows{value: int64(1)}, nil
	}
	return &testRows{value: "lakehouse-compose"}, nil
}
func (c *testConnection) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if strings.HasPrefix(query, "ALTER SYSTEM BOOTSTRAP") {
		c.state.bootstraps++
		if c.state.failBootstrap {
			return nil, errors.New("storage secret=do-not-leak-this")
		}
		c.state.ready = true
	}
	if strings.HasPrefix(query, "ALTER USER") {
		c.state.passwordChanges++
		c.state.password = args[0].Value.(string)
	}
	return driver.RowsAffected(1), nil
}

type testConnector struct{ *testDriver }

func (c testConnector) Connect(context.Context) (driver.Conn, error) { return c.Open("") }
func (c testConnector) Driver() driver.Driver                        { return c.testDriver }

func TestRootPasswordLifecycle(t *testing.T) {
	for _, tc := range []struct {
		name                       string
		state                      testDatabase
		legacy                     bool
		mode                       string
		wantError                  bool
		bootstrap, passwordChanges int
	}{
		{name: "fresh database", mode: "bootstrap", bootstrap: 1, passwordChanges: 1},
		{name: "initial connections accept password", state: testDatabase{acceptInitialPassword: true}, mode: "bootstrap", bootstrap: 1, passwordChanges: 1},
		{name: "already secured", state: testDatabase{password: "test-only-password", ready: true}, mode: "bootstrap"},
		{name: "wrong persisted password", state: testDatabase{password: "other-password", ready: true}, mode: "bootstrap", wantError: true},
		{name: "legacy requires opt in", state: testDatabase{ready: true}, mode: "bootstrap", wantError: true},
		{name: "explicit legacy migration", state: testDatabase{ready: true}, mode: "bootstrap", legacy: true, passwordChanges: 1},
		{name: "check never initializes", mode: "check", wantError: true},
		{name: "bootstrap errors redact secrets", state: testDatabase{failBootstrap: true}, mode: "bootstrap", wantError: true, bootstrap: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			cfg := config{ObserverHost: "observer", RPCPort: 2882, Zone: "zone1", LogServiceBucketURL: "s3://logs?host=http://minio:9000", SharedStorageBucketURL: "s3://data?host=http://minio:9000", LogServiceClusterID: 1001, RetryInterval: time.Millisecond, AllowLegacyEmptyRoot: tc.legacy, AccessIDFile: filepath.Join(dir, "id"), AccessKeyFile: filepath.Join(dir, "key")}
			for _, path := range []string{cfg.AccessIDFile, cfg.AccessKeyFile} {
				if err := os.WriteFile(path, []byte("test-only-credential"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err := run(ctx, cfg, tc.mode, "test-only-password", func(password string) (*sql.DB, error) {
				if password == "" {
					tc.state.emptyOpens++
				}
				return sql.OpenDB(testConnector{&testDriver{state: &tc.state, password: password}}), nil
			})
			if (err != nil) != tc.wantError {
				t.Fatalf("error=%v", err)
			}
			if err != nil && strings.Contains(err.Error(), "do-not-leak-this") {
				t.Fatal("bootstrap secret leaked")
			}
			if tc.state.bootstraps != tc.bootstrap || tc.state.passwordChanges != tc.passwordChanges {
				t.Fatalf("unexpected mutations: %+v", tc.state)
			}
			if tc.mode == "check" && tc.state.emptyOpens != 0 {
				t.Fatal("check attempted empty-password fallback")
			}
		})
	}
}

func TestAuthenticationDenied(t *testing.T) {
	if !authenticationDenied(&mysql.MySQLError{Number: 1045}) || authenticationDenied(errors.New("connection refused")) {
		t.Fatal("authentication error classification")
	}
}

func TestDatabaseTimeoutAndPasswordEncoding(t *testing.T) {
	cfg := config{ObserverHost: "::1", SQLPort: 2881, Timeout: 30 * time.Minute}
	password := "test-only:@/'\\password"
	dbConfig := databaseConfig(cfg, password)
	parsed, err := mysql.ParseDSN(dbConfig.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Passwd != password || parsed.Addr != "[::1]:2881" {
		t.Fatal("database credential or address was changed")
	}
	if parsed.ReadTimeout != cfg.Timeout || parsed.Timeout != 5*time.Second {
		t.Fatal("bootstrap deadline truncated to short socket timeout")
	}
}

func TestInvalidModeAndEmptyPasswordDoNotConnect(t *testing.T) {
	for _, tc := range []struct{ mode, password string }{{"invalid", "secret"}, {"bootstrap", ""}} {
		err := run(context.Background(), config{}, tc.mode, tc.password, func(string) (*sql.DB, error) { t.Fatal("unexpected connection"); return nil, nil })
		if err == nil {
			t.Fatal("invalid operation was accepted")
		}
	}
}
