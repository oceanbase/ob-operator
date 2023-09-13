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
	"time"

	"github.com/go-logr/logr"
	"github.com/oceanbase/ob-operator/api/constants"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	"github.com/oceanbase/ob-operator/pkg/task"
	flow "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
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
		Status: constants.BackupPolicyStatusPreparing,
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
	if m.BackupPolicy.Spec.Suspend && m.BackupPolicy.Status.Status == constants.BackupPolicyStatusRunning {
		m.BackupPolicy.Status.Status = constants.BackupPolicyStatusPausing
		m.BackupPolicy.Status.OperationContext = nil
	} else if !m.BackupPolicy.Spec.Suspend && m.BackupPolicy.Status.Status == constants.BackupPolicyStatusPaused {
		m.BackupPolicy.Status.Status = constants.BackupPolicyStatusResuming
	} else if m.BackupPolicy.Status.Status == constants.BackupPolicyStatusRunning {
		if m.BackupPolicy.Status.TenantInfo == nil {
			tenant, err := m.getTenantInfo()
			if err != nil {
				return err
			}
			m.BackupPolicy.Status.TenantInfo = tenant
		}
		err := m.syncLatestJobs()
		if err != nil {
			return err
		}
		backupPath := m.getBackupDestPath()
		latestFull, err := m.getLatestBackupJobOfTypeAndPath(constants.BackupJobTypeFull, backupPath)
		if err != nil {
			return err
		}
		latestIncr, err := m.getLatestBackupJobOfTypeAndPath(constants.BackupJobTypeIncr, backupPath)
		if err != nil {
			return err
		}
		m.BackupPolicy.Status.LatestFullBackupJob = latestFull
		m.BackupPolicy.Status.LatestIncrementalJob = latestIncr

		if latestFull == nil || latestFull.Status == "CANCELED" {
			m.BackupPolicy.Status.NextFull = time.Now().Format(time.DateTime)
		} else if latestFull.Status == "COMPLETED" {
			fullCron, err := cron.ParseStandard(m.BackupPolicy.Spec.DataBackup.FullCrontab)
			if err != nil {
				return err
			}
			var lastFullBackupFinishedAt time.Time
			if latestFull.EndTimestamp != nil {
				lastFullBackupFinishedAt, err = time.ParseInLocation(time.DateTime, *latestFull.EndTimestamp, time.Local)
			}
			nextFull := fullCron.Next(lastFullBackupFinishedAt)
			m.BackupPolicy.Status.NextFull = nextFull.Format(time.DateTime)
			if nextFull.After(time.Now()) {
				incrCron, err := cron.ParseStandard(m.BackupPolicy.Spec.DataBackup.IncrementalCrontab)
				if err != nil {
					return err
				}
				if latestIncr == nil || latestIncr.Status == "CANCELED" {
					m.BackupPolicy.Status.NextIncremental = time.Now().Format(time.DateTime)
				} else if latestIncr.Status == "COMPLETED" {
					var lastIncrBackupFinishedAt time.Time
					if latestIncr.EndTimestamp != nil {
						lastIncrBackupFinishedAt, err = time.ParseInLocation(time.DateTime, *latestIncr.EndTimestamp, time.Local)
					}
					m.BackupPolicy.Status.NextIncremental = incrCron.Next(lastIncrBackupFinishedAt).Format(time.DateTime)
				}
			}
		}
	}

	return m.retryUpdateStatus()
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
	case taskname.CleanOldBackupJobs:
		return m.CleanOldBackupJobs, nil
	case taskname.PauseBackup:
		return m.PauseBackup, nil
	case taskname.ResumeBackup:
		return m.ResumeBackup, nil
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
		return task.GetRegistry().Get(flow.MaintainRunningPolicy)
	case constants.BackupPolicyStatusPausing:
		return task.GetRegistry().Get(flow.PauseBackup)
	case constants.BackupPolicyStatusResuming:
		return task.GetRegistry().Get(flow.ResumeBackup)
	default:
		// Paused, Stopped or Failed
		return nil, nil
	}
}

func (m *ObTenantBackupPolicyManager) retryUpdateStatus() error {
	policy := &v1alpha1.OBTenantBackupPolicy{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.BackupPolicy.GetNamespace(),
		Name:      m.BackupPolicy.GetName(),
	}, policy)
	if err != nil {
		return client.IgnoreNotFound(err)
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		policy.Status = m.BackupPolicy.Status
		return m.Client.Status().Update(m.Ctx, policy)
	})
}
