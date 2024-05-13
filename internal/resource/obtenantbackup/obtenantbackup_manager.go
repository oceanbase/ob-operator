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

package obtenantbackup

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	opresource "github.com/oceanbase/ob-operator/pkg/coordinator"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

var _ opresource.ResourceManager = &OBTenantBackupManager{}

type OBTenantBackupManager struct {
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
	ignoreDel, ok := resourceutils.GetAnnotationField(m.Resource, oceanbaseconst.AnnotationsIgnoreDeletion)
	return !m.Resource.ObjectMeta.DeletionTimestamp.IsZero() && (!ok || ignoreDel != "true")
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

func (m *OBTenantBackupManager) SetOperationContext(c *tasktypes.OperationContext) {
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
			if m.Resource.Status.Status != apitypes.BackupJobStatus(failureRule.NextTryStatus) {
				m.Resource.Status.Status = apitypes.BackupJobStatus(failureRule.NextTryStatus)
				m.Resource.Status.OperationContext = nil
			} else {
				m.Resource.Status.OperationContext.Idx = 0
				m.Resource.Status.OperationContext.TaskStatus = ""
				m.Resource.Status.OperationContext.TaskId = ""
				m.Resource.Status.OperationContext.Task = ""
			}
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

func (m *OBTenantBackupManager) GetTaskFunc(name tasktypes.TaskName) (tasktypes.TaskFunc, error) {
	return taskMap.GetTask(name, m)
}

func (m *OBTenantBackupManager) GetTaskFlow() (*tasktypes.TaskFlow, error) {
	// exists unfinished task flow, return the last task flow
	if m.Resource.Status.OperationContext != nil {
		return tasktypes.NewTaskFlow(m.Resource.Status.OperationContext), nil
	}
	var taskFlow *tasktypes.TaskFlow
	if m.Resource.Status.Status == constants.BackupJobStatusInitializing {
		switch m.Resource.Spec.Type {
		case constants.BackupJobTypeFull, constants.BackupJobTypeIncr:
			taskFlow = genCreateBackupJobInDBFlow(m)
		}
	}
	return taskFlow, nil
}

func (m *OBTenantBackupManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.Resource, corev1.EventTypeWarning, "Task failed", err.Error())
}

func (m *OBTenantBackupManager) ArchiveResource() {
	m.Logger.Info("Archive obtenant backup job", "obtenant backup job", m.Resource.Name)
	m.Recorder.Event(m.Resource, "Archive", "", "Archive obtenant backup job")
	m.Resource.Status.Status = "Failed"
	m.Resource.Status.OperationContext = nil
}
