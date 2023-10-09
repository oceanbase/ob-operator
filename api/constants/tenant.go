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

package constants

import "github.com/oceanbase/ob-operator/api/types"

const (
	TenantRolePrimary types.TenantRole = "PRIMARY"
	TenantRoleStandby types.TenantRole = "STANDBY"
)

const (
	TenantOpSwitchover types.TenantOperationType = "SWITCHOVER"
	TenantOpFailover   types.TenantOperationType = "FAILOVER"
	TenantOpChangePwd  types.TenantOperationType = "CHANGE_PASSWORD"
)

const (
	TenantOpStarting   types.TenantOperationStatus = "STARTING"
	TenantOpRunning    types.TenantOperationStatus = "RUNNING"
	TenantOpSuccessful types.TenantOperationStatus = "SUCCESSFUL"
	TenantOpFailed     types.TenantOperationStatus = "FAILED"
	TenantOpReverting  types.TenantOperationStatus = "REVERTING"
)
