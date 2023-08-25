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
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	"github.com/oceanbase/ob-operator/pkg/task"
	flow "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObTenantBackupPolicyManager struct {
	ResourceManager
	Ctx          context.Context
	BackupPolicy *v1alpha1.OBTenantBackupPolicy
	Client       client.Client
	Recorder     record.EventRecorder
	Logger       *logr.Logger

	con *operation.OceanbaseOperationManager
}

func (m *ObTenantBackupPolicyManager) IsNewResource() bool {
	return m.BackupPolicy.Status.Status == ""
}

func (m *ObTenantBackupPolicyManager) IsDeleting() bool {
	return !m.BackupPolicy.ObjectMeta.DeletionTimestamp.IsZero()
}

func (m *ObTenantBackupPolicyManager) CheckAndUpdateFinalizers() error {
	// Unnecessary by now
	return nil
}

func (m *ObTenantBackupPolicyManager) InitStatus() {
	m.Logger.Info("Initialize status for new BackupPolicy")
	m.BackupPolicy.Status = v1alpha1.OBTenantBackupPolicyStatus{
		Status:                 v1alpha1.BackupPolicyStatusPreparing,
		LogArchiveDestDisabled: false,
	}
}

func (m *ObTenantBackupPolicyManager) SetOperationContext(c *v1alpha1.OperationContext) {
	m.BackupPolicy.Status.OperationContext = c
}

func (m *ObTenantBackupPolicyManager) ClearTaskInfo() {
	m.BackupPolicy.Status.Status = v1alpha1.BackupPolicyStatusRunning
	m.BackupPolicy.Status.OperationContext = nil
}

func (m *ObTenantBackupPolicyManager) FinishTask() {
	m.BackupPolicy.Status.Status = v1alpha1.BackupPolicyStatusType(m.BackupPolicy.Status.OperationContext.TargetStatus)
	m.BackupPolicy.Status.OperationContext = nil
}

func (m *ObTenantBackupPolicyManager) UpdateStatus() error {
	// TODO: check status of jobs to update BackupPolicy status
	err := m.Client.Status().Update(m.Ctx, m.BackupPolicy)
	if err != nil {
		m.Logger.Error(err, "Got error when update observer status")
	}
	return err
}

func (m *ObTenantBackupPolicyManager) GetTaskFunc(name string) (func() error, error) {
	switch name {
	case taskname.ConfigureServerForBackup:
		return m.ConfigureServerForBackup, nil
	case taskname.GetTenantInfo:
		return m.GetTenantInfo, nil
	case taskname.StartBackupJob:
		return m.StartBackup, nil
	case taskname.StopBackupJob:
		return m.StopBackup, nil
	default:
		return nil, errors.Errorf("unknown task name %s", name)
	}
}

func (m *ObTenantBackupPolicyManager) GetTaskFlow() (*task.TaskFlow, error) {
	// exists unfinished task flow, return the last task flow
	if m.BackupPolicy.Status.OperationContext != nil {
		m.Logger.Info("get task flow from BackupPolicy status")
		return task.NewTaskFlow(m.BackupPolicy.Status.OperationContext), nil
	}
	status := m.BackupPolicy.Status.Status
	// get task flow depending on BackupPolicy status
	switch status {
	case v1alpha1.BackupPolicyStatusPreparing:
		return task.GetRegistry().Get(flow.PrepareBackupPolicy)
	case v1alpha1.BackupPolicyStatusPrepared:
		return task.GetRegistry().Get(flow.StartBackupJob)
	case v1alpha1.BackupPolicyStatusRunning:
		return nil, nil
	default:
		return nil, nil
	}
}
