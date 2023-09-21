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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/constants"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	"github.com/oceanbase/ob-operator/pkg/task"
	flow "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
	"github.com/oceanbase/ob-operator/pkg/task/strategy"
)

type ObTenantRestoreManager struct {
	ResourceManager

	Ctx      context.Context
	Resource *v1alpha1.OBTenantRestore
	Client   client.Client
	Recorder record.EventRecorder
	Logger   *logr.Logger

	con *operation.OceanbaseOperationManager
}

func (m ObTenantRestoreManager) IsNewResource() bool {
	return m.Resource.Status.Status == ""
}

func (m ObTenantRestoreManager) IsDeleting() bool {
	return m.Resource.GetDeletionTimestamp() != nil
}

func (m ObTenantRestoreManager) CheckAndUpdateFinalizers() error {
	return nil
}

func (m ObTenantRestoreManager) InitStatus() {
	m.Resource.Status.Status = constants.RestoreJobRunning
}

func (m ObTenantRestoreManager) SetOperationContext(c *v1alpha1.OperationContext) {
	m.Resource.Status.OperationContext = c
}

func (m ObTenantRestoreManager) ClearTaskInfo() {
	m.Resource.Status.Status = constants.RestoreJobRunning
	m.Resource.Status.OperationContext = nil
}

func (m ObTenantRestoreManager) FinishTask() {
	m.Resource.Status.Status = constants.RestoreJobStatus(m.Resource.Status.OperationContext.TargetStatus)
	m.Resource.Status.OperationContext = nil
}

func (m ObTenantRestoreManager) HandleFailure() {

}

func (m ObTenantRestoreManager) UpdateStatus() error {
	return m.Client.Status().Update(m.Ctx, m.Resource)
}

func (m ObTenantRestoreManager) GetTaskFunc(name string) (func() error, error) {
	switch name {
	case taskname.StartRestoreJob:
		return m.StartRestoreJobInOB, nil
	case taskname.StartLogReplay:
		return m.StartLogReplay, nil
	case taskname.CancelRestoreJob:
		return m.CancelRestoreJob, nil
	case taskname.ActivateStandby:
		return m.ActivateStandby, nil
	case taskname.CheckRestoreProgress:
		return m.CheckRestoreProgress, nil
	}
	return nil, nil
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
		taskFlow, err = task.GetRegistry().Get(flow.PrepareBackupPolicy)
	case constants.RestoreJobRunning:
		taskFlow, err = task.GetRegistry().Get(flow.PrepareBackupPolicy)
	case constants.RestoreJobCanceling:
		taskFlow, err = task.GetRegistry().Get(flow.PrepareBackupPolicy)
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
