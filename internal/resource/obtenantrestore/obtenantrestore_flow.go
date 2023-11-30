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

package obtenantrestore

import (
	"github.com/oceanbase/ob-operator/api/constants"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func StartRestoreJob() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fStartRestoreFlow,
			Tasks:        []tasktypes.TaskName{tStartRestoreJob},
			TargetStatus: string(constants.RestoreJobRunning),
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: string(constants.RestoreJobFailed),
			},
		},
	}
}

func RestoreAsPrimary() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fRestoreAsPrimaryFlow,
			Tasks:        []tasktypes.TaskName{tActivateStandby},
			TargetStatus: string(constants.RestoreJobSuccessful),
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: string(constants.RestoreJobFailed),
			},
		},
	}
}

func RestoreAsStandby() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fRestoreAsStandbyFlow,
			Tasks:        []tasktypes.TaskName{tStartLogReplay},
			TargetStatus: string(constants.RestoreJobSuccessful),
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: string(constants.RestoreJobFailed),
			},
		},
	}
}
