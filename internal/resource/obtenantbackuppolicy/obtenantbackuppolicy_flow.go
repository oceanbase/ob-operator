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

package obtenantbackuppolicy

import (
	"github.com/oceanbase/ob-operator/api/constants"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func PrepareBackupPolicy() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fPrepareBackupPolicy,
			Tasks:        []tasktypes.TaskName{tConfigureServerForBackup},
			TargetStatus: string(constants.BackupPolicyStatusPrepared),
		},
	}
}

func StartBackupJob() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fStartBackupJob,
			Tasks:        []tasktypes.TaskName{tStartBackupJob},
			TargetStatus: string(constants.BackupPolicyStatusRunning),
		},
	}
}

func StopBackupPolicy() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fStopBackupPolicy,
			Tasks:        []tasktypes.TaskName{tStopBackupPolicy},
			TargetStatus: string(constants.BackupPolicyStatusStopped),
		},
	}
}

func MaintainRunningPolicy() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fMaintainRunningPolicy,
			Tasks:        []tasktypes.TaskName{tConfigureServerForBackup, tCleanOldBackupJobs, tCheckAndSpawnJobs},
			TargetStatus: string(constants.BackupPolicyStatusRunning),
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: string(constants.BackupPolicyStatusRunning),
			},
		},
	}
}

func PauseBackup() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fPauseBackup,
			Tasks:        []tasktypes.TaskName{tPauseBackup},
			TargetStatus: string(constants.BackupPolicyStatusPaused),
		},
	}
}

func ResumeBackup() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fResumeBackup,
			Tasks:        []tasktypes.TaskName{tResumeBackup},
			TargetStatus: string(constants.BackupPolicyStatusRunning),
		},
	}
}
