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
	"github.com/oceanbase/ob-operator/api/constants"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func genChangeTenantRootPasswordFlow(_ *ObTenantOperationManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: fChangeTenantRootPasswordFlow,
			Tasks: []tasktypes.TaskName{
				tChangeTenantRootPassword,
			},
			TargetStatus: string(constants.TenantOpSuccessful),
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: string(constants.TenantOpFailed),
			},
		},
	}
}

func genActivateStandbyTenantOpFlow(_ *ObTenantOperationManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: fActivateStandbyTenantFlow,
			Tasks: []tasktypes.TaskName{
				tActivateStandbyTenant,
				tCreateUsersForActivatedStandby,
			},
			TargetStatus: string(constants.TenantOpSuccessful),
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: string(constants.TenantOpFailed),
			},
		},
	}
}

func genSwitchoverTenantsFlow(_ *ObTenantOperationManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: fSwitchoverTenantsFlow,
			Tasks: []tasktypes.TaskName{
				tSwitchTenantsRole,
				tSetTenantLogRestoreSource,
			},
			TargetStatus: string(constants.TenantOpSuccessful),
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: string(constants.TenantOpReverting),
			},
		},
	}
}

func genRevertSwitchoverTenantsFlow(_ *ObTenantOperationManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: fRevertSwitchoverTenantsFlow,
			Tasks: []tasktypes.TaskName{
				tSwitchTenantsRole,
			},
			TargetStatus: string(constants.TenantOpFailed),
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: string(constants.TenantOpReverting),
			},
		},
	}
}

func genUpgradeTenantFlow(_ *ObTenantOperationManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: fOpUpgradeTenant,
			Tasks: []tasktypes.TaskName{
				tUpgradeTenant,
			},
			TargetStatus: string(constants.TenantOpSuccessful),
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: string(constants.TenantOpFailed),
			},
		},
	}
}

func genReplayLogOfStandbyFlow(_ *ObTenantOperationManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: fOpReplayLog,
			Tasks: []tasktypes.TaskName{
				tReplayLogOfStandby,
			},
			TargetStatus: string(constants.TenantOpSuccessful),
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: string(constants.TenantOpFailed),
			},
		},
	}
}
