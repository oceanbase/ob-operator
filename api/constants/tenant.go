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

type TenantRole string

const (
	TenantRolePrimary TenantRole = "PRIMARY"
	TenantRoleStandby TenantRole = "STANDBY"
)

type TenantOperationType string

const (
	TenantOpSwitchover TenantOperationType = "SWITCHOVER"
	TenantOpFailover                       = "FAILOVER"
	TenantOpChangePwd                      = "CHANGE_PASSWORD"
)

type TenantOperationStatus string

const (
	TenantOpStarting   TenantOperationStatus = "STARTING"
	TenantOpRunning                          = "RUNNING"
	TenantOpSuccessful                       = "SUCCESSFUL"
	TenantOpFailed                           = "FAILED"
)
