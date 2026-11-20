package operation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/go-logr/logr"

	"github.com/oceanbase/ob-operator/pkg/database"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/config"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

func TestTenantCreateRecordsOwnerWithoutChangingSDKDefaults(t *testing.T) {
	param := model.TenantSQLParam{TenantName: "example", Charset: "utf8mb4", PrimaryZone: "zone1",
		PoolList: []string{"p1", "p2"}, Locality: "FULL{1}@zone1", Collate: "utf8mb4_general_ci", VariableList: "ob_tcp_invited_nodes = '%'"}
	baseArgs := []any{param.Charset, param.PrimaryZone, "p1", "p2", param.Locality, param.Collate}
	for _, owner := range []string{"", "ob-operator/cluster/tenant/3"} {
		query, args := preparedSQLForAddTenant(param, owner)
		prefix := "CREATE TENANT IF NOT EXISTS example"
		expectArgs := baseArgs
		if owner != "" {
			prefix = "CREATE TENANT example"
			expectArgs = append(append([]any(nil), baseArgs...), owner)
		}
		if !strings.HasPrefix(query, prefix) || strings.Contains(query, ", COMMENT = ? SET VARIABLES") != (owner != "") ||
			strings.Count(query, "?") != len(args) || !reflect.DeepEqual(args, expectArgs) {
			t.Fatalf("unexpected SQL or bind order: %q %v", query, args)
		}
	}
}

type tenantCreateDataSource struct {
	*connector.OceanBaseDataSource
	driverName string
}

func (ds tenantCreateDataSource) DriverName() string { return ds.driverName }

var tenantCreateDriverID atomic.Uint64

func TestCreateTenantDistinguishesSetupFromSentDDLOnPinnedConnection(t *testing.T) {
	for _, tc := range []struct {
		name                string
		reject, setup, send bool
	}{
		{name: "preflight", reject: true}, {name: "timeout setup", setup: true}, {name: "lost DDL", send: true}, {name: "success"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := &bootstrapDriver{timeout: config.TenantSqlTimeout, failSetup: tc.setup, failSend: tc.send}
			name := fmt.Sprintf("owned-tenant-%d", tenantCreateDriverID.Add(1))
			sql.Register(name, d)
			conn := database.NewConnector(tenantCreateDataSource{connector.NewOceanBaseDataSource("example.invalid", 2881, "test", "sys", "", "oceanbase"), name})
			if err := conn.Init(); err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			manager := NewOceanbaseOperationManager(conn)
			logger := logr.Discard()
			manager.Logger = &logger
			sent, err := manager.CreateTenantWithOwner(context.Background(), model.TenantSQLParam{TenantName: "example", PoolList: []string{"pool"}}, "owner", func(session *OceanbaseOperationManager) error {
				if session.session == nil {
					t.Fatal("preflight was not pinned")
				}
				if tc.reject {
					return errors.New("allocation changed")
				}
				return nil
			})
			if sent != (!tc.reject && !tc.setup) || (err != nil) != (tc.reject || tc.setup || tc.send) {
				t.Fatalf("sent=%v err=%v", sent, err)
			}
			if !tc.reject && !tc.setup && (len(d.queries) != 2 || d.connections[0] != d.connections[1]) {
				t.Fatalf("unexpected session or SQL count: %v %v", d.connections, d.queries)
			}
			if conn.GetClient().Stats().InUse != 0 || d.closed != 1 {
				t.Fatalf("pinned session remained active: %d", d.closed)
			}
		})
	}
}
