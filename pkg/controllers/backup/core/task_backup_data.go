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
	"time"

	"github.com/gorhill/cronexpr"
	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	backupconst "github.com/oceanbase/ob-operator/pkg/controllers/backup/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/backup/model"
	"github.com/oceanbase/ob-operator/pkg/controllers/backup/sql"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

func (ctrl *BackupCtrl) GetSecret(name string) (model.Secret, error) {
	var secret model.Secret
	obcluster := ctrl.Backup.Spec.SourceCluster
	secretExecutor := resource.NewSecretResource(ctrl.Resource)
	backupSecret, err := secretExecutor.Get(context.TODO(), obcluster.ClusterNamespace, name)
	if err != nil {
		klog.Errorf("get backup secret error '%s'", err)
		return secret, err
	}
	secret.IncrementalSecret = strings.TrimRight(string(backupSecret.(corev1.Secret).Data[backupconst.IncrementalSecret]), "\n")
	secret.FullSecret = strings.TrimRight(string(backupSecret.(corev1.Secret).Data[backupconst.FullSecret]), "\n")
	return secret, nil
}

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

func (ctrl *BackupCtrl) GetBackupDatabasePasswordSQL() (string, error) {
	secretName := ctrl.Backup.Spec.Secret
	secret, err := ctrl.GetSecret(secretName)
	if err != nil {
		klog.Errorf("get secret '%s' error '%s'", secretName, err)
		return "", err
	}
	if secret.FullSecret != "" {
		klog.Infoln("need set backup database password ")
		return sql.ReplaceAll(sql.SetBackupPasswordTemplate, sql.SetBackupPasswordReplacer(secret.FullSecret)), nil
	}
	return "", nil
}

func (ctrl *BackupCtrl) GetBackupIncrementalPasswordSQL() (string, error) {
	secretName := ctrl.Backup.Spec.Secret
	secret, err := ctrl.GetSecret(secretName)
	if err != nil {
		klog.Errorf("get secret '%s' error '%s'", secretName, err)
		return "", err
	}
	if secret.IncrementalSecret != "" {
		klog.Infoln("begin set backup incremental password ")
		return sql.ReplaceAll(sql.SetBackupPasswordTemplate, sql.SetBackupPasswordReplacer(secret.IncrementalSecret)), nil
	}
	return "", nil
}

func (ctrl *BackupCtrl) StartBackupDatabase() error {
	klog.Infoln("begin backup database ")
	SQLs := make([]string, 0)
	pwdSQL, err := ctrl.GetBackupDatabasePasswordSQL()
	if err != nil {
		klog.Errorln("DoBackup: get backup database password err ", err)
		return err
	}
	if pwdSQL != "" {
		SQLs = append(SQLs, pwdSQL)
	}
	SQLs = append(SQLs, sql.StartBackupDatabaseSql)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to begin backup database ")
	}
	return sqlOperator.ExecSQLs(SQLs)
}

func (ctrl *BackupCtrl) StartBackupIncremental() error {
	klog.Infoln("begin backup database incremental")
	SQLs := make([]string, 0)
	pwdSQL, err := ctrl.GetBackupIncrementalPasswordSQL()
	if err != nil {
		klog.Errorln("DoBackup: get backup incremental password err ", err)
		return err
	}
	if pwdSQL != "" {
		SQLs = append(SQLs, pwdSQL)
	}
	SQLs = append(SQLs, sql.StartBackupIncrementalSql)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to begin backup database incremental")
	}
	return sqlOperator.ExecSQLs(SQLs)
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

func (ctrl *BackupCtrl) GetArchivelogStatus() (string, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return "", errors.Wrap(err, "get sql operator when checking whether archive log is doing")
	}
	statusList := sqlOperator.GetArchiveLogStatus()
	for _, status := range statusList {
		if status.Status != backupconst.ArchiveLogDoing {
			return status.Status, nil
		}
	}
	return backupconst.ArchiveLogDoing, nil
}

func (ctrl *BackupCtrl) IsArchivelogStop() (bool, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return false, errors.Wrap(err, "get sql operator when checking whether archive log is doing")
	}
	statusList := sqlOperator.GetArchiveLogStatus()
	for _, status := range statusList {
		if status.Status != backupconst.ArchiveLogStopping && status.Status != backupconst.ArchiveLogStop {
			return false, nil
		}
	}
	return true, nil
}

func (ctrl *BackupCtrl) isBackupDoing() (error, bool) {
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

func (ctrl *BackupCtrl) WaitArchivelogDoing() error {
	klog.Infoln("Wait Archivelog Doing")
	err := ctrl.TickerCheckArchivelogDoing()
	if err != nil {
		return err
	}
	klog.Infoln("Archivelog Doing")
	return nil
}

func (ctrl *BackupCtrl) TickerCheckArchivelogDoing() error {
	tick := time.Tick(backupconst.TickerPeriodLogArchiveCheck)
	var num int
	for {
		select {
		case <-tick:
			if num > backupconst.TickNumForLogArchiveCheck {
				return errors.New("wait for logarchive doing timeout")
			}
			num = num + 1
			status, err := ctrl.GetArchivelogStatus()
			if status == backupconst.ArchiveLogDoing {
				return err
			}
		}
	}
}

func (ctrl *BackupCtrl) CancelBackupTasks() error {
	klog.Infoln("begin cancel backup tasks")
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to cancel backup tasks")
	}
	return sqlOperator.CancelBackupTasks()
}

func (ctrl *BackupCtrl) CancelArchiveLog() error {
	klog.Infoln("begin cancel backup log archieve ")
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to cancel backup log archieve")
	}
	return sqlOperator.StopArchiveLog()
}
