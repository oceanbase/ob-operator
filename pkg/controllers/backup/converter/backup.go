/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package converter

import (
	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"github.com/oceanbase/ob-operator/pkg/controllers/backup/model"
)

func BackupSetListToStatus(backupCurrent cloudv1.Backup, backupSetList []model.AllBackupSet) cloudv1.Backup {
	var backupCurrentStatus cloudv1.BackupStatus
	backupCurrentStatus.BackupSet = backupSetList
	backupCurrentStatus.Schedule = backupCurrent.Spec.Schedule

}

func GenerateBackupSetStatusList(backupCurrent cloudv1.Backup, backupSetList []model.AllBackupSet) []cloudv1.BackupSetStatus {
	backupSetStatusList := make([]cloudv1.)
	var backupSetStatus cloudv1.BackupSetStatus
	for _, backupSet := range backupSetList {
		backupSetStatus.BSKey = int(backupSet.BSKey)
		backupSetStatus.BackupType = backupSet.BackupType
		backupSetStatus.Status = backupSet.Status
		backupSetStatus.TenantID = int(backupSet.TenantID)
		
	}

}
