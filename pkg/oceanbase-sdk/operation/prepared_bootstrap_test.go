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

package operation

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/jmoiron/sqlx"

	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/config"
)

type bootstrapDriver struct {
	opened      int
	closed      int
	queries     []string
	connections []int
	timeout     time.Duration
	failSetup   bool
	failSend    bool
}

func (d *bootstrapDriver) Open(string) (driver.Conn, error) { return d.Connect(context.Background()) }
func (d *bootstrapDriver) Driver() driver.Driver            { return d }
func (d *bootstrapDriver) Connect(context.Context) (driver.Conn, error) {
	d.opened++
	return &bootstrapConn{db: d, id: d.opened}, nil
}

type bootstrapConn struct {
	db *bootstrapDriver
	id int
}

func (c *bootstrapConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepared SQL")
}
func (c *bootstrapConn) Begin() (driver.Tx, error) { return nil, errors.New("unexpected transaction") }
func (c *bootstrapConn) Close() error              { c.db.closed++; return nil }
func (c *bootstrapConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.db.queries = append(c.db.queries, query)
	c.db.connections = append(c.db.connections, c.id)
	if strings.HasPrefix(query, "set ") {
		timeout := c.db.timeout
		if timeout == 0 {
			timeout = config.BootstrapTimeout
		}
		if len(args) != 1 || args[0].Value != int64(timeout/time.Microsecond) {
			return nil, errors.New("wrong timeout")
		}
		if c.db.failSetup {
			return nil, errors.New("session rejected timeout")
		}
	} else if c.db.failSend {
		return nil, errors.New("lost SQL response")
	}
	return driver.RowsAffected(0), nil
}

func TestPreparedClassicBootstrapPinsAndReleasesSQLSession(t *testing.T) {
	const query = "ALTER SYSTEM BOOTSTRAP ZONE 'z1' SERVER '192.0.2.1:2882'"
	const setup = "set ob_query_timeout=?"
	for _, tc := range []struct {
		name                            string
		setupFails, skipSend, sendFails bool
	}{
		{name: "preparation failed", setupFails: true},
		{name: "another worker owns claim", skipSend: true},
		{name: "dispatch completed"},
		{name: "dispatch failed", sendFails: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := &bootstrapDriver{failSetup: tc.setupFails, failSend: tc.sendFails}
			db := sqlx.NewDb(sql.OpenDB(d), "bootstrap")
			defer db.Close()
			db.SetMaxOpenConns(1)
			ctx := context.Background()
			statement, err := prepareClassicBootstrap(ctx, db, logr.Discard(), query)
			if (err != nil) != tc.setupFails {
				t.Fatalf("prepare: %v", err)
			}
			if tc.setupFails {
				if statement != nil {
					t.Fatal("failed preflight returned a dispatchable statement")
				}
			} else {
				if db.Stats().InUse != 1 {
					t.Fatal("preflight did not reserve a connection")
				}
				if !tc.skipSend {
					err = statement.Exec(ctx)
					if (err != nil) != tc.sendFails {
						t.Fatalf("dispatch: %v", err)
					}
				}
				statement.Close()
			}
			want := []string{setup}
			if !tc.setupFails && !tc.skipSend {
				want = append(want, query)
			}
			if !reflect.DeepEqual(d.queries, want) || d.opened != 1 || d.closed != 1 || db.Stats().InUse != 0 {
				t.Fatalf("queries=%v, opened=%d closed=%d in use=%d", d.queries, d.opened, d.closed, db.Stats().InUse)
			}
			for _, id := range d.connections {
				if id != 1 {
					t.Fatal("timeout and bootstrap used different physical sessions")
				}
			}
		})
	}
}
