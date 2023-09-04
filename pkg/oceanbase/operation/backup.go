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
	"github.com/oceanbase/ob-operator/api/constants"
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

func (m *OceanbaseOperationManager) CreateAndReturnBackupJob(jobType constants.BackupJobType) (*model.OBBackupJob, error) {
	var err error
	if jobType == constants.BackupJobTypeFull {
		err = m.ExecWithDefaultTimeout(sql.CreateBackupFull)
	} else if jobType == constants.BackupJobTypeIncr {
		err = m.ExecWithDefaultTimeout(sql.CreateBackupIncr)
	} else {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m.GetLatestBackupJobOfType(jobType)
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

func (m *OceanbaseOperationManager) ListArchiveLog() ([]*model.OBArchiveLogJob, error) {
	summaries := make([]*model.OBArchiveLogJob, 0)
	err := m.QueryList(&summaries, sql.QueryArchiveLog)
	if err != nil {
		m.Logger.Error(err, "Failed to query archive log job")
		return nil, errors.Wrap(err, "Query archive log job")
	}
	return summaries, nil
}

func (m *OceanbaseOperationManager) ListArchiveLogSummary() ([]*model.OBArchiveLogSummary, error) {
	summaries := make([]*model.OBArchiveLogSummary, 0)
	err := m.QueryList(&summaries, sql.QueryArchiveLogSummary)
	if err != nil {
		m.Logger.Error(err, "Failed to query archive log summary")
		return nil, errors.Wrap(err, "Query archive log summary")
	}
	return summaries, nil
}

func (m *OceanbaseOperationManager) ListBackupJobs() ([]*model.OBBackupJob, error) {
	return m.listBackupJobOrHistory(sql.QueryBackupJobs)
}

func (m *OceanbaseOperationManager) ListBackupJobHistory() ([]*model.OBBackupJob, error) {
	return m.listBackupJobOrHistory(sql.QueryBackupHistory)
}

func (m *OceanbaseOperationManager) listBackupJobOrHistory(statement string) ([]*model.OBBackupJob, error) {
	histories := make([]*model.OBBackupJob, 0)
	err := m.QueryList(&histories, statement)
	if err != nil {
		m.Logger.Error(err, "Failed to query backup history")
		return nil, errors.Wrap(err, "Query backup history")
	}
	return histories, nil
}

func (m *OceanbaseOperationManager) ListBackupCleanPolicy() ([]*model.OBBackupCleanPolicy, error) {
	policies := make([]*model.OBBackupCleanPolicy, 0)
	err := m.QueryList(&policies, sql.QueryBackupCleanPolicy)
	if err != nil {
		m.Logger.Error(err, "Failed to query backup clean policy")
		return nil, errors.Wrap(err, "Query backup clean policy")
	}
	return policies, nil
}

func (m *OceanbaseOperationManager) ListBackupCleanJobs() ([]*model.OBBackupCleanJob, error) {
	jobs := make([]*model.OBBackupCleanJob, 0)
	err := m.QueryList(&jobs, sql.QueryBackupCleanJobs)
	if err != nil {
		m.Logger.Error(err, "Failed to query backup clean job")
		return nil, errors.Wrap(err, "Query backup clean job")
	}
	return jobs, nil
}

func (m *OceanbaseOperationManager) ListBackupCleanHistory() ([]*model.OBBackupCleanJob, error) {
	histories := make([]*model.OBBackupCleanJob, 0)
	err := m.QueryList(&histories, sql.QueryBackupCleanJobHistory)
	if err != nil {
		m.Logger.Error(err, "Failed to query backup clean history")
		return nil, errors.Wrap(err, "Query backup clean history")
	}
	return histories, nil
}

func (m *OceanbaseOperationManager) ListArchiveLogParameters() ([]*model.OBArchiveDest, error) {
	configs := make([]*model.OBArchiveDest, 0)
	err := m.QueryList(&configs, sql.QueryArchiveLogDestConfigs)
	if err != nil {
		m.Logger.Error(err, "Failed to query archive log configs")
		return nil, errors.Wrap(err, "Query archive log configs")
	}
	if len(configs) == 0 {
		return nil, errors.Errorf("No archive log configs found")
	}
	return configs, nil
}

func (m *OceanbaseOperationManager) ListBackupParameters() ([]*model.OBBackupParameter, error) {
	configs := make([]*model.OBBackupParameter, 0)
	err := m.QueryList(&configs, sql.QueryBackupParameter)
	if err != nil {
		m.Logger.Error(err, "Failed to query archive log configs")
		return nil, errors.Wrap(err, "Query archive log configs")
	}
	if len(configs) == 0 {
		return nil, errors.Errorf("No archive log configs found")
	}
	return configs, nil
}

func (m *OceanbaseOperationManager) ListBackupTasks() ([]*model.OBBackupTask, error) {
	return m.listBackupTaskOrHistory(sql.QueryBackupTasks)
}

func (m *OceanbaseOperationManager) ListBackupTaskHistory() ([]*model.OBBackupTask, error) {
	return m.listBackupTaskOrHistory(sql.QueryBackupTaskHistory)
}

func (m *OceanbaseOperationManager) listBackupTaskOrHistory(statement string) ([]*model.OBBackupTask, error) {
	tasks := make([]*model.OBBackupTask, 0)
	err := m.QueryList(&tasks, statement)
	if err != nil {
		m.Logger.Error(err, "Failed to query backup tasks")
		return nil, errors.Wrap(err, "Query backup tasks")
	}
	return tasks, nil
}

func (m *OceanbaseOperationManager) GetLatestBackupJobOfType(jobType constants.BackupJobType) (*model.OBBackupJob, error) {
	return m.getLatestBackupJob([]string{sql.QueryLatestBackupJobOfType, sql.QueryLatestBackupJobHistoryOfType}, jobType)
}

func (m *OceanbaseOperationManager) GetLatestBackupJobOfTypeAndPath(jobType constants.BackupJobType, path string) (*model.OBBackupJob, error) {
	return m.getLatestBackupJob([]string{sql.QueryLatestBackupJobOfTypeAndPath, sql.QueryLatestBackupJobHistoryOfTypeAndPath}, jobType, path)
}

func (m *OceanbaseOperationManager) getLatestBackupJob(statements []string, params ...interface{}) (*model.OBBackupJob, error) {
	if len(statements) != 2 {
		return nil, errors.New("unexpected # of statements, require exactly 2 statement")
	}
	jobs := make([]*model.OBBackupJob, 0)
	err := m.QueryList(&jobs, statements[0], params...)
	if err != nil {
		m.Logger.Error(err, "Failed to query latest running backup job")
		return nil, errors.Wrap(err, "Query latest running backup job")
	}
	if len(jobs) == 0 {
		err = m.QueryList(&jobs, statements[1], params...)
		if err != nil {
			m.Logger.Error(err, "Failed to query latest backup job history")
			return nil, errors.Wrap(err, "Query latest backup job history")
		}
		if len(jobs) == 0 {
			return nil, nil
		}
	}
	return jobs[0], nil
}

func (m *OceanbaseOperationManager) GetBackupJobWithId(jobId int64) (*model.OBBackupJob, error) {
	jobs := make([]*model.OBBackupJob, 0)
	err := m.QueryList(&jobs, sql.QueryBackupJobWithId, jobId)
	if err != nil {
		return nil, err
	}
	if len(jobs) != 0 {
		return jobs[0], nil
	}
	err = m.QueryList(&jobs, sql.QueryBackupHistoryWithId, jobId)
	if err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, nil
	}
	return jobs[0], nil
}

func (m *OceanbaseOperationManager) ListBackupTaskWithJobId(jobId int64) ([]*model.OBBackupTask, error) {
	tasks := make([]*model.OBBackupTask, 0)
	taskHistory := make([]*model.OBBackupTask, 0)
	err := m.QueryList(&tasks, sql.QueryBackupTaskWithJobId, jobId)
	if err != nil {
		return nil, err
	}
	err = m.QueryList(&taskHistory, sql.QueryBackupTaskHistoryWithJobId, jobId)
	if err != nil {
		return nil, err
	}
	tasks = append(tasks, taskHistory...)
	return tasks, nil
}

func (m *OceanbaseOperationManager) GetLatestBackupCleanJob() (*model.OBBackupCleanJob, error) {
	jobs := make([]*model.OBBackupCleanJob, 0)
	err := m.QueryList(&jobs, sql.QueryLatestCleanJob)
	if err != nil {
		return nil, err
	}
	if len(jobs) != 0 {
		return jobs[0], nil
	}
	return nil, nil
}

func (m *OceanbaseOperationManager) GetLatestArchiveLogJob() (*model.OBArchiveLogJob, error) {
	jobs := make([]*model.OBArchiveLogJob, 0)
	err := m.QueryList(&jobs, sql.QueryLatestArchiveLogJob)
	if err != nil {
		return nil, err
	}
	if len(jobs) != 0 {
		return jobs[0], nil
	}
	return nil, nil
}

func (m *OceanbaseOperationManager) GetLatestRunningBackupJob() (*model.OBBackupJob, error) {
	jobs := make([]*model.OBBackupJob, 0, 1)
	err := m.QueryList(&jobs, sql.QueryLatestRunningBackupJob)
	if err != nil {
		return nil, err
	}
	if len(jobs) != 0 {
		return jobs[0], nil
	}
	return nil, nil
}
