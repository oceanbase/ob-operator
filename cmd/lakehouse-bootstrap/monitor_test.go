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
	"reflect"
	"testing"
)

func TestEnsureMonitorUserUsesReadOnlyGrant(t *testing.T) {
	state := &testDatabase{password: "root-password", ready: true}
	db := sql.OpenDB(testConnector{&testDriver{state: state, password: "root-password"}})
	defer db.Close()
	if err := ensureMonitorUser(context.Background(), db, "monitor-password"); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"CREATE USER IF NOT EXISTS monitor",
		"ALTER USER monitor IDENTIFIED BY ?",
		"GRANT SELECT ON *.* TO monitor",
	}
	if !reflect.DeepEqual(state.queries, want) {
		t.Fatalf("queries = %#v, want %#v", state.queries, want)
	}
}

func TestEnsureMonitorUserRejectsEmptyPassword(t *testing.T) {
	if err := ensureMonitorUser(context.Background(), nil, ""); err == nil {
		t.Fatal("empty password was accepted")
	}
}
