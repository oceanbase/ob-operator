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
			if got := observedPoolLocality(zone, active, configured); !reflect.DeepEqual(got, expected) {
				t.Errorf("observed locality = %+v, want %+v", got, expected)
			}
		})
	}
	if !configured["activating"].IsActive {
		t.Fatal("observing the pool changed the configured target")
	}
}
