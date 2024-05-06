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
	"context"

	"github.com/pkg/errors"

	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/sql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

func (m *OceanbaseOperationManager) SetLogArchiveDestForTenant(ctx context.Context, uri string) error {
	return m.SetParameter(ctx, "LOG_ARCHIVE_DEST", uri, nil)
}

func (m *OceanbaseOperationManager) SetLogArchiveDestState(ctx context.Context, state string) error {
	return m.SetParameter(ctx, "LOG_ARCHIVE_DEST_STATE", state, nil)
}

func (m *OceanbaseOperationManager) SetLogArchiveConcurrency(ctx context.Context, concurrency int) error {
	return m.SetParameter(ctx, "LOG_ARCHIVE_CONCURRENCY", concurrency, nil)
}

func (m *OceanbaseOperationManager) SetDataBackupDestForTenant(ctx context.Context, uri string) error {
	return m.SetParameter(ctx, "DATA_BACKUP_DEST", uri, nil)
}

func (m *OceanbaseOperationManager) EnableArchiveLogForTenant(ctx context.Context) error {
	return m.ExecWithDefaultTimeout(ctx, sql.EnableArchiveLog)
}

func (m *OceanbaseOperationManager) DisableArchiveLogForTenant(ctx context.Context) error {
	return m.ExecWithDefaultTimeout(ctx, sql.DisableArchiveLog)
}

func (m *OceanbaseOperationManager) SetBackupPassword(ctx context.Context, password string) error {
	return m.ExecWithDefaultTimeout(ctx, sql.SetBackupPassword, password)
}

func (m *OceanbaseOperationManager) CreateBackupFull(ctx context.Context) error {
	return m.ExecWithDefaultTimeout(ctx, sql.CreateBackupFull)
}

func (m *OceanbaseOperationManager) CreateBackupIncr(ctx context.Context) error {
	return m.ExecWithDefaultTimeout(ctx, sql.CreateBackupIncr)
}

func (m *OceanbaseOperationManager) CreateAndReturnBackupJob(ctx context.Context, jobType string) (*model.OBBackupJob, error) {
	var err error
	if jobType == "FULL" {
		err = m.ExecWithDefaultTimeout(ctx, sql.CreateBackupFull)
	} else if jobType == "INC" {
		err = m.ExecWithDefaultTimeout(ctx, sql.CreateBackupIncr)
	} else {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m.GetLatestBackupJobOfType(ctx, jobType)
}

func (m *OceanbaseOperationManager) StopBackupJobOfTenant(ctx context.Context) error {
	return m.ExecWithDefaultTimeout(ctx, sql.StopBackupJob)
}

func (m *OceanbaseOperationManager) AddCleanBackupPolicy(ctx context.Context, policyName, recoveryWindow string) error {
	return m.ExecWithDefaultTimeout(ctx, sql.AddCleanBackupPolicy, policyName, recoveryWindow)
}

func (m *OceanbaseOperationManager) RemoveCleanBackupPolicy(ctx context.Context, policyName string) error {
	return m.ExecWithDefaultTimeout(ctx, sql.RemoveCleanBackupPolicy, policyName)
}

func (m *OceanbaseOperationManager) CancelCleanBackup(ctx context.Context) error {
	return m.ExecWithDefaultTimeout(ctx, sql.CancelCleanBackup)
}

func (m *OceanbaseOperationManager) CancelAllCleanBackup(ctx context.Context) error {
	return m.ExecWithDefaultTimeout(ctx, sql.CancelAllCleanBackup)
}

func (m *OceanbaseOperationManager) ListArchiveLog(ctx context.Context) ([]*model.OBArchiveLogJob, error) {
	summaries := make([]*model.OBArchiveLogJob, 0)
	err := m.QueryList(ctx, &summaries, sql.QueryArchiveLog)
	if err != nil {
		m.Logger.Error(err, "Failed to query archive log job")
		return nil, errors.Wrap(err, "Query archive log job")
	}
	return summaries, nil
}

func (m *OceanbaseOperationManager) ListArchiveLogSummary(ctx context.Context) ([]*model.OBArchiveLogSummary, error) {
	summaries := make([]*model.OBArchiveLogSummary, 0)
	err := m.QueryList(ctx, &summaries, sql.QueryArchiveLogSummary)
	if err != nil {
		m.Logger.Error(err, "Failed to query archive log summary")
		return nil, errors.Wrap(err, "Query archive log summary")
	}
	return summaries, nil
}

func (m *OceanbaseOperationManager) ListBackupJobs(ctx context.Context) ([]*model.OBBackupJob, error) {
	return m.listBackupJobOrHistory(ctx, sql.QueryBackupJobs)
}

func (m *OceanbaseOperationManager) ListBackupJobHistory(ctx context.Context) ([]*model.OBBackupJob, error) {
	return m.listBackupJobOrHistory(ctx, sql.QueryBackupHistory)
}

func (m *OceanbaseOperationManager) listBackupJobOrHistory(ctx context.Context, statement string) ([]*model.OBBackupJob, error) {
	histories := make([]*model.OBBackupJob, 0)
	err := m.QueryList(ctx, &histories, statement)
	if err != nil {
		m.Logger.Error(err, "Failed to query backup history")
		return nil, errors.Wrap(err, "Query backup history")
	}
	return histories, nil
}

func (m *OceanbaseOperationManager) ListBackupCleanPolicy(ctx context.Context) ([]*model.OBBackupCleanPolicy, error) {
	policies := make([]*model.OBBackupCleanPolicy, 0)
	err := m.QueryList(ctx, &policies, sql.QueryBackupCleanPolicy)
	if err != nil {
		m.Logger.Error(err, "Failed to query backup clean policy")
		return nil, errors.Wrap(err, "Query backup clean policy")
	}
	return policies, nil
}

func (m *OceanbaseOperationManager) ListBackupCleanJobs(ctx context.Context) ([]*model.OBBackupCleanJob, error) {
	jobs := make([]*model.OBBackupCleanJob, 0)
	err := m.QueryList(ctx, &jobs, sql.QueryBackupCleanJobs)
	if err != nil {
		m.Logger.Error(err, "Failed to query backup clean job")
		return nil, errors.Wrap(err, "Query backup clean job")
	}
	return jobs, nil
}

func (m *OceanbaseOperationManager) ListBackupCleanHistory(ctx context.Context) ([]*model.OBBackupCleanJob, error) {
	histories := make([]*model.OBBackupCleanJob, 0)
	err := m.QueryList(ctx, &histories, sql.QueryBackupCleanJobHistory)
	if err != nil {
		m.Logger.Error(err, "Failed to query backup clean history")
		return nil, errors.Wrap(err, "Query backup clean history")
	}
	return histories, nil
}

func (m *OceanbaseOperationManager) ListArchiveLogParameters(ctx context.Context) ([]*model.OBArchiveDest, error) {
	configs := make([]*model.OBArchiveDest, 0)
	err := m.QueryList(ctx, &configs, sql.QueryArchiveLogDestConfigs)
	if err != nil {
		m.Logger.Error(err, "Failed to query archive log configs")
		return nil, errors.Wrap(err, "Query archive log configs")
	}
	return configs, nil
}

func (m *OceanbaseOperationManager) ListBackupParameters(ctx context.Context) ([]*model.OBBackupParameter, error) {
	configs := make([]*model.OBBackupParameter, 0)
	err := m.QueryList(ctx, &configs, sql.QueryBackupParameter)
	if err != nil {
		m.Logger.Error(err, "Failed to query archive log configs")
		return nil, errors.Wrap(err, "Query archive log configs")
	}
	return configs, nil
}

func (m *OceanbaseOperationManager) ListBackupTasks(ctx context.Context) ([]*model.OBBackupTask, error) {
	return m.listBackupTaskOrHistory(ctx, sql.QueryBackupTasks)
}

func (m *OceanbaseOperationManager) ListBackupTaskHistory(ctx context.Context) ([]*model.OBBackupTask, error) {
	return m.listBackupTaskOrHistory(ctx, sql.QueryBackupTaskHistory)
}

func (m *OceanbaseOperationManager) listBackupTaskOrHistory(ctx context.Context, statement string) ([]*model.OBBackupTask, error) {
	tasks := make([]*model.OBBackupTask, 0)
	err := m.QueryList(ctx, &tasks, statement)
	if err != nil {
		m.Logger.Error(err, "Failed to query backup tasks")
		return nil, errors.Wrap(err, "Query backup tasks")
	}
	return tasks, nil
}

func (m *OceanbaseOperationManager) GetLatestBackupJobOfType(ctx context.Context, jobType string) (*model.OBBackupJob, error) {
	return m.getLatestBackupJob(ctx, []string{sql.QueryLatestBackupJobOfType, sql.QueryLatestBackupJobHistoryOfType}, jobType)
}

func (m *OceanbaseOperationManager) GetLatestBackupJobOfTypeAndPath(ctx context.Context, jobType string, path string) (*model.OBBackupJob, error) {
	return m.getLatestBackupJob(ctx, []string{sql.QueryLatestBackupJobOfTypeAndPath, sql.QueryLatestBackupJobHistoryOfTypeAndPath}, jobType, path)
}

func (m *OceanbaseOperationManager) getLatestBackupJob(ctx context.Context, statements []string, params ...any) (*model.OBBackupJob, error) {
	if len(statements) != 2 {
		return nil, errors.New("unexpected # of statements, require exactly 2 statement")
	}
	jobs := make([]*model.OBBackupJob, 0)
	err := m.QueryList(ctx, &jobs, statements[0], params...)
	if err != nil {
		m.Logger.Error(err, "Failed to query latest running backup job")
		return nil, errors.Wrap(err, "Query latest running backup job")
	}
	if len(jobs) == 0 {
		err = m.QueryList(ctx, &jobs, statements[1], params...)
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

func (m *OceanbaseOperationManager) GetBackupJobWithId(ctx context.Context, jobId int64) (*model.OBBackupJob, error) {
	jobs := make([]*model.OBBackupJob, 0)
	err := m.QueryList(ctx, &jobs, sql.QueryBackupJobWithId, jobId)
	if err != nil {
		return nil, err
	}
	if len(jobs) != 0 {
		return jobs[0], nil
	}
	err = m.QueryList(ctx, &jobs, sql.QueryBackupHistoryWithId, jobId)
	if err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, nil
	}
	return jobs[0], nil
}

func (m *OceanbaseOperationManager) ListBackupTaskWithJobId(ctx context.Context, jobId int64) ([]*model.OBBackupTask, error) {
	tasks := make([]*model.OBBackupTask, 0)
	taskHistory := make([]*model.OBBackupTask, 0)
	err := m.QueryList(ctx, &tasks, sql.QueryBackupTaskWithJobId, jobId)
	if err != nil {
		return nil, err
	}
	err = m.QueryList(ctx, &taskHistory, sql.QueryBackupTaskHistoryWithJobId, jobId)
	if err != nil {
		return nil, err
	}
	tasks = append(tasks, taskHistory...)
	return tasks, nil
}

func (m *OceanbaseOperationManager) GetLatestBackupCleanJob(ctx context.Context) (*model.OBBackupCleanJob, error) {
	jobs := make([]*model.OBBackupCleanJob, 0)
	err := m.QueryList(ctx, &jobs, sql.QueryLatestCleanJob)
	if err != nil {
		return nil, err
	}
	if len(jobs) != 0 {
		return jobs[0], nil
	}
	err = m.QueryList(ctx, &jobs, sql.QueryLatestCleanJobHistory)
	if err != nil {
		return nil, err
	}
	if len(jobs) != 0 {
		return jobs[0], nil
	}
	return nil, nil
}

func (m *OceanbaseOperationManager) GetLatestArchiveLogJob(ctx context.Context) (*model.OBArchiveLogJob, error) {
	jobs := make([]*model.OBArchiveLogJob, 0)
	err := m.QueryList(ctx, &jobs, sql.QueryLatestArchiveLogJob)
	if err != nil {
		return nil, err
	}
	if len(jobs) != 0 {
		return jobs[0], nil
	}
	return nil, nil
}

func (m *OceanbaseOperationManager) GetLatestRunningBackupJob(ctx context.Context) (*model.OBBackupJob, error) {
	jobs := make([]*model.OBBackupJob, 0, 1)
	err := m.QueryList(ctx, &jobs, sql.QueryLatestRunningBackupJob)
	if err != nil {
		return nil, err
	}
	if len(jobs) != 0 {
		return jobs[0], nil
	}
	return nil, nil
}
