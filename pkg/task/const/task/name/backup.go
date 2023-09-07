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

package name

const (
	GetTenantInfo            = "get tenant info"
	ConfigureServerForBackup = "configure server for backup"
	CheckAndSpawnJobs        = "check and spawn jobs"
	StartBackupJob           = "start backup job"
	StopBackupJob            = "stop backup job"
	CleanOldBackupJobs       = "clean old backup jobs"
	PauseBackup              = "pause backup"
	ResumeBackup             = "resume backup"
)

const (
	StartRestoreJob  = "start restore job"
	StartLogReplay   = "start log replay"
	CancelRestoreJob = "cancel restore job"
	ActivateStandby  = "activate standby"
)
