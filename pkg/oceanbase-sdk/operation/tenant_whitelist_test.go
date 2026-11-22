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
	drivertypes "database/sql"
	"errors"
	"testing"

	"github.com/go-sql-driver/mysql"

	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/sql"
)

func TestReadTenantConnectWhiteList(t *testing.T) {
	missingTable := &mysql.MySQLError{Number: 1146}
	for _, tc := range []struct {
		name     string
		value    string
		viewErr  error
		fallback bool
		wantErr  error
	}{
		{name: "restricted", value: "10.0.%"},
		{name: "empty stored value"},
		{name: "legacy view", value: "192.168.%", viewErr: missingTable, fallback: true},
		{name: "missing row", viewErr: drivertypes.ErrNoRows, wantErr: drivertypes.ErrNoRows},
		{name: "query failure", viewErr: context.DeadlineExceeded, wantErr: context.DeadlineExceeded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			const tenantID int64 = 1002
			query := func(_ context.Context, out any, stmt string, args ...any) error {
				calls++
				if len(args) != 1 || args[0] != tenantID {
					t.Fatalf("unexpected target %v", args)
				}
				switch stmt {
				case sql.GetTenantConnectWhiteList:
					if tc.viewErr != nil {
						return tc.viewErr
					}
				case sql.GetLegacyTenantConnectWhiteList:
					if !tc.fallback {
						t.Fatal("unexpected fallback")
					}
				default:
					t.Fatalf("unexpected query %q", stmt)
				}
				*out.(*string) = tc.value
				return nil
			}
			got, err := readTenantConnectWhiteList(context.Background(), query, tenantID)
			if !errors.Is(err, tc.wantErr) || got != tc.value || (calls == 2) != tc.fallback {
				t.Fatalf("read = %q, %v after %d calls", got, err, calls)
			}
		})
	}
}
