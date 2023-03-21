/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package converter

import (
	"reflect"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	tenantconst "github.com/oceanbase/ob-operator/pkg/controllers/tenant/const"

	tenantcore "github.com/oceanbase/ob-operator/pkg/controllers/tenant/core"
)

func IsTenantInstanceRunning(tenant cloudv1.Tenant) bool {
	return tenant.Status.Status == tenantconst.TenantRunning
}

func IsTenantResourceUnitMatched(tenant cloudv1.Tenant) bool {
	specResourceUnit := tenantcore.GenerateSpecResourceUnitV3Map(tenant.Spec)
	statusResourceUnit := tenantcore.GenerateStatusResourceUnitV3Map(tenant.Status)
	for _, zone := range tenant.Spec.Topology {
		if !tenantcore.IsUnitV3Equal(specResourceUnit[zone.ZoneName], statusResourceUnit[zone.ZoneName]) {
			return false
		}
	}
	return true
}

func IsTenantPrimaryZoneMatched(tenant cloudv1.Tenant) bool {
	specPrimaryZone := tenantcore.GenerateSpecPrimaryZone(tenant.Spec.Topology)
	statusPrimaryZone := tenantcore.GenerateStatusPrimaryZone(tenant.Status.Topology)
	specPrimaryZoneMap := tenantcore.GeneratePrimaryZoneMap(specPrimaryZone)
	statusPrimaryZoneMap := tenantcore.GeneratePrimaryZoneMap(statusPrimaryZone)
	return reflect.DeepEqual(specPrimaryZoneMap, statusPrimaryZoneMap)
}

func IsTenantUnitNumMatched(tenant cloudv1.Tenant) bool {
	specUnitNumMap := tenantcore.GenerateSpecUnitNumMap(tenant.Spec)
	statusUnitNumMap := tenantcore.GenerateStatusUnitNumMap(tenant.Status)
	return reflect.DeepEqual(specUnitNumMap, statusUnitNumMap)
}

func IsTenantLocalityMatched(tenant cloudv1.Tenant) bool {
	specLocalityMap := tenantcore.GenerateSpecLocalityMap(tenant.Spec.Topology)
	statusLocalityMap := tenantcore.GenerateStatusLocalityMap(tenant.Status.Topology)
	return reflect.DeepEqual(specLocalityMap, statusLocalityMap)
}
