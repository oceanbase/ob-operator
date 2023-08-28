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

package operation

import (
	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/sql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/pkg/errors"
)

func (m *OceanbaseOperationManager) SetLogArchiveDestForTenant(uri string) error {
	return m.SetParameter("LOG_ARCHIVE_DEST", uri, nil)
}

func (m *OceanbaseOperationManager) SetLogArchiveConcurrency(concurrency int) error {
	return m.SetParameter("log_archive_concurrency", concurrency, nil)
}

func (m *OceanbaseOperationManager) SetDataBackupDestForTenant(uri string) error {
	return m.SetParameter("DATA_BACKUP_DEST", uri, nil)
}

func (m *OceanbaseOperationManager) EnableArchiveLogForTenant() error {
	return m.ExecWithDefaultTimeout(sql.EnableArchiveLog)
}

func (m *OceanbaseOperationManager) DisableArchiveLogForTenant() error {
	return m.ExecWithDefaultTimeout(sql.DisableArchiveLog)
}

func (m *OceanbaseOperationManager) SetBackupPassword(password string) error {
	return m.ExecWithDefaultTimeout(sql.SetBackupPassword, password)
}

func (m *OceanbaseOperationManager) CreateBackupFull() error {
	return m.ExecWithDefaultTimeout(sql.CreateBackupFull)
}

func (m *OceanbaseOperationManager) CreateBackupIncr() error {
	return m.ExecWithDefaultTimeout(sql.CreateBackupIncr)
}

func (m *OceanbaseOperationManager) StopBackupJobOfTenant() error {
	return m.ExecWithDefaultTimeout(sql.StopBackupJob)
}

func (m *OceanbaseOperationManager) AddCleanBackupPolicy(policyName, recoveryWindow string) error {
	return m.ExecWithDefaultTimeout(sql.AddCleanBackupPolicy, policyName, recoveryWindow)
}

func (m *OceanbaseOperationManager) RemoveCleanBackupPolicy(policyName string) error {
	return m.ExecWithDefaultTimeout(sql.RemoveCleanBackupPolicy, policyName)
}

func (m *OceanbaseOperationManager) CancelCleanBackup() error {
	return m.ExecWithDefaultTimeout(sql.CancelCleanBackup)
}

func (m *OceanbaseOperationManager) CancelAllCleanBackup() error {
	return m.ExecWithDefaultTimeout(sql.CancelAllCleanBackup)
}

func (m *OceanbaseOperationManager) QueryArchiveLog() ([]*model.OBArchiveLogJob, error) {
	summaries := make([]*model.OBArchiveLogJob, 0)
	err := m.QueryList(&summaries, sql.QueryArchiveLog)
	if err != nil {
		m.Logger.Error(err, "Failed to query archive log job")
		return nil, errors.Wrap(err, "Query archive log job")
	}
	return summaries, nil
}

func (m *OceanbaseOperationManager) QueryArchiveLogSummary() ([]*model.OBArchiveLogSummary, error) {
	summaries := make([]*model.OBArchiveLogSummary, 0)
	err := m.QueryList(&summaries, sql.QueryArchiveLogSummary)
	if err != nil {
		m.Logger.Error(err, "Failed to query archive log summary")
		return nil, errors.Wrap(err, "Query archive log summary")
	}
	return summaries, nil
}

func (m *OceanbaseOperationManager) QueryBackupJobs() ([]*model.OBBackupJob, error) {
	return m.queryBackupJobOrHistory(sql.QueryBackupJobs)
}

func (m *OceanbaseOperationManager) QueryBackupJobHistory() ([]*model.OBBackupJob, error) {
	return m.queryBackupJobOrHistory(sql.QueryBackupHistory)
}

func (m *OceanbaseOperationManager) queryBackupJobOrHistory(statement string) ([]*model.OBBackupJob, error) {
	histories := make([]*model.OBBackupJob, 0)
	err := m.QueryList(&histories, statement)
	if err != nil {
		m.Logger.Error(err, "Failed to query backup history")
		return nil, errors.Wrap(err, "Query backup history")
	}
	return histories, nil
}

func (m *OceanbaseOperationManager) QueryBackupCleanPolicy() ([]*model.OBBackupCleanPolicy, error) {
	policies := make([]*model.OBBackupCleanPolicy, 0)
	err := m.QueryList(&policies, sql.QueryBackupCleanPolicy)
	if err != nil {
		m.Logger.Error(err, "Failed to query backup clean policy")
		return nil, errors.Wrap(err, "Query backup clean policy")
	}
	return policies, nil
}

func (m *OceanbaseOperationManager) QueryBackupCleanJobs() ([]*model.OBBackupCleanJob, error) {
	jobs := make([]*model.OBBackupCleanJob, 0)
	err := m.QueryList(&jobs, sql.QueryBackupCleanJobs)
	if err != nil {
		m.Logger.Error(err, "Failed to query backup clean job")
		return nil, errors.Wrap(err, "Query backup clean job")
	}
	return jobs, nil
}

func (m *OceanbaseOperationManager) QueryBackupCleanHistory() ([]*model.OBBackupCleanJob, error) {
	histories := make([]*model.OBBackupCleanJob, 0)
	err := m.QueryList(&histories, sql.QueryBackupCleanJobHistory)
	if err != nil {
		m.Logger.Error(err, "Failed to query backup clean history")
		return nil, errors.Wrap(err, "Query backup clean history")
	}
	return histories, nil
}

func (m *OceanbaseOperationManager) QueryArchiveLogParameters() ([]*model.OBArchiveDest, error) {
	configs := make([]*model.OBArchiveDest, 0)
	err := m.QueryList(&configs, sql.QueryArchiveLogConfigs)
	if err != nil {
		m.Logger.Error(err, "Failed to query archive log configs")
		return nil, errors.Wrap(err, "Query archive log configs")
	}
	if len(configs) == 0 {
		return nil, errors.Errorf("No archive log configs found")
	}
	return configs, nil
}

func (m *OceanbaseOperationManager) QueryBackupTasks() ([]*model.OBBackupTask, error) {
	return m.queryBackupTaskOrHistory(sql.QueryBackupTasks)
}

func (m *OceanbaseOperationManager) QueryBackupTaskHistory() ([]*model.OBBackupTask, error) {
	return m.queryBackupTaskOrHistory(sql.QueryBackupTaskHistory)
}

func (m *OceanbaseOperationManager) queryBackupTaskOrHistory(statement string) ([]*model.OBBackupTask, error) {
	tasks := make([]*model.OBBackupTask, 0)
	err := m.QueryList(&tasks, statement)
	if err != nil {
		m.Logger.Error(err, "Failed to query backup tasks")
		return nil, errors.Wrap(err, "Query backup tasks")
	}
	return tasks, nil
}
