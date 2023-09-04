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
	"github.com/oceanbase/ob-operator/api/constants"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	"github.com/oceanbase/ob-operator/pkg/task"
	flow "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	policy := m.BackupPolicy
	finalizerName := "obtenantbackuppolicy.finalizers.oceanbase.com"
	if controllerutil.ContainsFinalizer(policy, finalizerName) {
		err := m.StopBackup()
		if err != nil {
			return err
		}
		// remove our finalizer from the list and update it.
		controllerutil.RemoveFinalizer(policy, finalizerName)
		if err := m.Client.Update(m.Ctx, policy); err != nil {
			return err
		}
	}
	return nil
}

func (m *ObTenantBackupPolicyManager) InitStatus() {
	m.BackupPolicy.Status = v1alpha1.OBTenantBackupPolicyStatus{
		Status:                 constants.BackupPolicyStatusPreparing,
		LogArchiveDestDisabled: false,
	}
}

func (m *ObTenantBackupPolicyManager) SetOperationContext(c *v1alpha1.OperationContext) {
	m.BackupPolicy.Status.OperationContext = c
}

func (m *ObTenantBackupPolicyManager) ClearTaskInfo() {
	m.BackupPolicy.Status.Status = constants.BackupPolicyStatusRunning
	m.BackupPolicy.Status.OperationContext = nil
}

func (m *ObTenantBackupPolicyManager) FinishTask() {
	m.BackupPolicy.Status.Status = constants.BackupPolicyStatusType(m.BackupPolicy.Status.OperationContext.TargetStatus)
	m.BackupPolicy.Status.OperationContext = nil
}

func (m *ObTenantBackupPolicyManager) UpdateStatus() error {
	return m.Client.Status().Update(m.Ctx, m.BackupPolicy)
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
	case taskname.CheckAndSpawnJobs:
		return m.CheckAndSpawnJobs, nil
	default:
		return nil, errors.Errorf("unknown task name %s", name)
	}
}

func (m *ObTenantBackupPolicyManager) GetTaskFlow() (*task.TaskFlow, error) {
	// exists unfinished task flow, return the last task flow
	if m.BackupPolicy.Status.OperationContext != nil {
		return task.NewTaskFlow(m.BackupPolicy.Status.OperationContext), nil
	}
	status := m.BackupPolicy.Status.Status
	// get task flow depending on BackupPolicy status
	switch status {
	case constants.BackupPolicyStatusPreparing:
		return task.GetRegistry().Get(flow.PrepareBackupPolicy)
	case constants.BackupPolicyStatusPrepared:
		return task.GetRegistry().Get(flow.StartBackupJob)
	case constants.BackupPolicyStatusRunning:
		return task.GetRegistry().Get(flow.MaintainCrontab)
	default:
		return nil, nil
	}
}
