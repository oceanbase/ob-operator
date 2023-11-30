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

package task

import (
	"github.com/oceanbase/ob-operator/api/constants"
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
	"github.com/oceanbase/ob-operator/pkg/task/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func ChangeTenantRootPassword() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: flowname.ChangeTenantRootPasswordFlow,
			Tasks: []string{
				taskname.OpChangeTenantRootPassword,
			},
			TargetStatus: string(constants.TenantOpSuccessful),
			OnFailure: strategy.FailureRule{
				NextTryStatus: string(constants.TenantOpFailed),
			},
		},
	}
}

func ActivateStandbyTenantOp() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: flowname.ActivateStandbyTenantFlow,
			Tasks: []string{
				taskname.OpActivateStandby,
				taskname.OpCreateUsersForActivatedStandby,
			},
			TargetStatus: string(constants.TenantOpSuccessful),
			OnFailure: strategy.FailureRule{
				NextTryStatus: string(constants.TenantOpFailed),
			},
		},
	}
}

func SwitchoverTenants() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: flowname.SwitchoverTenantsFlow,
			Tasks: []string{
				taskname.OpSwitchTenantsRole,
				taskname.OpSetTenantLogRestoreSource,
			},
			TargetStatus: string(constants.TenantOpSuccessful),
			OnFailure: strategy.FailureRule{
				NextTryStatus: string(constants.TenantOpReverting),
			},
		},
	}
}

func RevertSwitchoverTenants() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: flowname.RevertSwitchoverTenantsFlow,
			Tasks: []string{
				taskname.OpSwitchTenantsRole,
			},
			TargetStatus: string(constants.TenantOpFailed),
			OnFailure: strategy.FailureRule{
				NextTryStatus: string(constants.TenantOpReverting),
			},
		},
	}
}

func UpgradeTenant() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: flowname.OpUpgradeTenant,
			Tasks: []string{
				taskname.OpUpgradeTenant,
			},
			TargetStatus: string(constants.TenantOpSuccessful),
			OnFailure: strategy.FailureRule{
				NextTryStatus: string(constants.TenantOpFailed),
			},
		},
	}
}

func ReplayLogOfStandby() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: flowname.OpReplayLog,
			Tasks: []string{
				taskname.OpReplayLog,
			},
			TargetStatus: string(constants.TenantOpSuccessful),
			OnFailure: strategy.FailureRule{
				NextTryStatus: string(constants.TenantOpFailed),
			},
		},
	}
}
