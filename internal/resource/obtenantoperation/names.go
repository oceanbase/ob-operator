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

package obtenantoperation

import (
	ttypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

// tenant operation flows
const (
	fChangeTenantRootPasswordFlow ttypes.FlowName = "change tenant root password"
	fActivateStandbyTenantFlow    ttypes.FlowName = "activate standby tenant"
	fSwitchoverTenantsFlow        ttypes.FlowName = "switchover tenants"
	fRevertSwitchoverTenantsFlow  ttypes.FlowName = "revert switchover tenants"
	fOpUpgradeTenant              ttypes.FlowName = "upgrade tenant"
	fOpReplayLog                  ttypes.FlowName = "replay log"
)

const (
	tChangeTenantRootPassword       ttypes.TaskName = "change tenant root password"
	tActivateStandbyTenant          ttypes.TaskName = "activate standby"
	tCreateUsersForActivatedStandby ttypes.TaskName = "create users for activated standby"
	tSwitchTenantsRole              ttypes.TaskName = "switch tenants role"
	tSetTenantLogRestoreSource      ttypes.TaskName = "set tenant log restore source"
	tUpgradeTenant                  ttypes.TaskName = "upgrade tenant"
	tReplayLogOfStandby             ttypes.TaskName = "replay log"
)
