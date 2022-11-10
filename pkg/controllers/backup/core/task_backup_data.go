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
	"time"

	"github.com/gorhill/cronexpr"
	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	backupconst "github.com/oceanbase/ob-operator/pkg/controllers/backup/const"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

func (ctrl *BackupCtrl) SetBackupDest(dest string) error {
	klog.Infoln("begin set backup destination: ", dest)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to set backup_dest = "+dest)
	}
	return sqlOperator.SetParameter("backup_dest", dest)
}

func (ctrl *BackupCtrl) setBackupDestOption() error {
	klog.Infoln("begin set backup dest option")
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to set backup dest option")
	}
	params := ctrl.Backup.Spec.Parameters
	optionList := []string{"log_archive_checkpoint_interval", "recovery_window", "auto_delete_obsolete_backup", "log_archive_piece_switch_interval"}
	for _, parameter := range params {
		for _, option := range optionList {
			if parameter.Name == option {
				err = sqlOperator.SetParameter(parameter.Name, parameter.Value)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (ctrl *BackupCtrl) setBackupLogArchiveOption() error {
	params := ctrl.Backup.Spec.Parameters
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to set backup log archieve option")
	}

	var backupMode string
	var backupCompress string
	var backupCompressAlgorithm string
	for _, parameter := range params {
		if parameter.Name == "backup_mode" {
			backupMode = parameter.Value
		}
		if parameter.Name == "backup_compress" {
			backupCompress = parameter.Value
		}
		if parameter.Name == "backup_compress_algorithm" {
			backupCompressAlgorithm = parameter.Value
		}
	}
	if backupMode == "" {
		backupMode = "optional"
	}
	if backupCompress == "" {
		backupCompress = "disable"
	}
	if backupCompressAlgorithm == "" {
		backupCompressAlgorithm = "lz4_1.0"
	}
	option := backupMode + " compression= " + backupCompress + " compression= " + backupCompressAlgorithm
	return sqlOperator.SetParameter("backup_log_archive_option", option)
}

func (ctrl *BackupCtrl) setBackupLogArchive() error {
	klog.Infoln("begin set backup log archieve ")
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to set backup log archieve")
	}
	return sqlOperator.StartArchieveLog()

}

func (ctrl *BackupCtrl) setBackupDatabasePassword() error {
	klog.Infoln("begin set backup database password ")
	params := ctrl.Backup.Spec.Parameters
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to set backup database password")
	}

	var backupDatabasePassword string
	for _, parameter := range params {
		if parameter.Name == "backup_database_password" {
			backupDatabasePassword = parameter.Value
		}
	}
	if backupDatabasePassword != "" {
		return sqlOperator.SetBackupPassword(backupDatabasePassword)
	}
	return nil
}

func (ctrl *BackupCtrl) setBackupIncrementalPassword() error {
	klog.Infoln("begin set backup incremental password ")
	params := ctrl.Backup.Spec.Parameters
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to set backup incremental password")
	}

	var backupIncrementalPassword string
	for _, parameter := range params {
		if parameter.Name == "backup_incremental_password" {
			backupIncrementalPassword = parameter.Value
		}
	}
	if backupIncrementalPassword != "" {
		return sqlOperator.SetBackupPassword(backupIncrementalPassword)
	}
	return nil
}

func (ctrl *BackupCtrl) StartBackupDatabase() error {
	klog.Infoln("begin backup database ")
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to begin backup database ")
	}
	return sqlOperator.StartBackupDatabase()
}

func (ctrl *BackupCtrl) StartBackupIncremental() error {
	klog.Infoln("begin backup database incremental")
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to begin backup database incremental")
	}
	return sqlOperator.StartBackupIncremental()
}

func (ctrl *BackupCtrl) isBackupDestSet() (error, bool) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when checking whether backup_dest is set"), false
	}
	valueList := sqlOperator.GetBackupDest()
	for _, value := range valueList {
		if value.Value != backupconst.BackupDest {
			return nil, false
		}
	}
	return nil, true
}

func (ctrl *BackupCtrl) isArchivelogDoing() (error, bool) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when checking whether archive log is doing"), false
	}
	statusList := sqlOperator.GetArchieveLogStatus()
	for _, status := range statusList {
		if status.Status != backupconst.ArchiveLogDoing {
			return nil, false
		}
	}
	return nil, true
}

func (ctrl *BackupCtrl) isBackupDoing() (error, bool) {
	klog.Infoln("Check whether backup is doing")
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when checking whether backup is doing"), false
	}
	backupList := sqlOperator.GetAllBackupSet()
	for _, backup := range backupList {
		if backup.Status == backupconst.BackupDoing {
			return nil, true
		}
	}
	return nil, false
}

func (ctrl *BackupCtrl) getBackupSchedule(backupType string) cloudv1.ScheduleSpec {
	scheduleSpec := ctrl.Backup.Spec.Schedule
	var res cloudv1.ScheduleSpec
	for _, schedule := range scheduleSpec {
		if schedule.BackupType == backupType {
			res = schedule
		}
	}
	return res
}

func (ctrl *BackupCtrl) getBackupScheduleStatus(backupType string) cloudv1.ScheduleSpec {
	scheduleSpec := ctrl.Backup.Status.Schedule
	var res cloudv1.ScheduleSpec
	for _, schedule := range scheduleSpec {
		if schedule.BackupType == backupType {
			res = schedule
		}
	}
	return res
}

func (ctrl *BackupCtrl) getNextCron(schedule string) (time.Time, error) {
	expr, err := cronexpr.Parse(schedule)
	if err != nil {
		return time.Time{}, err
	}
	nextTime := expr.Next(time.Now())
	return nextTime, nil
}
