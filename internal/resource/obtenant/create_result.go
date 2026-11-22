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

package obtenant

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	dbconfig "github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/config"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func createOwner(b *v1alpha1.TenantCreation) string {
	return fmt.Sprintf("ob-operator/%s/%s", b.ClusterUID, b.TenantUID)
}

func confirmTenantCreate(sent bool, createErr error, observe func() error) error {
	if createErr == nil || !sent {
		return createErr
	}
	var serverErr *mysql.MySQLError
	var networkErr net.Error
	ambiguous := errors.As(createErr, &networkErr) || errors.Is(createErr, driver.ErrBadConn) ||
		errors.Is(createErr, mysql.ErrInvalidConn) || errors.Is(createErr, io.EOF) || errors.Is(createErr, io.ErrUnexpectedEOF)
	if errors.As(createErr, &serverErr) {
		switch serverErr.Number {
		case 2006, 2013, 4012, 4638, 5156: // lost connection, timeout, root service changed or duplicate tenant
			ambiguous = true
		default:
			// Other server codes preserve the transport error classification above.
		}
	}
	if !ambiguous {
		return createErr
	}
	if err := observe(); err != nil {
		return errors.Join(createErr, err)
	}
	return nil
}

// A lost result can belong to a previous leader that already issued CREATE.
// Verification uses only SELECT statements and never resubmits the DDL.
func (m *OBTenantManager) ObserveLostTask(name tasktypes.TaskName) tasktypes.TaskFunc {
	if name == tCreateResourcePoolAndConfig && m.OBTenant != nil && m.OBTenant.Status.Creation != nil {
		return func() tasktypes.TaskError {
			_, db, err := m.creationDB(m.Ctx)
			if err != nil {
				return err
			}
			// No DDL is repeated when strict CREATE succeeded but its ID was lost.
			return m.checkCreateAllocation(db)
		}
	}
	if name != tCreateTenantWithClear {
		return nil
	}
	return func() tasktypes.TaskError {
		if m.OBTenant.Status.Creation == nil {
			live, _, err := m.readCreationObjects(m.Ctx)
			if err != nil {
				return err
			}
			if live.Status.Creation == nil {
				return errors.New("pre-upgrade CREATE result is unknown; retaining the tenant and its allocations without resubmission")
			}
		}
		return m.waitForCreatedTenant()
	}
}

type createdTenantRow struct {
	ClusterMatch int    `db:"cluster_match"`
	TenantID     int64  `db:"tenant_id"`
	Owned        bool   `db:"owned"`
	TenantType   string `db:"tenant_type"`
	Role         string `db:"tenant_role"`
	Status       string `db:"status"`
	Pool         string `db:"pool_name"`
	Zone         string `db:"zone_list"`
	Units        int    `db:"unit_count"`
	PoolID       int64  `db:"pool_id"`
	UnitID       int64  `db:"unit_id"`
	UnitConfig   string `db:"unit_name"`
}

// DBA_OB_TENANTS does not expose a tenant's COMMENT. The schema row stores it
// as info. Read it together with the public views so all checks share a snapshot.
const createdTenantSQL = `SELECT c.cluster_match, COALESCE(t.tenant_id,0) AS tenant_id,
COALESCE(BINARY s.info=BINARY ?,0) AS owned, COALESCE(t.tenant_type,'') AS tenant_type,
COALESCE(t.tenant_role,'') AS tenant_role, COALESCE(t.status,'') AS status,
COALESCE(p.resource_pool_id,0) AS pool_id, COALESCE(u.unit_config_id,0) AS unit_id,
COALESCE(p.name,'') AS pool_name, COALESCE(p.zone_list,'') AS zone_list,
COALESCE(p.unit_count,0) AS unit_count, COALESCE(u.name,'') AS unit_name
FROM (SELECT COUNT(DISTINCT name) AS cluster_match FROM oceanbase.GV$OB_PARAMETERS
      WHERE (name='cluster' AND value=?) OR (name='cluster_id' AND value=?)) c
LEFT JOIN oceanbase.DBA_OB_TENANTS t ON t.tenant_name=?
LEFT JOIN oceanbase.__all_tenant s ON s.tenant_id=t.tenant_id
LEFT JOIN oceanbase.DBA_OB_RESOURCE_POOLS p ON p.tenant_id=t.tenant_id
LEFT JOIN oceanbase.DBA_OB_UNIT_CONFIGS u ON u.unit_config_id=p.unit_config_id`

func (m *OBTenantManager) verifyCreatedTenant(rows []createdTenantRow, tenantID *int64) (bool, error) {
	if len(rows) == 0 || rows[0].ClusterMatch != 2 {
		return false, errors.New("database cluster identity does not match the tenant's cluster")
	}
	first := rows[0]
	if first.TenantID == 0 {
		return false, nil
	}
	if !first.Owned || first.TenantID <= 1 || first.TenantType != "USER" || (*tenantID != 0 && *tenantID != first.TenantID) {
		return false, errors.New("database tenant does not match the submitted creation")
	}
	*tenantID = first.TenantID
	if first.Status == "CREATING" {
		return false, nil
	}
	if first.Status != "NORMAL" || first.Role != "PRIMARY" {
		return false, fmt.Errorf("created tenant has unexpected status %q or role %q", first.Status, first.Role)
	}
	expected := make(map[string]bool, len(m.OBTenant.Spec.Pools))
	for _, pool := range m.OBTenant.Spec.Pools {
		expected[pool.Zone] = true
	}
	if len(expected) == 0 || len(expected) != len(rows) {
		return false, errors.New("created tenant has a different set of resource pools")
	}
	for _, row := range rows {
		ids := m.OBTenant.Status.Creation.Resources[row.Zone]
		if ids.PoolID == 0 || ids.UnitConfigID == 0 || ids.PoolID != row.PoolID || ids.UnitConfigID != row.UnitID ||
			!expected[row.Zone] || row.TenantID != first.TenantID || row.Pool != m.generatePoolName(row.Zone) ||
			row.UnitConfig != m.generateUnitName(row.Zone) || row.Units != m.OBTenant.Spec.UnitNumber {
			return false, errors.New("created tenant resource pool or unit configuration differs from the requested allocation")
		}
		delete(expected, row.Zone)
	}
	return true, nil
}

func createdRows(ctx context.Context, db *operation.OceanbaseOperationManager, b *v1alpha1.TenantCreation) ([]createdTenantRow, error) {
	var rows []createdTenantRow
	err := db.QueryList(ctx, &rows, createdTenantSQL, createOwner(b), b.DatabaseCluster, strconv.FormatInt(b.DatabaseID, 10), b.TenantName)
	return rows, err
}

func (m *OBTenantManager) waitForCreatedTenant() error {
	if !m.OBTenant.DeletionTimestamp.IsZero() {
		return errors.New("tenant is being deleted")
	}
	var tenantID int64
	var lastReadErr error
	err := wait.PollUntilContextTimeout(m.Ctx, 5*time.Second, dbconfig.TenantSqlTimeout, true, func(ctx context.Context) (bool, error) {
		live, db, err := m.creationDB(ctx)
		if err != nil {
			if errors.Is(err, errCreationBindingChanged) || errors.Is(err, errForeignCluster) || errors.Is(err, errStaleTenant) {
				return false, err
			}
			lastReadErr = err
			return false, nil
		}
		rows, err := createdRows(ctx, db, live.Status.Creation)
		lastReadErr = err
		if err != nil {
			return false, nil //nolint:nilerr // Retry transient reads and report the saved error if polling times out.
		}
		manager := *m
		manager.OBTenant = live
		return manager.verifyCreatedTenant(rows, &tenantID)
	})
	if err != nil {
		return fmt.Errorf("could not verify completion of tenant creation: %w", errors.Join(err, lastReadErr))
	}
	return nil
}
