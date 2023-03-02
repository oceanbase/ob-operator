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

package sql

import (
	"strconv"
	"strings"

	v1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"github.com/oceanbase/ob-operator/pkg/controllers/tenant/model"
)

func ReplaceAll(template string, replacers ...*strings.Replacer) string {
	s := template
	for _, replacer := range replacers {
		s = replacer.Replace(s)
	}
	return s
}

func SetParameterSQLReplacer(name, value string) *strings.Replacer {
	return strings.NewReplacer("${NAME}", name, "${VALUE}", value)
}

func CreateUnitV3SQLReplacer(unitName string, resourceUnit v1.ResourceUnit) *strings.Replacer {
	return strings.NewReplacer("${UNIT_NAME}", unitName, "${MAX_CPU}", strconv.FormatFloat(resourceUnit.MaxCPU.AsApproximateFloat64(), 'f', -1, 64), "${MAX_MEMORY}", strconv.Itoa(int(resourceUnit.MemorySize.Value()))+"B", "${MAX_IOPS}", strconv.Itoa(resourceUnit.MaxIops), "${MAX_DISK_SIZE}", strconv.Itoa(int(resourceUnit.MaxDiskSize.Value()))+"B", "${MAX_SESSION_NUM}", strconv.Itoa(resourceUnit.MaxSessionNum), "${MIN_CPU}", strconv.FormatFloat(resourceUnit.MinCPU.AsApproximateFloat64(), 'f', -1, 64), "${MIN_MEMORY}", strconv.Itoa(int(resourceUnit.MemorySize.Value()))+"B", "${MIN_IOPS}", strconv.Itoa(resourceUnit.MinIops))
}

func CreateUnitV4SQLReplacer(unitName string, resourceUnit v1.ResourceUnit, option string) *strings.Replacer {
	return strings.NewReplacer("${UNIT_NAME}", unitName, "${MAX_CPU}", strconv.FormatFloat(resourceUnit.MaxCPU.AsApproximateFloat64(), 'f', -1, 64), "${MEMORY_SIZE}", strconv.Itoa(int(resourceUnit.MemorySize.Value())), "${OPTION}", option)
}

func CreatePoolSQLReplacer(poolName, unitName string, zone v1.TenantReplica) *strings.Replacer {
	return strings.NewReplacer("${POOL_NAME}", poolName, "${UNIT_NAME}", unitName, "${UNIT_NUM}", strconv.Itoa(zone.UnitNumber), "${ZONE_NAME}", zone.ZoneName)
}

func GetResourceSQLReplacer(zoneName string) *strings.Replacer {
	return strings.NewReplacer("${ZONE_NAME}", zoneName)
}

func CreateTenantSQLReplacer(tenantName, charset, zoneList, primaryZone, poolList, locality, collate, variableList string) *strings.Replacer {
	return strings.NewReplacer("${TENANT_NAME}", tenantName, "${CHARSET}", charset, "${ZONE_LIST}", zoneList, "${PRIMARY_ZONE}", primaryZone, "${RESOURCE_POOL_LIST}", poolList, "${LOCALITY}", locality, "${COLLATE}", collate, "${VARIABLE_LIST}", variableList)
}

func GetVariableSQLReplacer(name string, tenantID int) *strings.Replacer {
	return strings.NewReplacer("${NAME}", name, "${TENANT_ID}", strconv.Itoa(tenantID))
}

func SetTenantVariableSQLReplacer(tenantName, name, value string) *strings.Replacer {
	return strings.NewReplacer("${TENANT_NAME}", tenantName, "${NAME}", name, "${VALUE}", value)
}

func SetUnitConfigV3SQLReplacer(unitName string, resourceUnit model.ResourceUnitV3) *strings.Replacer {
	return strings.NewReplacer("${UNIT_NAME}", unitName, "${MAX_CPU}", strconv.FormatFloat(resourceUnit.MaxCPU.AsApproximateFloat64(), 'f', -1, 64), "${MAX_MEMORY}", strconv.Itoa(int(resourceUnit.MemorySize.Value()))+"B", "${MAX_IOPS}", strconv.Itoa(resourceUnit.MaxIops), "${MAX_DISK_SIZE}", strconv.Itoa(int(resourceUnit.MaxDiskSize.Value()))+"B", "${MAX_SESSION_NUM}", strconv.Itoa(resourceUnit.MaxSessionNum), "${MIN_CPU}", strconv.FormatFloat(resourceUnit.MinCPU.AsApproximateFloat64(), 'f', -1, 64), "${MIN_MEMORY}", strconv.Itoa(int(resourceUnit.MemorySize.Value()))+"B", "${MIN_IOPS}", strconv.Itoa(resourceUnit.MinIops))
}

func SetUnitConfigV4SQLReplacer(unitName string, resourceUnit model.ResourceUnitV4) *strings.Replacer {
	return strings.NewReplacer("${UNIT_NAME}", unitName, "${MAX_CPU}", strconv.FormatFloat(resourceUnit.MaxCPU.AsApproximateFloat64(), 'f', -1, 64), "${MAX_MEMORY}", "${MAX_IOPS}", strconv.Itoa(resourceUnit.MaxIops), "${MAX_DISK_SIZE}", "${MIN_CPU}", strconv.FormatFloat(resourceUnit.MinCPU.AsApproximateFloat64(), 'f', -1, 64), "${MIN_IOPS}", strconv.Itoa(resourceUnit.MinIops))
}

func SetPoolUnitNumSQLReplacer(poolName string, unitNum int) *strings.Replacer {
	return strings.NewReplacer("${POOL_NAME}", poolName, "${UNIT_NUM}", strconv.Itoa(unitNum))
}

func SetTenantLocalitySQLReplacer(name, locality string) *strings.Replacer {
	return strings.NewReplacer("${TENANT_NAME}", name, "${LOCALITY}", locality)
}

func SetTenantPoolListSQLReplacer(name, locality string) *strings.Replacer {
	return strings.NewReplacer("${TENANT_NAME}", name, "${POOL_LIST}", locality)
}

func SetTenantSQLReplacer(name, zoneList, primaryZone, poolList, charset, locality string) *strings.Replacer {
	return strings.NewReplacer("${TENANT_NAME}", name, "${ZONE_LIST}", zoneList, "${PRIMARY_ZONE}", primaryZone, "${RESOURCE_POOL_LIST}", poolList, "${CHARSET}", charset, "${LOCALITY}", locality)
}

func SetNameReplacer(name string) *strings.Replacer {
	return strings.NewReplacer("${NAME}", name)
}
