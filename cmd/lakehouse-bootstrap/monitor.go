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
	"errors"
)

// ensureMonitorUser keeps the Compose OBAgent account aligned with the
// persisted credential. The account intentionally receives read-only access.
func ensureMonitorUser(ctx context.Context, db *sql.DB, password string) error {
	if password == "" {
		return errors.New("monitor password must not be empty")
	}
	statements := []struct {
		query string
		args  []any
	}{
		{"CREATE USER IF NOT EXISTS monitor", nil},
		{"ALTER USER monitor IDENTIFIED BY ?", []any{password}},
		{"GRANT SELECT ON *.* TO monitor", nil},
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement.query, statement.args...); err != nil {
			return errors.New("failed to provision the OBAgent monitor user")
		}
	}
	return nil
}
