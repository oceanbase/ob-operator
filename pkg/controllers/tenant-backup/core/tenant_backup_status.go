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
	"strings"

	"github.com/oceanbase/ob-operator/pkg/controllers/tenant-backup/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/tenant-backup/model"
	"k8s.io/klog/v2"

	"reflect"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
)

func (ctrl *TenantBackupCtrl) UpdateBackupStatus(tenant cloudv1.TenantConfigSpec, tenantBackupType string) error {
	tenantBackup := ctrl.TenantBackup
	tenantBackupExecuter := resource.NewTenantBackupResource(ctrl.Resource)
	tenantBackupTmp, err := tenantBackupExecuter.Get(context.TODO(), tenantBackup.Namespace, tenantBackup.Name)
	if err != nil {
		return err
	}
	tenantBackupCurrent := tenantBackupTmp.(cloudv1.TenantBackup)
	tenantBackupCurrentDeepCopy := tenantBackupCurrent.DeepCopy()

	ctrl.TenantBackup = *tenantBackupCurrentDeepCopy
	tenantBackupNew, err := ctrl.buildTenantBackupStatus(*tenantBackupCurrentDeepCopy, tenant, tenantBackupType)
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

func (ctrl *TenantBackupCtrl) buildTenantBackupStatus(tenantBackup cloudv1.TenantBackup, tenant cloudv1.TenantConfigSpec, tenantBackupType string) (cloudv1.TenantBackup, error) {
	var tenantBackupCurrentStatus cloudv1.TenantBackupStatus
	var tenantBackupSet []cloudv1.TenantBackupSetStatus
	for _, status := range ctrl.TenantBackup.Status.TenantBackupSet {
		if status.TenantName == "" {
			continue
		}
		exist := false
		for _, spec := range ctrl.TenantBackup.Spec.Tenants {
			if spec.Name == status.TenantName {
				exist = true
			}
		}
		if !exist {
			tenantBackupSet = append(tenantBackupSet, status)
		}
	}
	tenantList := ctrl.TenantBackup.Spec.Tenants
	for _, t := range tenantList {
		if t.Name == tenant.Name {
			tenantBackupStatus, err := ctrl.buildSingleTenantBackupStatus(tenant, tenantBackupType)
			if err != nil {
				klog.Errorf("Build tenant '%s' backup status error '%s'", tenant.Name, err)
				return tenantBackup, err
			}
			tenantBackupSet = append(tenantBackupSet, tenantBackupStatus)
		} else {
			tenantBackupStatus := ctrl.GetSingleTenantBackupStatus(t)
			tenantBackupSet = append(tenantBackupSet, tenantBackupStatus)
		}
	}
	tenantBackupCurrentStatus.TenantBackupSet = tenantBackupSet
	tenantBackup.Status = tenantBackupCurrentStatus
	return tenantBackup, nil
}

func (ctrl *TenantBackupCtrl) buildSingleTenantBackupStatus(tenant cloudv1.TenantConfigSpec, tenantBackupType string) (cloudv1.TenantBackupSetStatus, error) {
	var tenantBackupSetStatus cloudv1.TenantBackupSetStatus
	var err error
	backupJobList, err := ctrl.buildBackupJobListFromDB(tenant.Name)
	if err != nil {
		return tenantBackupSetStatus, err
	}
	scheduleList, err := ctrl.buildScheduleList(tenant, tenantBackupType)
	if err != nil {
		return tenantBackupSetStatus, err
	}
	tenantBackupSetStatus.TenantName = tenant.Name
	tenantBackupSetStatus.ClusterName = ctrl.TenantBackup.Spec.SourceCluster.ClusterName
	tenantBackupSetStatus.BackupJobs = ctrl.TenantBackupJobListToStatusList(backupJobList)
	tenantBackupSetStatus.Schedule = scheduleList
	return tenantBackupSetStatus, nil
}

func (ctrl *TenantBackupCtrl) buildBackupJobListFromDB(name string) ([]model.BackupJob, error) {
	res := make([]model.BackupJob, 0)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		klog.Errorf("Get tenant '%s' sql operator error '%s'", name, err)
		return nil, err
	}
	backupDatabaseJob := sqlOperator.GetBackupDatabaseJob(name)
	backupIncrementalJob := sqlOperator.GetBackupIncrementalJob(name)
	backupDatabaseJobHistory := sqlOperator.GetBackupDatabaseJobHistory(name)
	backupIncrementalJobHistory := sqlOperator.GetBackupIncrementalJobHistory(name)
	if len(backupDatabaseJob) != 0 {
		res = append(res, backupDatabaseJob...)
	} else {
		res = append(res, backupDatabaseJobHistory...)
	}
	if len(backupIncrementalJob) != 0 {
		res = append(res, backupIncrementalJob...)
	} else {
		res = append(res, backupIncrementalJobHistory...)
	}
	return res, nil
}

func (ctrl *TenantBackupCtrl) TenantBackupJobListToStatusList(backupJobList []model.BackupJob) []cloudv1.BackupJobStatus {
	backupJobStatusList := make([]cloudv1.BackupJobStatus, 0)
	for _, backupJob := range backupJobList {
		backupJobStatus := cloudv1.BackupJobStatus{}
		backupJobStatus.BackupSetId = int(backupJob.BackupSetId)
		backupJobStatus.BackupType = backupJob.BackupType
		backupJobStatus.Status = backupJob.Status
		backupJobStatusList = append(backupJobStatusList, backupJobStatus)
	}
	return backupJobStatusList
}

func (ctrl *TenantBackupCtrl) buildScheduleList(tenant cloudv1.TenantConfigSpec, backupType string) ([]cloudv1.ScheduleSpec, error) {
	scheduleSpec := tenant.Schedule
	backupScheduleList := make([]cloudv1.ScheduleSpec, 0)
	for _, schedule := range scheduleSpec {
		var backupSchedule cloudv1.ScheduleSpec
		if strings.ToUpper(schedule.BackupType) == tenantBackupconst.FullBackup || strings.ToUpper(schedule.BackupType) == tenantBackupconst.FullBackupType {
			backupSchedule.BackupType = tenantBackupconst.FullBackupType
			backupSchedule.Schedule = schedule.Schedule
			if schedule.Schedule != tenantBackupconst.BackupOnce && backupSchedule.Schedule != "" {
				if backupType == tenantBackupconst.FullBackupType || backupType == "" {
					nextTime, err := ctrl.getNextCron(schedule.Schedule)
					if err != nil {
						klog.Errorf("Get tenant '%s' next time of backup type full error '%s'", tenant.Name, err)
						return scheduleSpec, err
					}
					backupSchedule.NextTime = nextTime.String()
				}
			} else {
				backupSchedule.NextTime = ""
			}
		}
		if strings.ToUpper(schedule.BackupType) == tenantBackupconst.IncrementalBackup || strings.ToUpper(schedule.BackupType) == tenantBackupconst.IncrementalBackupType {
			backupSchedule.BackupType = tenantBackupconst.IncrementalBackupType
			backupSchedule.Schedule = schedule.Schedule
			if schedule.Schedule != tenantBackupconst.BackupOnce && backupSchedule.Schedule != "" {
				if backupType == tenantBackupconst.IncrementalBackupType || backupType == "" {
					nextTime, err := ctrl.getNextCron(schedule.Schedule)
					if err != nil {
						klog.Errorf("Get tenant '%s' next time of backup type incremental error '%s'", tenant.Name, err)
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

func (ctrl *TenantBackupCtrl) DeleteSingleTenantStatus(name string) error {
	tenantBackup := ctrl.TenantBackup
	tenantBackupExecuter := resource.NewTenantBackupResource(ctrl.Resource)
	tenantBackupTmp, err := tenantBackupExecuter.Get(context.TODO(), tenantBackup.Namespace, tenantBackup.Name)
	if err != nil {
		return err
	}
	tenantBackupCurrent := tenantBackupTmp.(cloudv1.TenantBackup)
	tenantBackupCurrentDeepCopy := tenantBackupCurrent.DeepCopy()

	ctrl.TenantBackup = *tenantBackupCurrentDeepCopy
	tenantBackupNew, err := ctrl.DeleteStatus(*tenantBackupCurrentDeepCopy, name)
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

func (ctrl *TenantBackupCtrl) DeleteStatus(tenantBackup cloudv1.TenantBackup, name string) (cloudv1.TenantBackup, error) {
	var tenantBackupCurrentStatus cloudv1.TenantBackupStatus
	var tenantBackupSet []cloudv1.TenantBackupSetStatus
	for _, status := range ctrl.TenantBackup.Status.TenantBackupSet {
		if status.TenantName != name {
			tenantBackupSet = append(tenantBackupSet, status)
		}
	}
	tenantBackupCurrentStatus.TenantBackupSet = tenantBackupSet
	tenantBackup.Status = tenantBackupCurrentStatus
	return tenantBackup, nil
}
