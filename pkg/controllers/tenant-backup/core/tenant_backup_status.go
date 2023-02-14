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
	"github.com/oceanbase/ob-operator/pkg/controllers/tenant-backup/model"
	"reflect"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
)

func (ctrl *TenantBackupCtrl) UpdateBackupStatus(tenantBackupType string) error {
	tenantBackup := ctrl.TenantBackup
	tenantBackupExecuter := resource.NewBackupResource(ctrl.Resource)
	tenantBackupTmp, err := tenantBackupExecuter.Get(context.TODO(), tenantBackup.Namespace, tenantBackup.Name)
	if err != nil {
		return err
	}
	tenantBackupCurrent := tenantBackupTmp.(cloudv1.TenantBackup)
	tenantBackupCurrentDeepCopy := tenantBackupCurrent.DeepCopy()

	ctrl.TenantBackup = *tenantBackupCurrentDeepCopy
	tenantBackupNew, err := ctrl.buildTenantBackupStatus(*tenantBackupCurrentDeepCopy, tenantBackupType)
	if err != nil {
		return err
	}
	compareStatus := reflect.DeepEqual(tenantBackupCurrent.Status, tenantBackupNew.Status)
	if !compareStatus {
		err = tenantBackupExecuter.UpdateStatus(context.TODO(), tenantBackupNew)
		if err != nil {
			return err
		}
	}
	ctrl.TenantBackup = tenantBackupNew
	return nil
}

func (ctrl *TenantBackupCtrl) buildBackupStatus(tenantBackup cloudv1.TenantBackup, tenantBackupType string) (cloudv1.TenantBackup, error) {
	var tenantBackupCurrentStatus cloudv1.TenantBackupStatus
	tenantBackupJobList, err := ctrl.buildBackupJobListFromDB()
	if err != nil {
		return tenantBackup, err
	}
	tenantBackupJobStatus := ctrl.TenantBackupJobListToStatusList(tenantBackupJobList)
	tenantBackupScheduleList, err := ctrl.buildScheduleList(tenantBackupType)
	if err != nil {
		return tenantBackup, err
	}
	tenantBackupCurrentStatus.TenantBackupJob = tenantBackupJobStatus
	tenantBackupCurrentStatus.Schedule = tenantBackupScheduleList
	tenantBackup.Status = tenantBackupCurrentStatus
	return tenantBackup, nil
}

func (ctrl *TenantBackupCtrl) buildBackupJobListFromDB() ([]model.AllBackupJob, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return nil, err
	}
	return sqlOperator.GetAllBackupJob(), nil
}

func (ctrl *TenantBackupCtrl) TenantBackupJobListToStatusList(backupJobList []model.AllBackupJob) []cloudv1.TenantBackupJobStatus {
	backupJobStatusList := make([]cloudv1.TenantBackupJobStatus, 0)
	for _, backupJob := range backupJobList {
		backupJobStatus := cloudv1.TenantBackupJobStatus{}
		backupJobStatus.ClusterName = ctrl.TenantBackup.Spec.SourceCluster.ClusterName
		backupJobStatus.BackupType = backupJob.BackupType
		backupJobStatus.BackupSetId = int(backupJob.BackupSetId)
		backupJobStatus.TenantID = int(backupJob.TenantId)
		backupJobStatus.Status = backupJob.Status
		backupJobStatusList = append(backupJobStatusList, backupJobStatus)
	}
	return backupJobStatusList
}

func (ctrl *TenantBackupCtrl) buildScheduleList(backupType string) ([]cloudv1.ScheduleSpec, error) {
	tenantList := ctrl.TenantBackup.Spec.Tenants
	for _, tenant := range tenantList {

	}
	scheduleSpec := ctrl.TenantBackup.Spec.Schedule
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
