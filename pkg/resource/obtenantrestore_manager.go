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

type ObTenantRestoreManager struct {
	ResourceManager

	Ctx      context.Context
	Resource *v1alpha1.OBTenantRestore
	Client   client.Client
	Recorder telemetry.Recorder
	Logger   *logr.Logger

	con *operation.OceanbaseOperationManager
}

func (m ObTenantRestoreManager) IsNewResource() bool {
	return m.Resource.Status.Status == ""
}

func (m *ObTenantRestoreManager) GetStatus() string {
	return string(m.Resource.Status.Status)
}

func (m ObTenantRestoreManager) IsDeleting() bool {
	return m.Resource.GetDeletionTimestamp() != nil
}

func (m ObTenantRestoreManager) CheckAndUpdateFinalizers() error {
	return nil
}

func (m ObTenantRestoreManager) InitStatus() {
	m.Resource.Status.Status = constants.RestoreJobStarting
}

func (m ObTenantRestoreManager) SetOperationContext(c *v1alpha1.OperationContext) {
	m.Resource.Status.OperationContext = c
}

func (m ObTenantRestoreManager) ClearTaskInfo() {
	m.Resource.Status.Status = constants.RestoreJobRunning
	m.Resource.Status.OperationContext = nil
}

func (m ObTenantRestoreManager) FinishTask() {
	m.Resource.Status.Status = apitypes.RestoreJobStatus(m.Resource.Status.OperationContext.TargetStatus)
	m.Resource.Status.OperationContext = nil
}

func (m ObTenantRestoreManager) HandleFailure() {
	if m.IsDeleting() {
		m.Resource.Status.OperationContext = nil
	} else {
		operationContext := m.Resource.Status.OperationContext
		failureRule := operationContext.OnFailure
		switch failureRule.Strategy {
		case "":
			fallthrough
		case strategy.StartOver:
			m.Resource.Status.Status = apitypes.RestoreJobStatus(failureRule.NextTryStatus)
			m.Resource.Status.OperationContext.Idx = 0
			m.Resource.Status.OperationContext.TaskStatus = ""
			m.Resource.Status.OperationContext.TaskId = ""
			m.Resource.Status.OperationContext.Task = ""
		case strategy.RetryFromCurrent:
			operationContext.TaskStatus = taskstatus.Pending
		case strategy.Pause:
		}
	}
}

func (m *ObTenantRestoreManager) checkRestoreProgress() error {
	con, err := m.getClusterSysClient()
	if err != nil {
		return err
	}
	restoreJob, err := con.GetLatestRestoreProgressOfTenant(m.Resource.Spec.TargetTenant)
	if err != nil {
		return err
	}
	if restoreJob != nil {
		m.Resource.Status.RestoreProgress = &model.RestoreHistory{RestoreProgress: *restoreJob}
		if restoreJob.Status == "SUCCESS" {
			m.Recorder.Event(m.Resource, corev1.EventTypeNormal, "Restore job finished", "Restore job finished")
			if m.Resource.Spec.RestoreRole == constants.TenantRoleStandby {
				m.Resource.Status.Status = constants.RestoreJobStatusReplaying
			} else {
				m.Resource.Status.Status = constants.RestoreJobStatusActivating
			}
		} else if restoreJob.Status == "FAIL" {
			m.Recorder.Event(m.Resource, corev1.EventTypeWarning, "Restore job is failed", "Restore job is failed")
			m.Resource.Status.Status = constants.RestoreJobFailed
		}
	} else {
		restoreHistory, err := con.GetLatestRestoreHistoryOfTenant(m.Resource.Spec.TargetTenant)
		if err != nil {
			return err
		}
		m.Resource.Status.RestoreProgress = restoreHistory
		if restoreHistory != nil && restoreHistory.Status == "SUCCESS" {
			m.Recorder.Event(m.Resource, corev1.EventTypeNormal, "Restore job finished", "Restore job finished")
			if m.Resource.Spec.RestoreRole == constants.TenantRoleStandby {
				if m.Resource.Spec.Source.ReplayEnabled {
					// Only if replay is enabled start log replay
					m.Resource.Status.Status = constants.RestoreJobStatusReplaying
				} else {
					m.Resource.Status.Status = constants.RestoreJobSuccessful
				}
			} else {
				m.Resource.Status.Status = constants.RestoreJobStatusActivating
			}
		} else if restoreHistory != nil && restoreHistory.Status == "FAIL" {
			m.Recorder.Event(m.Resource, corev1.EventTypeWarning, "Restore job is failed", "Restore job is failed")
			m.Resource.Status.Status = constants.RestoreJobFailed
		}
	}
	return nil
}

func (m ObTenantRestoreManager) UpdateStatus() error {
	var err error
	if m.Resource.Status.Status == constants.RestoreJobRunning {
		err = m.checkRestoreProgress()
		if err != nil {
			return err
		}
	} else if m.Resource.Status.Status == apitypes.RestoreJobStatus("Failed") {
		return nil
	}
	return m.retryUpdateStatus()
}

func (m ObTenantRestoreManager) GetTaskFunc(name string) (func() error, error) {
	switch name {
	case taskname.StartRestoreJob:
		return m.StartRestoreJobInOB, nil
	case taskname.StartLogReplay:
		return m.StartLogReplay, nil
	case taskname.ActivateStandby:
		return m.ActivateStandby, nil
	default:
		return nil, errors.New("Task name not registered")
	}
}

func (m ObTenantRestoreManager) GetTaskFlow() (*task.TaskFlow, error) {
	if m.Resource.Status.OperationContext != nil {
		return task.NewTaskFlow(m.Resource.Status.OperationContext), nil
	}
	var taskFlow *task.TaskFlow
	var err error
	status := m.Resource.Status.Status
	// get task flow depending on BackupPolicy status
	switch status {
	case constants.RestoreJobStarting:
		taskFlow, err = task.GetRegistry().Get(flow.StartRestoreFlow)
	case constants.RestoreJobStatusActivating:
		taskFlow, err = task.GetRegistry().Get(flow.RestoreAsPrimaryFlow)
	case constants.RestoreJobStatusReplaying:
		taskFlow, err = task.GetRegistry().Get(flow.RestoreAsStandbyFlow)
	case constants.RestoreJobRunning:
		fallthrough
	case constants.RestoreJobCanceled:
		fallthrough
	case constants.RestoreJobSuccessful:
		fallthrough
	case constants.RestoreJobFailed:
		fallthrough
	default:
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if taskFlow.OperationContext.OnFailure.Strategy == "" {
		taskFlow.OperationContext.OnFailure.Strategy = strategy.StartOver
		if taskFlow.OperationContext.OnFailure.NextTryStatus == "" {
			taskFlow.OperationContext.OnFailure.NextTryStatus = string(status)
		}
	}

	return taskFlow, nil
}

func (m ObTenantRestoreManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.Resource, corev1.EventTypeWarning, "task exec failed", err.Error())
}

func (m *ObTenantRestoreManager) ArchiveResource() {
	m.Logger.Info("Archive obtenant restore job", "obtenant restore job", m.Resource.Name)
	m.Recorder.Event(m.Resource, "Archive", "", "archive obtenant restore job")
	m.Resource.Status.Status = "Failed"
	m.Resource.Status.OperationContext = nil
}

func (m *ObTenantRestoreManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		resource := &v1alpha1.OBTenantRestore{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.Resource.GetNamespace(),
			Name:      m.Resource.GetName(),
		}, resource)
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		resource.Status = m.Resource.Status
		return m.Client.Status().Update(m.Ctx, resource)
	})
}
