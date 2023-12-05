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
	ttypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

const (
	fPrepareBackupPolicy   ttypes.FlowName = "prepare backup policy"
	fStartBackupJob        ttypes.FlowName = "start backup job"
	fStopBackupPolicy      ttypes.FlowName = "stop backup policy"
	fMaintainRunningPolicy ttypes.FlowName = "maintain running policy"
	fPauseBackup           ttypes.FlowName = "pause backup"
	fResumeBackup          ttypes.FlowName = "resume backup"
)

const (
	tConfigureServerForBackup ttypes.TaskName = "configure server for backup"
	tCheckAndSpawnJobs        ttypes.TaskName = "check and spawn jobs"
	tStartBackupJob           ttypes.TaskName = "start backup job"
	tStopBackupPolicy         ttypes.TaskName = "stop backup policy"
	tCleanOldBackupJobs       ttypes.TaskName = "clean old backup jobs"
	tPauseBackup              ttypes.TaskName = "pause backup"
	tResumeBackup             ttypes.TaskName = "resume backup"
	tDeleteBackupPolicy       ttypes.TaskName = "delete backup policy"
)
