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
	"strings"
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
	return sqlOperator.SetParameter(backupconst.DestPathName, dest)
}

func (ctrl *BackupCtrl) setBackupDestOption() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to set backup dest option")
	}
	var logArchiveCheckpointInterval = cloudv1.Parameter{Name: backupconst.LogArchiveCheckpointIntervalName, Value: backupconst.LogArchiveCheckpointIntervalDefault}
	var recoveryWindow = cloudv1.Parameter{Name: backupconst.RecoveryWindowName, Value: backupconst.RecoveryWindowDefault}
	var autoDeleteObsoleteBackup = cloudv1.Parameter{Name: backupconst.AutoDeleteObsoleteBackupName, Value: backupconst.AutoDeleteObsoleteBackupDefault}
	var logArchivePieceSwitchInterval = cloudv1.Parameter{Name: backupconst.LogArchivePieceSwitchIntervalName, Value: backupconst.LogArchivePieceSwitchIntervalDefault}
	paramList := [4]cloudv1.Parameter{logArchiveCheckpointInterval, recoveryWindow, autoDeleteObsoleteBackup, logArchivePieceSwitchInterval}
	params := ctrl.Backup.Spec.Parameters
	var isSet bool
	var option string
	for _, p := range paramList {
		for _, param := range params {
			if param.Name == p.Name {
				isSet = true
				break
			}
		}
	}
	if !isSet {
		return nil
	}
	for _, p := range paramList {
		if ctrl.getParameter(p.Name) != "" {
			p.Value = ctrl.getParameter(p.Name)
		}
		option += p.Name + "=" + p.Value + "&"
	}
	option = strings.TrimRight(option, "&")
	return sqlOperator.SetParameter(backupconst.BackupDestOptionName, option)
}

func (ctrl *BackupCtrl) setBackupLogArchiveOption() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to set backup log archieve option")
	}

	var backupMode = cloudv1.Parameter{Name: backupconst.BackupModeName, Value: backupconst.BackupModeDefault}
	var backupCompress = cloudv1.Parameter{Name: backupconst.BackupCompressName, Value: backupconst.BackupCompressDefault}
	var backupCompressAlgorithm = cloudv1.Parameter{Name: backupconst.BackupCompressAlgorithmName, Value: backupconst.BackupCompressAlgorithmDefault}
	paramList := [3]cloudv1.Parameter{backupMode, backupCompress, backupCompressAlgorithm}
	params := ctrl.Backup.Spec.Parameters
	var isSet bool
	var option string
	for _, p := range paramList {
		for _, param := range params {
			if param.Name == p.Name {
				isSet = true
				break
			}
		}
	}
	if !isSet {
		return nil
	}
	for _, p := range paramList {
		if ctrl.getParameter(p.Name) != "" {
			p.Value = ctrl.getParameter(p.Name)
		}
		option += p.Name + "=" + p.Value + " "
	}
	return sqlOperator.SetParameter(backupconst.BackupLogArchiveOptionName, option)
}

func (ctrl *BackupCtrl) getParameter(name string) string {
	params := ctrl.Backup.Spec.Parameters
	for _, parameter := range params {
		if parameter.Name == name {
			return parameter.Value
		}
	}
	return ""
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
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to set backup database password")
	}
	backupDatabasePassword := ctrl.getParameter(backupconst.BackupDatabasePasswordName)

	if backupDatabasePassword != "" {
		klog.Infoln("begin set backup database password ")
		return sqlOperator.SetBackupPassword(backupDatabasePassword)
	} else {
		return nil
	}
}

func (ctrl *BackupCtrl) setBackupIncrementalPassword() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to set backup incremental password")
	}
	backupIncrementalPassword := ctrl.getParameter(backupconst.BackupIncrementalPasswordName)
	if backupIncrementalPassword != "" {
		klog.Infoln("begin set backup incremental password ")
		return sqlOperator.SetBackupPassword(backupIncrementalPassword)
	} else {
		return nil
	}
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
		return errors.Wrap(err, "get sql operator when checking whether backup_dest is set or changed"), false
	}
	valueList := sqlOperator.GetBackupDest()
	for _, value := range valueList {
		if value.Value == "" || value.Value != ctrl.Backup.Spec.DestPath {
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
		if status.Status != backupconst.ArchiveLogDoing && status.Status != backupconst.ArchiveLogBeginning {
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
