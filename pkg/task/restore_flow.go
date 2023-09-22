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
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
	"github.com/oceanbase/ob-operator/pkg/task/strategy"
)

func StartRestoreJob() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.StartRestoreFlow,
			Tasks:        []string{taskname.StartBackupJob},
			TargetStatus: string(constants.RestoreJobRunning),
			OnFailure: strategy.FailureRule{
				NextTryStatus: string(constants.RestoreJobStarting),
			},
		},
	}
}

func RestoreAsPrimary() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.RestoreAsPrimaryFlow,
			Tasks:        []string{taskname.ActivateStandby},
			TargetStatus: string(constants.RestoreJobSuccessful),
			OnFailure: strategy.FailureRule{
				NextTryStatus: string(constants.RestoreJobFailed),
			},
		},
	}
}

func RestoreAsStandby() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.RestoreAsStandbyFlow,
			Tasks:        []string{taskname.StartLogReplay},
			TargetStatus: string(constants.RestoreJobSuccessful),
			OnFailure: strategy.FailureRule{
				NextTryStatus: string(constants.RestoreJobFailed),
			},
		},
	}
}
