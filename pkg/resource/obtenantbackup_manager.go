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

package resource

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	"github.com/oceanbase/ob-operator/pkg/task"
	flow "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/task/status"
	"github.com/oceanbase/ob-operator/pkg/task/strategy"
	"github.com/oceanbase/ob-operator/pkg/telemetry"
)

type OBTenantBackupManager struct {
	ResourceManager

	Ctx      context.Context
	Resource *v1alpha1.OBTenantBackup
	Client   client.Client
	Recorder telemetry.Recorder
	Logger   *logr.Logger
}

func (m *OBTenantBackupManager) IsNewResource() bool {
	return m.Resource.Status.Status == ""
}

func (m *OBTenantBackupManager) GetStatus() string {
	return string(m.Resource.Status.Status)
}

func (m *OBTenantBackupManager) IsDeleting() bool {
	return m.Resource.GetDeletionTimestamp() != nil
}

func (m *OBTenantBackupManager) CheckAndUpdateFinalizers() error {
	return nil
}

func (m *OBTenantBackupManager) InitStatus() {
	switch m.Resource.Spec.Type {
	case constants.BackupJobTypeFull, constants.BackupJobTypeIncr:
		m.Resource.Status.Status = constants.BackupJobStatusInitializing
	case constants.BackupJobTypeArchive, constants.BackupJobTypeClean:
		m.Resource.Status.Status = constants.BackupJobStatusRunning
	}
}

func (m *OBTenantBackupManager) SetOperationContext(c *v1alpha1.OperationContext) {
	m.Resource.Status.OperationContext = c
}

func (m *OBTenantBackupManager) ClearTaskInfo() {
	m.Resource.Status.Status = constants.BackupJobStatusRunning
	m.Resource.Status.OperationContext = nil
}

func (m *OBTenantBackupManager) HandleFailure() {
	if m.IsDeleting() {
		m.Resource.Status.OperationContext = nil
	} else {
		operationContext := m.Resource.Status.OperationContext
		failureRule := operationContext.OnFailure
		switch failureRule.Strategy {
		case "", strategy.StartOver:
			m.Resource.Status.Status = apitypes.BackupJobStatus(failureRule.NextTryStatus)
			operationContext.Idx = 0
			operationContext.TaskStatus = ""
			operationContext.TaskId = ""
			operationContext.Task = ""
		case strategy.RetryFromCurrent:
			operationContext.TaskStatus = taskstatus.Pending
		case strategy.Pause:
		}
	}
}

func (m *OBTenantBackupManager) FinishTask() {
	m.Resource.Status.Status = apitypes.BackupJobStatus(m.Resource.Status.OperationContext.TargetStatus)
	m.Resource.Status.OperationContext = nil
}

func (m *OBTenantBackupManager) UpdateStatus() error {
	var err error
	switch m.Resource.Spec.Type {
	case constants.BackupJobTypeFull, constants.BackupJobTypeIncr:
		if m.Resource.Status.Status == constants.BackupJobStatusRunning {
			err = m.maintainRunningBackupJob()
		}
	case constants.BackupJobTypeArchive:
		err = m.maintainRunningArchiveLogJob()
	case constants.BackupJobTypeClean:
		err = m.maintainRunningBackupCleanJob()
	}
	if err != nil {
		return err
	}
	return m.retryUpdateStatus()
}

func (m *OBTenantBackupManager) GetTaskFunc(name string) (func() error, error) {
	if name == taskname.CreateBackupJobInDB {
		return m.CreateBackupJobInOB, nil
	}

	return nil, nil
}

func (m *OBTenantBackupManager) GetTaskFlow() (*task.TaskFlow, error) {
	// exists unfinished task flow, return the last task flow
	if m.Resource.Status.OperationContext != nil {
		return task.NewTaskFlow(m.Resource.Status.OperationContext), nil
	}
	var taskFlow *task.TaskFlow
	var err error
	if m.Resource.Status.Status == constants.BackupJobStatusInitializing {
		switch m.Resource.Spec.Type {
		case constants.BackupJobTypeFull, constants.BackupJobTypeIncr:
			taskFlow, err = task.GetRegistry().Get(flow.CreateBackupJobInDB)
		}
	}
	return taskFlow, err
}

func (m *OBTenantBackupManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.Resource, corev1.EventTypeWarning, "task exec failed", err.Error())
}

func (m *OBTenantBackupManager) ArchiveResource() {
	m.Logger.Info("Archive obtenant backup job", "obtenant backup job", m.Resource.Name)
	m.Recorder.Event(m.Resource, "Archive", "", "archive obtenant backup job")
	m.Resource.Status.Status = "Failed"
	m.Resource.Status.OperationContext = nil
}

func (m *OBTenantBackupManager) CreateBackupJobInOB() error {
	job := m.Resource
	con, err := m.getObOperationClient()
	if err != nil {
		m.Logger.Error(err, "failed to get ob operation client")
		return err
	}
	if job.Spec.EncryptionSecret != "" {
		password, err := ReadPassword(m.Client, job.Namespace, job.Spec.EncryptionSecret)
		if err != nil {
			m.Logger.Error(err, "failed to read backup encryption secret")
			m.Recorder.Event(job, "Warning", "ReadBackupEncryptionSecretFailed", err.Error())
		} else if password != "" {
			err = con.SetBackupPassword(password)
			if err != nil {
				m.Logger.Error(err, "failed to set backup password")
				m.Recorder.Event(job, "Warning", "SetBackupPasswordFailed", err.Error())
			}
		}
	}
	_, err = con.CreateAndReturnBackupJob(job.Spec.Type)
	if err != nil {
		m.Logger.Error(err, "failed to create and return backup job")
		m.Recorder.Event(job, "Warning", "CreateAndReturnBackupJobFailed", err.Error())
		return err
	}

	// job.Status.BackupJob = latest
	m.Recorder.Event(job, "Create", "", "create backup job successfully")
	return nil
}

func (m *OBTenantBackupManager) getObOperationClient() (*operation.OceanbaseOperationManager, error) {
	var err error
	job := m.Resource
	obcluster := &v1alpha1.OBCluster{}
	err = m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: job.Namespace,
		Name:      job.Spec.ObClusterName,
	}, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	con, err := GetTenantRootOperationClient(m.Client, m.Logger, obcluster, job.Spec.TenantName, job.Spec.TenantSecret)
	if err != nil {
		return nil, errors.Wrap(err, "get oceanbase operation manager")
	}
	return con, nil
}

func (m *OBTenantBackupManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		newestJob := &v1alpha1.OBTenantBackup{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.Resource.GetNamespace(),
			Name:      m.Resource.GetName(),
		}, newestJob)
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		newestJob.Status = m.Resource.Status
		return m.Client.Status().Update(m.Ctx, newestJob)
	})
}

func (m *OBTenantBackupManager) maintainRunningBackupJob() error {
	logger := m.Logger
	job := m.Resource
	con, err := m.getObOperationClient()
	if err != nil {
		logger.Error(err, "failed to get ob operation client")
		return err
	}
	var targetJob *model.OBBackupJob
	if job.Status.BackupJob == nil {
		// occasionally happen, try to fetch the job from OB view
		if job.Spec.Type == constants.BackupJobTypeFull || job.Spec.Type == constants.BackupJobTypeIncr {
			latest, err := con.GetLatestBackupJobOfType(job.Spec.Type)
			if err != nil {
				return err
			}
			job.Status.BackupJob = latest
			targetJob = latest
		}
		// archive log and data clean job should not be here
	} else {
		modelJob, err := con.GetBackupJobWithId(job.Status.BackupJob.JobID)
		if err != nil {
			return err
		}
		if modelJob == nil {
			return fmt.Errorf("backup job with id %d not found", job.Status.BackupJob.JobID)
		}
		job.Status.BackupJob = modelJob
		targetJob = modelJob
	}
	job.Status.StartedAt = targetJob.StartTimestamp
	if targetJob.EndTimestamp != nil {
		job.Status.EndedAt = *targetJob.EndTimestamp
	}
	switch targetJob.Status {
	case "COMPLETED":
		job.Status.Status = constants.BackupJobStatusSuccessful
	case "FAILED":
		job.Status.Status = constants.BackupJobStatusFailed
	case "CANCELED":
		job.Status.Status = constants.BackupJobStatusCanceled
	}
	return nil
}

func (m *OBTenantBackupManager) maintainRunningBackupCleanJob() error {
	logger := m.Logger
	job := m.Resource
	con, err := m.getObOperationClient()
	if err != nil {
		logger.Error(err, "failed to get ob operation client")
		return err
	}

	latest, err := con.GetLatestBackupCleanJob()
	if err != nil {
		logger.Error(err, "failed to query latest backup clean job")
		return err
	}
	if latest != nil {
		job.Status.DataCleanJob = latest
		job.Status.StartedAt = latest.StartTimestamp
		if latest.EndTimestamp != nil {
			job.Status.EndedAt = *latest.EndTimestamp
		}
		switch latest.Status {
		case "COMPLETED":
			job.Status.Status = constants.BackupJobStatusSuccessful
		case "FAILED":
			job.Status.Status = constants.BackupJobStatusFailed
		case "CANCELED":
			job.Status.Status = constants.BackupJobStatusCanceled
		case "DOING":
			job.Status.Status = constants.BackupJobStatusRunning
		}
	}
	return nil
}

func (m *OBTenantBackupManager) maintainRunningArchiveLogJob() error {
	logger := m.Logger
	job := m.Resource
	con, err := m.getObOperationClient()
	if err != nil {
		logger.Error(err, "failed to get ob operation client")
		return err
	}

	latest, err := con.GetLatestArchiveLogJob()
	if err != nil {
		logger.Error(err, "failed to query latest archive log job")
		return err
	}
	if latest != nil {
		job.Status.ArchiveLogJob = latest
		if latest.StartScnDisplay != nil {
			job.Status.StartedAt = *latest.StartScnDisplay
		}
		job.Status.EndedAt = latest.CheckpointScnDisplay
		switch latest.Status {
		case "STOP":
			job.Status.Status = constants.BackupJobStatusStopped
		case "DOING":
			job.Status.Status = constants.BackupJobStatusRunning
		case "SUSPEND":
			job.Status.Status = constants.BackupJobStatusSuspend
		}
	}
	return nil
}
