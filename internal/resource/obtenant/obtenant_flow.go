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

func genCreateTenantFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "create tenant",
			Tasks: []tasktypes.TaskName{
				tCheckTenant,
				tCheckPoolAndConfig,
				tCreateResourcePoolAndConfig,
				tCreateTenantWithClear,
				tCreateUserWithCredentialSecrets,
				tOptimizeTenantByScenario,
			},
			TargetStatus: tenantstatus.Running,
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: tenantstatus.Failed,
			},
		},
	}
}

func genMaintainWhiteListFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "maintain white list",
			Tasks:        []tasktypes.TaskName{tCheckAndApplyWhiteList},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func genMaintainCharsetFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "maintain charset",
			Tasks:        []tasktypes.TaskName{tCheckAndApplyCharset},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func genMaintainUnitNumFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "maintain unit num",
			Tasks:        []tasktypes.TaskName{tCheckAndApplyUnitNum},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func genMaintainLocalityFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "maintain locality",
			Tasks:        []tasktypes.TaskName{tCheckAndApplyLocality},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func genMaintainPrimaryZoneFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "maintain primary zone",
			Tasks:        []tasktypes.TaskName{tCheckAndApplyPrimaryZone},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func genAddPoolFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "add pool",
			Tasks:        []tasktypes.TaskName{tCheckPoolAndConfig, tAddPool},
			TargetStatus: tenantstatus.Running,
			OnFailure: tasktypes.FailureRule{
				Strategy: strategy.RetryFromCurrent,
			},
		},
	}
}

func genDeletePoolFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "delete pool",
			Tasks:        []tasktypes.TaskName{tDeletePool},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func genMaintainUnitConfigFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "maintain unit config",
			Tasks:        []tasktypes.TaskName{tMaintainUnitConfig},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func genDeleteTenantFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "delete tenant",
			Tasks:        []tasktypes.TaskName{tDeleteTenant},
			TargetStatus: tenantstatus.FinalizerFinished,
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: tenantstatus.DeletingTenant,
			},
		},
	}
}

func genRestoreTenantFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "restore tenant",
			Tasks: []tasktypes.TaskName{
				tCheckTenant,
				tCheckPoolAndConfig,
				tCreateResourcePoolAndConfig,
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

func genCancelRestoreFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "cancel restore",
			Tasks: []tasktypes.TaskName{
				tCancelTenantRestoreJob,
			},
			TargetStatus: tenantstatus.RestoreCanceled,
		},
	}
}

func genCreateEmptyStandbyTenantFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "create empty standby tenant",
			Tasks: []tasktypes.TaskName{
				tCheckPrimaryTenantLsIntegrity,
				tCheckTenant,
				tCheckPoolAndConfig,
				tCreateResourcePoolAndConfig,
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

func genMaintainTenantParametersFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "maintain tenant parameters",
			Tasks: []tasktypes.TaskName{
				tMaintainTenantParameters,
			},
			TargetStatus: tenantstatus.Running,
		},
	}
}
