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
	"fmt"
	"strings"
	"time"

	"github.com/gorhill/cronexpr"
	"github.com/pkg/errors"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	tenantBackupconst "github.com/oceanbase/ob-operator/pkg/controllers/tenant-backup/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/tenant-backup/model"
	"github.com/oceanbase/ob-operator/pkg/controllers/tenant-backup/sql"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

func (ctrl *TenantBackupCtrl) GetTenantSecret(tenant cloudv1.TenantSpec) (model.TenantSecret, error) {
	var tenantSecret model.TenantSecret
	obcluster := ctrl.TenantBackup.Spec.SourceCluster
	secretExecutor := resource.NewSecretResource(ctrl.Resource)
	userSecret, err := secretExecutor.Get(context.TODO(), obcluster.ClusterNamespace, tenant.UserSecret)
	if err != nil {
		klog.Errorf("get tenant '%s' user secret error '%s'", tenant.Name, err)
		return tenantSecret, err
	}
	backupSecret, err := secretExecutor.Get(context.TODO(), obcluster.ClusterNamespace, tenant.BackupSecret)
	if err != nil {
		klog.Errorf("get tenant '%s' backup secret error '%s'", tenant.Name, err)
		return tenantSecret, err
	}
	tenantSecret.User = strings.Replace(string(userSecret.(corev1.Secret).Data[tenantBackupconst.User]), "\n", "", -1)
	tenantSecret.UserSecret = strings.Replace(string(userSecret.(corev1.Secret).Data[tenantBackupconst.UserSecret]), "\n", "", -1)
	tenantSecret.IncrementalSecret = strings.Replace(string(backupSecret.(corev1.Secret).Data[tenantBackupconst.IncrementalSecret]), "\n", "", -1)
	tenantSecret.DatabaseSecret = strings.Replace(string(backupSecret.(corev1.Secret).Data[tenantBackupconst.DatabaseSecret]), "\n", "", -1)
	return tenantSecret, nil
}

func (ctrl *TenantBackupCtrl) GetTenantSqlOperator(tenant cloudv1.TenantSpec) (*sql.SqlOperator, error) {
	tenantSecret, err := ctrl.GetTenantSecret(tenant)
	if err != nil {
		klog.Errorf("get tenant '%s' secret error '%s'", tenant.Name, err)
		return nil, err
	}
	clusterIP, err := ctrl.GetServiceClusterIPByName(ctrl.TenantBackup.Namespace, ctrl.TenantBackup.Spec.SourceCluster.ClusterName)
	// get svc failed
	if err != nil {
		return nil, errors.New("failed to get service address")
	}
	p := &sql.DBConnectProperties{
		IP:       clusterIP,
		Port:     observerconst.MysqlPort,
		User:     fmt.Sprint(tenantSecret.User, "@", tenant.Name),
		Password: tenantSecret.UserSecret,
		Database: "oceanbase",
		Timeout:  10,
	}
	so := sql.NewSqlOperator(p)
	if so.TestOK() {
		return so, nil
	}
	return nil, errors.New("failed to get tenant sql operator")
}

func (ctrl *TenantBackupCtrl) CheckAndSetLogArchiveDest(tenant cloudv1.TenantSpec) error {
	logArchiveDest, err := ctrl.GetLogArchiveDest(tenant)
	if err != nil {
		klog.Errorf("get tenant '%s' LogArchiveDest error '%s'", tenant.Name, err)
		return err
	}
	if ctrl.NeedSetArchiveDest(tenant, logArchiveDest) {
		return ctrl.SetArchiveDest(tenant)
	}
	return nil
}

func (ctrl *TenantBackupCtrl) SetArchiveDest(tenant cloudv1.TenantSpec) error {
	sqlOperator, err := ctrl.GetTenantSqlOperator(tenant)
	if err != nil {
		klog.Errorf("tenant '%s' get sql operator error when set LogArchiveDest", tenant.Name)
		return errors.Wrap(err, "get sql operator error when set LogArchiveDest")
	}
	value := fmt.Sprint("LOCATION=", tenant.LogArchiveDest)
	if tenant.Binding != "" {
		value = fmt.Sprint(value, " BINDING=", tenant.Binding)
	}
	if tenant.PieceSwitchInterval != "" {
		value = fmt.Sprint(value, " PIECE_SWITCH_INTERVAL=", tenant.PieceSwitchInterval)
	}
	return sqlOperator.SetParameter(tenantBackupconst.LogAechiveDest, value)
}

func (ctrl *TenantBackupCtrl) GetLogArchiveDest(tenant cloudv1.TenantSpec) ([]model.TenantArchiveDest, error) {
	sqlOperator, err := ctrl.GetTenantSqlOperator(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "get sql operator error when get LogArchiveDest")
	}
	return sqlOperator.GetArchiveLogDest(), nil
}

func (ctrl *TenantBackupCtrl) NeedSetArchiveDest(tenant cloudv1.TenantSpec, logArchiveDestList []model.TenantArchiveDest) bool {
	if len(logArchiveDestList) == 0 {
		return true
	}
	for _, logArchiveDest := range logArchiveDestList {
		if (logArchiveDest.Name == tenantBackupconst.Path && logArchiveDest.Value != tenant.LogArchiveDest) ||
			(logArchiveDest.Name == tenantBackupconst.Binding && !strings.EqualFold(strings.ToLower(logArchiveDest.Value), strings.ToLower(tenant.Binding))) ||
			(logArchiveDest.Name == tenantBackupconst.PieceSwitchInterval && !strings.EqualFold(strings.ToLower(logArchiveDest.Value), strings.ToLower(tenant.PieceSwitchInterval))) {
			return true
		}
	}
	return false
}

func (ctrl *TenantBackupCtrl) CheckAndStartArchive(tenant cloudv1.TenantSpec) error {
	archiveLogList, err := ctrl.GetTenantArchiveLog(tenant)
	if err != nil {
		klog.Errorf("get tenant '%s' archive summary list error '%s'", tenant.Name, err)
		return err
	}
	needStartAchiveLog, err := ctrl.NeedStartAchiveLog(tenant, archiveLogList)
	if err != nil {
		return nil
	}
	if needStartAchiveLog {
		return ctrl.StartAchiveLog(tenant)
	}
	return nil
}

func (ctrl *TenantBackupCtrl) GetTenantArchiveLog(tenant cloudv1.TenantSpec) ([]model.TenantArchiveLog, error) {
	sqlOperator, err := ctrl.GetTenantSqlOperator(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "get sql operator error when get ArchiveLog")
	}
	return sqlOperator.GetArchiveLog(), nil
}

func (ctrl *TenantBackupCtrl) NeedStartAchiveLog(tenant cloudv1.TenantSpec, archiveLogList []model.TenantArchiveLog) (bool, error) {
	if len(archiveLogList) == 0 {
		return true, nil
	}
	for _, archiveLog := range archiveLogList {
		if archiveLog.Status == tenantBackupconst.ArchiveLogPrepare || archiveLog.Status == tenantBackupconst.ArchiveLogBeginning || archiveLog.Status == tenantBackupconst.ArchiveLogStopping {
			klog.Infof("Tenant '%s' archivelog status '%s'", tenant.Name, archiveLog.Status)
			return false, errors.Errorf("Tenant '%s' archivelog status '%s'", tenant.Name, archiveLog.Status)
		}
		if archiveLog.Status == tenantBackupconst.ArchiveLogInterrupted {
			klog.Errorf("Tenant '%s' archivelog status '%s'", tenant.Name, archiveLog.Status)
			return false, errors.Errorf("Tenant '%s' archivelog status '%s'", tenant.Name, archiveLog.Status)
		}
		if archiveLog.Status == tenantBackupconst.ArchiveLogStop {
			klog.Infof("Tenant '%s' archivelog status '%s'", tenant.Name, archiveLog.Status)
			return true, nil
		}
		if archiveLog.Status == tenantBackupconst.ArchiveLogDoing {
			return false, nil
		}
	}
	return false, nil
}

func (ctrl *TenantBackupCtrl) StartAchiveLog(tenant cloudv1.TenantSpec) error {
	sqlOperator, err := ctrl.GetTenantSqlOperator(tenant)
	if err != nil {
		return errors.Wrap(err, "get sql operator error when start ArchiveLog")
	}
	return sqlOperator.StartAchiveLog()
}

func (ctrl *TenantBackupCtrl) CheckTenantBackupExist(tenant cloudv1.TenantSpec) (bool, []string) {
	backupTypeList := make([]string, 0)
	backupSets := ctrl.TenantBackup.Status.TenantBackupSet
	exist := false
	for _, backupSet := range backupSets {
		if backupSet.ClusterName == ctrl.TenantBackup.Spec.SourceCluster.ClusterName && backupSet.TenantName == tenant.Name {
			exist = true
			for _, backupJob := range backupSet.BackupJobs {
				backupTypeList = append(backupTypeList, backupJob.BackupType)
			}
		}
	}
	return exist, backupTypeList
}

func (ctrl *TenantBackupCtrl) CheckTenantBackupOnce(tenant cloudv1.TenantSpec, backupTypeList []string) (bool, bool) {
	var backupOnce, finished bool
	for _, schedule := range tenant.Schedule {
		if schedule.Schedule == tenantBackupconst.BackupOnce {
			backupOnce = true
			for _, t := range backupTypeList {
				if t == schedule.BackupType {
					finished = true
				}
			}
		}
	}
	return backupOnce, finished
}

func (ctrl *TenantBackupCtrl) getNextCron(schedule string) (time.Time, error) {
	expr, err := cronexpr.Parse(schedule)
	if err != nil {
		return time.Time{}, err
	}
	nextTime := expr.Next(time.Now())
	return nextTime, nil
}

func (ctrl *TenantBackupCtrl) CheckAndSetBackupDest(tenant cloudv1.TenantSpec) error {
	backupDest, err := ctrl.GetBackupDest(tenant)
	if err != nil {
		klog.Errorf("get tenant '%s' LogArchiveDest error '%s'", tenant.Name, err)
		return err
	}
	if ctrl.NeedSetBackupDest(tenant, backupDest) {
		return ctrl.SetBackupDest(tenant)
	}
	return nil
}

func (ctrl *TenantBackupCtrl) GetBackupDest(tenant cloudv1.TenantSpec) ([]model.TenantBackupDest, error) {
	sqlOperator, err := ctrl.GetTenantSqlOperator(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "get sql operator error when get LogArchiveDest")
	}
	return sqlOperator.GetBackupDest(), nil
}

func (ctrl *TenantBackupCtrl) NeedSetBackupDest(tenant cloudv1.TenantSpec, backupDestList []model.TenantBackupDest) bool {
	if len(backupDestList) == 0 {
		return true
	}
	for _, backupDest := range backupDestList {
		if backupDest.Name == tenantBackupconst.DataBackupDest && backupDest.Value != tenant.DataBackupDest {
			return true
		}
	}
	return false
}

func (ctrl *TenantBackupCtrl) SetBackupDest(tenant cloudv1.TenantSpec) error {
	sqlOperator, err := ctrl.GetTenantSqlOperator(tenant)
	if err != nil {
		klog.Errorf("tenant '%s' get sql operator error when set DataBackupDest", tenant.Name)
		return errors.Wrap(err, "get sql operator error when set DataBackupDest")
	}
	return sqlOperator.SetParameter(tenantBackupconst.DataBackupDest, tenant.DataBackupDest)
}

func (ctrl *TenantBackupCtrl) CheckAndSetBackupDatabasePassword(tenant cloudv1.TenantSpec) error {
	tenantSecret, err := ctrl.GetTenantSecret(tenant)
	if err != nil {
		klog.Errorf("get tenant '%s' secret error '%s'", tenant.Name, err)
		return err
	}
	sqlOperator, err := ctrl.GetTenantSqlOperator(tenant)
	if err != nil {
		klog.Errorf("tenant '%s' get sql operator error when set Backup Database Password", tenant.Name)
		return errors.Wrap(err, "get sql operator error when set Backup Database Password")
	}
	if tenantSecret.DatabaseSecret != "" {
		klog.Infof("Begin set tenant '%s' backup database password", tenant.Name)
		return sqlOperator.SetBackupPassword(tenantSecret.DatabaseSecret)
	}
	return nil
}

func (ctrl *TenantBackupCtrl) CheckAndSetBackupIncrementalPassword(tenant cloudv1.TenantSpec) error {
	tenantSecret, err := ctrl.GetTenantSecret(tenant)
	if err != nil {
		klog.Errorf("get tenant '%s' secret error '%s'", tenant.Name, err)
		return err
	}
	sqlOperator, err := ctrl.GetTenantSqlOperator(tenant)
	if err != nil {
		klog.Errorf("tenant '%s' get sql operator error when set Backup Incremental Password", tenant.Name)
		return errors.Wrap(err, "get sql operator error when set Backup Incremental Password")
	}
	if tenantSecret.IncrementalSecret != "" {
		klog.Infof("Begin set tenant '%s' backup incremental password", tenant.Name)
		return sqlOperator.SetBackupPassword(tenantSecret.IncrementalSecret)
	}
	return nil
}

func (ctrl *TenantBackupCtrl) CheckAndDoBackup(tenant cloudv1.TenantSpec) error {
	for _, schedule := range tenant.Schedule {
		// deal with full backup
		if schedule.BackupType == tenantBackupconst.FullBackup {
			// full backup once
			if schedule.Schedule == tenantBackupconst.BackupOnce {
				isBackupRunning, err := ctrl.isBackupDoing(tenant)
				if err != nil {
					klog.Errorf("Tenant '%s' backup database check whether backup is doing err '%s'", tenant.Name, err)
					return err
				}
				if !isBackupRunning {
					err = ctrl.CheckAndSetBackupDatabasePassword(tenant)
					if err != nil {
						klog.Errorf("Tenant '%s' backup database check and set password err '%s'", tenant.Name, err)
						return err
					}
					err = ctrl.StartBackupDatabase(tenant)
					if err != nil {
						klog.Errorf("Tenant '%s' start backup database err '%s'", tenant.Name, err)
						return err
					}
				}
				return ctrl.UpdateBackupStatus(tenant, "")
				//full backup, periodic
			} else {
				scheduleStatus := ctrl.getBackupScheduleStatus(tenant, tenantBackupconst.FullBackupType)
				// first time
				if scheduleStatus.NextTime == "" {
					return ctrl.UpdateBackupStatus(tenant, "")
				}
				nextTime, err := time.ParseInLocation("2006-01-02 15:04:05 +0800 CST", scheduleStatus.NextTime, time.Local)
				if err != nil {
					klog.Errorf("Tenant '%s' backup database time parse err '%s'", tenant.Name, err)
					return err
				}
				if nextTime.Before(time.Now()) || nextTime.Equal(time.Now()) {
					err = ctrl.StartBackupDatabase(tenant)
					if err != nil {
						klog.Errorf("Tenant '%s' backup database err '%s'", tenant.Name, err)
						return err
					}
					return ctrl.UpdateBackupStatus(tenant, tenantBackupconst.FullBackupType)
				}
			}

		}
		// deal with incremental backup
		if schedule.BackupType == tenantBackupconst.IncrementalBackup {
			// incremental backup once
			if schedule.Schedule == tenantBackupconst.BackupOnce {
				isBackupDoing, err := ctrl.isBackupDoing(tenant)
				if err != nil {
					klog.Errorf("Tenant '%s' backup incremental check whether backup is doing err '%s'", tenant.Name, err)
					return err
				}
				if !isBackupDoing {
					err = ctrl.CheckAndSetBackupIncrementalPassword(tenant)
					if err != nil {
						klog.Errorf("Tenant '%s' backup incremental check and set password err '%s'", tenant.Name, err)
						return err
					}
					err = ctrl.StartBackupIncremental(tenant)
					if err != nil {
						klog.Errorf("Tenant '%s' start backup incremental err '%s'", tenant.Name, err)
						return err
					}
				}
				return ctrl.UpdateBackupStatus(tenant, "")
				// incremental backup, periodic
			} else {
				scheduleStatus := ctrl.getBackupScheduleStatus(tenant, tenantBackupconst.IncrementalBackupType)
				// first time
				if scheduleStatus.NextTime == "" {
					return ctrl.UpdateBackupStatus(tenant, "")
				}
				nextTime, err := time.ParseInLocation("2006-01-02 15:04:05 +0800 CST", scheduleStatus.NextTime, time.Local)
				if err != nil {
					klog.Errorf("Tenant '%s' backup incremental time parse err '%s'", tenant.Name, err)
					return err
				}
				if nextTime.Before(time.Now()) || nextTime.Equal(time.Now()) {
					err = ctrl.StartBackupIncremental(tenant)
					if err != nil {
						klog.Errorf("Tenant '%s' backup incremental err '%s'", tenant.Name, err)
						return err
					}
					return ctrl.UpdateBackupStatus(tenant, tenantBackupconst.IncrementalBackupType)
				}
			}
		}
	}
	return ctrl.UpdateBackupStatus(tenant, "")
}

func (ctrl *TenantBackupCtrl) isBackupDoing(tenant cloudv1.TenantSpec) (bool, error) {
	sqlOperator, err := ctrl.GetTenantSqlOperator(tenant)
	if err != nil {
		klog.Errorf("tenant '%s' get sql operator error when get backup status", tenant.Name)
		return false, errors.Wrap(err, "get sql operator error when get backup status")
	}
	backupList := sqlOperator.GetAllBackupSet()
	for _, backup := range backupList {
		if backup.Status == tenantBackupconst.BackupDoing {
			return true, nil
		}
	}
	return false, nil
}

func (ctrl *TenantBackupCtrl) StartBackupDatabase(tenant cloudv1.TenantSpec) error {
	klog.Infof("Tenant '%s' begin backup database", tenant.Name)
	sqlOperator, err := ctrl.GetTenantSqlOperator(tenant)
	if err != nil {
		klog.Errorf("tenant '%s' get sql operator error when start backup database", tenant.Name)
		return errors.Wrap(err, "get sql operator error when start backup database")
	}
	return sqlOperator.StartBackupDatabase()
}

func (ctrl *TenantBackupCtrl) getBackupScheduleStatus(tenant cloudv1.TenantSpec, backupType string) cloudv1.ScheduleSpec {
	tenantBackupStatus := ctrl.GetSingleTenantBackupStatus(tenant)
	var res cloudv1.ScheduleSpec
	for _, schedule := range tenantBackupStatus.Schedule {
		if schedule.BackupType == backupType {
			res = schedule
		}
	}
	return res
}

func (ctrl *TenantBackupCtrl) GetSingleTenantBackupStatus(tenant cloudv1.TenantSpec) cloudv1.TenantBackupSetStatus {
	var res cloudv1.TenantBackupSetStatus
	tenantBackupSetList := ctrl.TenantBackup.Status.TenantBackupSet
	for _, tenantBackupSet := range tenantBackupSetList {
		if tenantBackupSet.TenantName == tenant.Name {
			res = tenantBackupSet
		}
	}
	return res
}

func (ctrl *TenantBackupCtrl) StartBackupIncremental(tenant cloudv1.TenantSpec) error {
	klog.Infof("Tenant '%s' begin backup database", tenant.Name)
	sqlOperator, err := ctrl.GetTenantSqlOperator(tenant)
	if err != nil {
		klog.Errorf("tenant '%s' get sql operator error when start backup incremental", tenant.Name)
		return errors.Wrap(err, "get sql operator error when start backup incremental")
	}
	return sqlOperator.StartBackupIncremental()
}
