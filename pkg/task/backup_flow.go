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

func PrepareBackupPolicy() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.PrepareBackupPolicy,
			Tasks:        []string{taskname.ConfigureServerForBackup},
			TargetStatus: string(constants.BackupPolicyStatusPrepared),
		},
	}
}

func StartBackupJob() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.StartBackupJob,
			Tasks:        []string{taskname.StartBackupJob},
			TargetStatus: string(constants.BackupPolicyStatusRunning),
		},
	}
}

func StopBackupJob() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.StopBackupJob,
			Tasks:        []string{taskname.StopBackupJob},
			TargetStatus: string(constants.BackupPolicyStatusStopped),
		},
	}
}

func MaintainRunningPolicy() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.MaintainRunningPolicy,
			Tasks:        []string{taskname.ConfigureServerForBackup, taskname.CleanOldBackupJobs, taskname.CheckAndSpawnJobs},
			TargetStatus: string(constants.BackupPolicyStatusRunning),
			OnFailure: strategy.FailureRule{
				NextTryStatus: string(constants.BackupPolicyStatusFailed),
			},
		},
	}
}

func PauseBackup() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.PauseBackup,
			Tasks:        []string{taskname.PauseBackup},
			TargetStatus: string(constants.BackupPolicyStatusPaused),
		},
	}
}

func ResumeBackup() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.ResumeBackup,
			Tasks:        []string{taskname.ResumeBackup},
			TargetStatus: string(constants.BackupPolicyStatusRunning),
		},
	}
}
