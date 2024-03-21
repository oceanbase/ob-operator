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
	tenantstatus "github.com/oceanbase/ob-operator/internal/const/status/tenantstatus"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func CreateTenant() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: fCreateTenant,
			Tasks: []tasktypes.TaskName{
				tCheckTenantTask,
				tCheckPoolAndConfigTask,
				tCreateResourcePoolAndConfigTask,
				tCreateTenantTaskWithClear,
				tCreateUserWithCredentialSecrets,
			},
			TargetStatus: tenantstatus.Running,
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: tenantstatus.Failed,
			},
		},
	}
}

func MaintainWhiteList() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fMaintainWhiteList,
			Tasks:        []tasktypes.TaskName{tCheckAndApplyWhiteList},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func MaintainCharset() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fMaintainCharset,
			Tasks:        []tasktypes.TaskName{tCheckAndApplyCharset},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func MaintainUnitNum() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fMaintainUnitNum,
			Tasks:        []tasktypes.TaskName{tCheckAndApplyUnitNum},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func MaintainLocality() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fMaintainLocality,
			Tasks:        []tasktypes.TaskName{tCheckAndApplyLocality},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func MaintainPrimaryZone() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fMaintainPrimaryZone,
			Tasks:        []tasktypes.TaskName{tCheckAndApplyPrimaryZone},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func AddPool() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fAddPool,
			Tasks:        []tasktypes.TaskName{tCheckPoolAndConfigTask, tAddPoolTask},
			TargetStatus: tenantstatus.Running,
			OnFailure: tasktypes.FailureRule{
				Strategy: strategy.RetryFromCurrent,
			},
		},
	}
}

func DeletePool() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fDeletePool,
			Tasks:        []tasktypes.TaskName{tDeletePoolTask},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func MaintainUnitConfig() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fMaintainUnitConfig,
			Tasks:        []tasktypes.TaskName{tMaintainUnitConfigTask},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func DeleteTenant() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fDeleteTenant,
			Tasks:        []tasktypes.TaskName{tDeleteTenantTask},
			TargetStatus: tenantstatus.FinalizerFinished,
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: tenantstatus.DeletingTenant,
			},
		},
	}
}

func RestoreTenant() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: fRestoreTenant,
			Tasks: []tasktypes.TaskName{
				tCheckTenantTask,
				tCheckPoolAndConfigTask,
				tCreateResourcePoolAndConfigTask,
				tCreateTenantRestoreJobCR,
				tWatchRestoreJobToFinish,
				tCheckAndApplyWhiteList,
				tCreateUserWithCredentialSecrets,
			},
			TargetStatus: tenantstatus.Running,
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: tenantstatus.RestoreFailed,
			},
		},
	}
}

func FlowCancelRestore() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: fCancelRestoreFlow,
			Tasks: []tasktypes.TaskName{
				tCancelTenantRestoreJob,
			},
			TargetStatus: tenantstatus.RestoreCanceled,
		},
	}
}

func FlowCreateEmptyStandbyTenant() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: fCreateEmptyStandbyTenant,
			Tasks: []tasktypes.TaskName{
				tCheckPrimaryTenantLSIntegrity,
				tCheckTenantTask,
				tCheckPoolAndConfigTask,
				tCreateResourcePoolAndConfigTask,
				tCreateEmptyStandbyTenant,
				tCheckAndApplyWhiteList,
			},
			TargetStatus: tenantstatus.Running,
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: tenantstatus.Failed,
			},
		},
	}
}
