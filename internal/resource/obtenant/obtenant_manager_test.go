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

package obtenant

import (
	"database/sql"
	"reflect"
	"testing"

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

func TestTenantRoleFromDatabase(t *testing.T) {
	tests := []struct {
		name     string
		current  apitypes.TenantRole
		database string
		want     apitypes.TenantRole
	}{
		{
			name:     "failed primary activation remains standby",
			current:  apiconst.TenantRolePrimary,
			database: string(apiconst.TenantRoleStandby),
			want:     apiconst.TenantRoleStandby,
		},
		{
			name:     "database primary replaces stale standby",
			current:  apiconst.TenantRoleStandby,
			database: string(apiconst.TenantRolePrimary),
			want:     apiconst.TenantRolePrimary,
		},
		{
			name:     "missing database role preserves current status",
			current:  apiconst.TenantRolePrimary,
			database: "",
			want:     apiconst.TenantRolePrimary,
		},
		{
			name:     "unknown database role preserves current status",
			current:  apiconst.TenantRoleStandby,
			database: "UNKNOWN",
			want:     apiconst.TenantRoleStandby,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tenantRoleFromDatabase(tt.current, &model.OBTenant{TenantRole: tt.database})
			if got != tt.want {
				t.Fatalf("tenantRoleFromDatabase() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTenantRoleFromDatabaseHandlesMissingRecord(t *testing.T) {
	got := tenantRoleFromDatabase(apiconst.TenantRolePrimary, nil)
	if got != apiconst.TenantRolePrimary {
		t.Fatalf("tenantRoleFromDatabase() = %q, want %q", got, apiconst.TenantRolePrimary)
	}
}

func TestObservedPoolLocality(t *testing.T) {
	active := map[string]v1alpha1.LocalityType{"active": {Name: "READONLY", Replica: 1, IsActive: true}}
	configured := map[string]*v1alpha1.LocalityType{
		"active":     {Name: "FULL", Replica: 1, IsActive: true},
		"reserved":   {Name: "FULL", Replica: 1},
		"activating": {Name: "READONLY", Replica: 2, IsActive: true},
	}
	want := map[string]*v1alpha1.LocalityType{
		"active":     {Name: "READONLY", Replica: 1, IsActive: true},
		"reserved":   {Name: "FULL", Replica: 1},
		"activating": {Name: "READONLY", Replica: 2},
		"unknown":    nil,
	}
	for zone, expected := range want {
		t.Run(zone, func(t *testing.T) {
			if got := observedPoolLocality(zone, active, configured, zone != "unknown"); !reflect.DeepEqual(got, expected) {
				t.Errorf("observed locality = %+v, want %+v", got, expected)
			}
		})
	}
	if !configured["activating"].IsActive {
		t.Fatal("observing the pool changed the configured target")
	}
}

func TestTenantPoolRecord(t *testing.T) {
	m := &OBTenantManager{OBTenant: &v1alpha1.OBTenant{Spec: v1alpha1.OBTenantSpec{TenantName: "orders"}}}
	if _, err := m.poolRecord(nil); err == nil {
		t.Fatal("expected a retryable status-read error before any resource pools are observed")
	}
	got, err := m.poolRecord([]v1alpha1.ResourcePoolStatus{{ZoneList: "z1", UnitNumber: 2}, {ZoneList: "z2", UnitNumber: 2}})
	if err != nil || got.UnitNumber != 2 || got.PoolList != "pool_orders_z1,pool_orders_z2" || got.ZoneList != "z1,z2" {
		t.Fatalf("pool record = %+v, error = %v", got, err)
	}
}

func TestInterruptedPoolRemoval(t *testing.T) {
	m := &OBTenantManager{OBTenant: &v1alpha1.OBTenant{Spec: v1alpha1.OBTenantSpec{TenantName: "orders"}}}
	owner := sql.NullInt64{Int64: 1002, Valid: true}
	pools := []model.Pool{
		{ResourcePoolID: 1, Name: "pool_orders_zone1", ZoneList: "zone1", TenantID: owner, UnitNum: 1},
		{ResourcePoolID: 2, Name: "pool_orders_zone3", ZoneList: "zone3", TenantID: owner, UnitNum: 1},
		{ResourcePoolID: 3, Name: "manual", ZoneList: "zone2, zone3", TenantID: owner, UnitNum: 2},
		{ResourcePoolID: 4, Name: "pool_other_zone3", ZoneList: "zone3", TenantID: sql.NullInt64{Int64: 1004, Valid: true}, UnitNum: 1},
		{ResourcePoolID: 5, Name: "pool_orders_zone4", ZoneList: "zone4", UnitNum: 1},
	}
	units := m.managedPoolUnits(1002, pools)
	if !reflect.DeepEqual(units, map[string]int{"zone1": 1, "zone3": 1}) {
		t.Fatalf("unexpected managed pool units %v", units)
	}
	for _, zone := range []string{"zone1", "zone3", "zone4"} {
		_, managed := units[zone]
		configured := map[string]*v1alpha1.LocalityType{"zone1": {Name: "FULL", Replica: 1}}
		locality := observedPoolLocality(zone, nil, configured, managed)
		if zone == "zone4" {
			if locality != nil {
				t.Fatalf("unassigned pool should be ignored, got %+v", locality)
			}
			continue
		}
		m.OBTenant.Status.Pools = append(m.OBTenant.Status.Pools, v1alpha1.ResourcePoolStatus{ZoneList: zone, Type: locality})
	}
	m.OBTenant.Spec.Pools = []v1alpha1.ResourcePoolSpec{{Zone: "zone1", Type: &v1alpha1.LocalityType{Name: "FULL", Replica: 1}}}
	toRemove := m.getPoolsForDelete()
	if len(toRemove) != 1 || toRemove[0].ZoneList != "zone3" || toRemove[0].Type == nil || toRemove[0].Type.IsActive {
		t.Fatalf("expected only the interrupted removal, got %+v", toRemove)
	}
	plan := planPoolRemoval(pools, 1002, "pool_orders_zone3", "zone3")
	if plan.target.ResourcePoolID != 2 || !plan.sharedZone || !reflect.DeepEqual(plan.remaining, []string{"pool_orders_zone1", "manual"}) {
		t.Fatalf("unexpected removal plan %+v", plan)
	}
	for _, target := range pools[3:] {
		if planPoolRemoval(pools, 1002, target.Name, target.ZoneList).target.ResourcePoolID != 0 {
			t.Fatalf("claimed foreign or detached pool %s", target.Name)
		}
	}
	// After the canonical pool is gone an external pool still serves zone3.
	// The surviving database locality must not schedule another removal.
	remaining := append([]model.Pool{pools[0]}, pools[2:]...)
	remainingUnits := m.managedPoolUnits(1002, remaining)
	_, managed := remainingUnits["zone3"]
	active := map[string]v1alpha1.LocalityType{"zone3": {Name: "FULL", Replica: 1, IsActive: true}}
	if got := observedPoolLocality("zone3", active, nil, managed); got != nil {
		t.Fatalf("external-only locality still scheduled a removal: %+v", got)
	}
	configs := []model.UnitConfigV4{{UnitConfigID: 1003, Name: "unit_orders_zone3"}}
	if !removableUnitConfig(1003, "unit_orders_zone3", configs, nil) ||
		removableUnitConfig(1003, "unit_orders_zone3", configs, []model.Pool{{UnitConfigID: 1003}}) ||
		removableUnitConfig(1004, "unit_orders_zone3", configs, nil) ||
		removableUnitConfig(1003, "unit_other_zone3", configs, nil) {
		t.Fatal("unit configuration ownership or sharing check failed")
	}
}
