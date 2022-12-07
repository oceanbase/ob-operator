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

package core

import (
	"context"
	"reflect"
	"time"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"

	backupconst "github.com/oceanbase/ob-operator/pkg/controllers/backup/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/backup/model"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
)

func (ctrl *BackupCtrl) UpdateBackupStatus(backupType string) error {
	backup := ctrl.Backup
	backupExecuter := resource.NewBackupResource(ctrl.Resource)
	backupTmp, err := backupExecuter.Get(context.TODO(), backup.Namespace, backup.Name)
	if err != nil {
		return err
	}
	backupCurrent := backupTmp.(cloudv1.Backup)
	backupCurrentDeepCopy := backupCurrent.DeepCopy()

	ctrl.Backup = *backupCurrentDeepCopy
	backupNew, err := ctrl.buildBackupStatus(*backupCurrentDeepCopy, backupType)
	if err != nil {
		return err
	}
	compareStatus := reflect.DeepEqual(backupCurrent.Status, backupNew.Status)
	if !compareStatus {
		err = backupExecuter.UpdateStatus(context.TODO(), backupNew)
		if err != nil {
			return err
		}
	}
	ctrl.Backup = backupNew
	return nil
}

func (ctrl *BackupCtrl) buildBackupStatus(backup cloudv1.Backup, backupType string) (cloudv1.Backup, error) {
	var backupCurrentStatus cloudv1.BackupStatus
	backupSetList, err := ctrl.buildBackupSetListFromDB()
	if err != nil {
		return backup, err
	}
	backupSetStatus := ctrl.BackupSetListToStatusList(backupSetList)
	backupScheduleList, err := ctrl.buildScheduleList(backupType)
	if err != nil {
		return backup, err
	}
	backupCurrentStatus.BackupSet = backupSetStatus
	backupCurrentStatus.Schedule = backupScheduleList
	backup.Status = backupCurrentStatus
	return backup, nil
}

func (ctrl *BackupCtrl) buildScheduleList(backupType string) ([]cloudv1.ScheduleSpec, error) {
	scheduleSpec := ctrl.Backup.Spec.Schedule
	backupScheduleList := make([]cloudv1.ScheduleSpec, 0)
	for _, schedule := range scheduleSpec {
		var backupSchedule cloudv1.ScheduleSpec
		if schedule.BackupType == backupconst.FullBackup {
			backupSchedule.BackupType = backupconst.FullBackupType
			backupSchedule.Schedule = schedule.Schedule
			if schedule.Schedule != backupconst.BackupOnce && backupSchedule.Schedule != "" {
				if backupType == backupconst.FullBackupType || backupType == "" {
					nextTime, err := ctrl.getNextCron(schedule.Schedule)
					if err != nil {
						return scheduleSpec, err
					}
					backupSchedule.NextTime = nextTime.String()
				}
			} else {
				backupSchedule.NextTime = ""
			}
		}
		if schedule.BackupType == backupconst.IncrementalBackup {
			backupSchedule.BackupType = backupconst.IncrementalBackupType
			backupSchedule.Schedule = schedule.Schedule
			if schedule.Schedule != backupconst.BackupOnce && backupSchedule.Schedule != "" {
				if backupType == backupconst.IncDatabaseBackupType || backupType == "" {
					nextTime, err := ctrl.getNextCron(schedule.Schedule)
					if err != nil {
						return scheduleSpec, err
					}
					backupSchedule.NextTime = nextTime.String()
				}
			} else {
				backupSchedule.NextTime = ""
			}
		}
		backupScheduleList = append(backupScheduleList, backupSchedule)
	}
	return backupScheduleList, nil

}

func (ctrl *BackupCtrl) BackupSetListToStatusList(backupSetList []model.AllBackupSet) []cloudv1.BackupSetStatus {
	backupSetStatusList := make([]cloudv1.BackupSetStatus, 0)
	for _, backupSet := range backupSetList {
		backupSetStatus := cloudv1.BackupSetStatus{}
		backupSetStatus.ClusterName = ctrl.Backup.Spec.SourceCluster.ClusterName
		backupSetStatus.BackupType = backupSet.BackupType
		backupSetStatus.BSKey = int(backupSet.BSKey)
		backupSetStatus.TenantID = int(backupSet.TenantID)
		backupSetStatus.Status = backupSet.Status
		backupSetStatusList = append(backupSetStatusList, backupSetStatus)
	}
	return backupSetStatusList
}

func (ctrl *BackupCtrl) buildBackupSetListFromDB() ([]model.AllBackupSet, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return nil, err
	}
	return sqlOperator.GetAllBackupSet(), nil
}

func (ctrl *BackupCtrl) UpdateBackupScheduleStatus(next time.Time, backupType string) error {

	schedule := ctrl.Backup.Status.Schedule
	for _, scheduleSpec := range schedule {
		if scheduleSpec.BackupType == backupType {
			scheduleSpec.NextTime = next.String()
		}
	}
	return nil
}
