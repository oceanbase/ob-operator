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
			Name: fCreateTenant,
			Tasks: []tasktypes.TaskName{
				tCheckTenant,
				tCheckPoolAndConfig,
				tCreateResourcePoolAndConfig,
				tCreateTenantWithClear,
				tCreateUserWithCredentialSecrets,
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
			Name:         fMaintainWhiteList,
			Tasks:        []tasktypes.TaskName{tCheckAndApplyWhiteList},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func genMaintainCharsetFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fMaintainCharset,
			Tasks:        []tasktypes.TaskName{tCheckAndApplyCharset},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func genMaintainUnitNumFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fMaintainUnitNum,
			Tasks:        []tasktypes.TaskName{tCheckAndApplyUnitNum},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func genMaintainLocalityFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fMaintainLocality,
			Tasks:        []tasktypes.TaskName{tCheckAndApplyLocality},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func genMaintainPrimaryZoneFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fMaintainPrimaryZone,
			Tasks:        []tasktypes.TaskName{tCheckAndApplyPrimaryZone},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func genAddPoolFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fAddPool,
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
			Name:         fDeletePool,
			Tasks:        []tasktypes.TaskName{tDeletePool},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func genMaintainUnitConfigFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fMaintainUnitConfig,
			Tasks:        []tasktypes.TaskName{tMaintainUnitConfig},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func genDeleteTenantFlow(_ *OBTenantManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fDeleteTenant,
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
			Name: fRestoreTenant,
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
			Name: fCancelRestore,
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
			Name: fCreateEmptyStandbyTenant,
			Tasks: []tasktypes.TaskName{
				tCheckPrimaryTenantLSIntegrity,
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
