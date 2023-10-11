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
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	"github.com/oceanbase/ob-operator/pkg/task"
	flow "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/task/status"
	"github.com/oceanbase/ob-operator/pkg/task/strategy"
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
	finalizerFinished := false
	if controllerutil.ContainsFinalizer(policy, finalizerName) {
		obcluster, err := m.getOBCluster()
		if err != nil {
			if kubeerrors.IsNotFound(err) {
				m.Logger.Info("OBCluster is deleted, no need to wait finalizer")
				finalizerFinished = true
			} else {
				m.Logger.Error(err, "query obcluster failed")
				return errors.Wrap(err, "Get obcluster failed")
			}
		} else if !obcluster.ObjectMeta.DeletionTimestamp.IsZero() {
			m.Logger.Info("OBCluster is deleting, no need to wait finalizer")
			finalizerFinished = true
		}

		if !finalizerFinished {
			tenant, err := m.getOBTenant()
			if err != nil {
				// the tenant is deleted, no need to wait finalizer
				if kubeerrors.IsNotFound(err) {
					finalizerFinished = true
				} else {
					return errors.Wrap(err, "Get obtenant failed")
				}
			} else if !tenant.GetDeletionTimestamp().IsZero() {
				// the tenant is being deleted
				finalizerFinished = true
			} else {
				err := m.StopBackup()
				// the policy is being deleted, connection still exists, stop backup
				if err != nil {
					return err
				}
				finalizerFinished = true
			}
		}

		if finalizerFinished {
			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(policy, finalizerName)
			if err := m.Client.Update(m.Ctx, policy); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *ObTenantBackupPolicyManager) InitStatus() {
	var err error
	m.BackupPolicy.Status = v1alpha1.OBTenantBackupPolicyStatus{
		Status: constants.BackupPolicyStatusPreparing,
	}
	err = m.syncTenantInformation()
	if err != nil {
		m.PrintErrEvent(err)
	}
}

func (m *ObTenantBackupPolicyManager) syncTenantInformation() error {
	tenant := &v1alpha1.OBTenant{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.BackupPolicy.Namespace,
		Name:      m.BackupPolicy.Spec.TenantName,
	}, tenant)
	if err != nil {
		return err
	}
	m.BackupPolicy.Status.TenantCR = tenant

	tenantRecord, err := m.getTenantRecord(false)
	if err != nil {
		return err
	}
	m.BackupPolicy.Status.TenantInfo = tenantRecord
	return nil
}

func (m *ObTenantBackupPolicyManager) SetOperationContext(c *v1alpha1.OperationContext) {
	m.BackupPolicy.Status.OperationContext = c
}

func (m *ObTenantBackupPolicyManager) ClearTaskInfo() {
	m.BackupPolicy.Status.Status = constants.BackupPolicyStatusRunning
	m.BackupPolicy.Status.OperationContext = nil
}

func (m *ObTenantBackupPolicyManager) FinishTask() {
	m.BackupPolicy.Status.Status = apitypes.BackupPolicyStatusType(m.BackupPolicy.Status.OperationContext.TargetStatus)
	m.BackupPolicy.Status.OperationContext = nil
}

func (m *ObTenantBackupPolicyManager) UpdateStatus() error {
	if m.BackupPolicy.Spec.Suspend && m.BackupPolicy.Status.Status == constants.BackupPolicyStatusRunning {
		m.BackupPolicy.Status.Status = constants.BackupPolicyStatusPausing
		m.BackupPolicy.Status.OperationContext = nil
	} else if !m.BackupPolicy.Spec.Suspend && m.BackupPolicy.Status.Status == constants.BackupPolicyStatusPaused {
		m.BackupPolicy.Status.Status = constants.BackupPolicyStatusResuming
	} else if m.BackupPolicy.Status.Status == constants.BackupPolicyStatusRunning {
		err := m.syncTenantInformation()
		if err != nil {
			m.PrintErrEvent(err)
			return err
		}
		err = m.syncLatestJobs()
		if err != nil {
			m.PrintErrEvent(err)
			return err
		}
		tenantRecordName, err := m.getTenantRecordName()
		if err != nil {
			m.Logger.Error(err, "Failed to get tenant record name")
			return err
		}
		var backupPath string
		if m.BackupPolicy.Spec.DataBackup.Destination.Type == constants.BackupDestTypeOSS {
			backupPath = m.BackupPolicy.Spec.DataBackup.Destination.Path
		} else {
			backupPath = m.getBackupDestPath(tenantRecordName)
		}

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
				if err != nil {
					return err
				}
			}
			nextFull := fullCron.Next(lastFullBackupFinishedAt)
			m.BackupPolicy.Status.NextFull = nextFull.Format(time.DateTime)
			if nextFull.After(time.Now()) {
				incrCron, err := cron.ParseStandard(m.BackupPolicy.Spec.DataBackup.IncrementalCrontab)
				if err != nil {
					return err
				}
				if latestIncr != nil {
					if latestIncr.Status == "COMPLETED" || latestIncr.Status == "CANCELED" {
						var lastIncrBackupFinishedAt time.Time
						if latestIncr.EndTimestamp == nil {
							// TODO: check if this is possible
							lastIncrBackupFinishedAt, err = time.ParseInLocation(time.DateTime, latestIncr.StartTimestamp, time.Local)
						} else {
							lastIncrBackupFinishedAt, err = time.ParseInLocation(time.DateTime, *latestIncr.EndTimestamp, time.Local)
						}
						if err != nil {
							m.Logger.Error(err, "Failed to parse end timestamp of completed backup job")
						}

						nextIncrTime := incrCron.Next(lastIncrBackupFinishedAt)
						m.BackupPolicy.Status.NextIncremental = nextIncrTime.Format(time.DateTime)
					} else if latestIncr.Status == "INIT" || latestIncr.Status == "DOING" {
						// do nothing
						_ = latestIncr
					} else {
						m.Logger.Info("Incremental BackupJob are in status " + latestIncr.Status)
					}
				} else {
					nextIncrTime := incrCron.Next(lastFullBackupFinishedAt)
					m.BackupPolicy.Status.NextIncremental = nextIncrTime.Format(time.DateTime)
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
	var taskFlow *task.TaskFlow
	var err error
	status := m.BackupPolicy.Status.Status
	// get task flow depending on BackupPolicy status
	switch status {
	case constants.BackupPolicyStatusPreparing:
		taskFlow, err = task.GetRegistry().Get(flow.PrepareBackupPolicy)
	case constants.BackupPolicyStatusPrepared:
		taskFlow, err = task.GetRegistry().Get(flow.StartBackupJob)
	case constants.BackupPolicyStatusRunning:
		taskFlow, err = task.GetRegistry().Get(flow.MaintainRunningPolicy)
	case constants.BackupPolicyStatusPausing:
		taskFlow, err = task.GetRegistry().Get(flow.PauseBackup)
	case constants.BackupPolicyStatusResuming:
		taskFlow, err = task.GetRegistry().Get(flow.ResumeBackup)
	default:
		// Paused, Stopped or Failed
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

func (m *ObTenantBackupPolicyManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		policy := &v1alpha1.OBTenantBackupPolicy{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.BackupPolicy.GetNamespace(),
			Name:      m.BackupPolicy.GetName(),
		}, policy)
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		policy.Status = m.BackupPolicy.Status
		return m.Client.Status().Update(m.Ctx, policy)
	})
}

func (m *ObTenantBackupPolicyManager) HandleFailure() {
	if m.IsDeleting() {
		m.BackupPolicy.Status.OperationContext = nil
	} else {
		operationContext := m.BackupPolicy.Status.OperationContext
		failureRule := operationContext.OnFailure
		switch failureRule.Strategy {
		case "":
			fallthrough
		case strategy.StartOver:
			m.BackupPolicy.Status.Status = apitypes.BackupPolicyStatusType(failureRule.NextTryStatus)
			m.BackupPolicy.Status.OperationContext = nil
		case strategy.RetryFromCurrent:
			operationContext.TaskStatus = taskstatus.Pending
		case strategy.Pause:
		}
	}
}

func (m *ObTenantBackupPolicyManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.BackupPolicy, corev1.EventTypeWarning, "task exec failed", err.Error())
}

func (m *ObTenantBackupPolicyManager) getOBCluster() (*v1alpha1.OBCluster, error) {
	clusterName := m.BackupPolicy.Spec.ObClusterName
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.BackupPolicy.Namespace,
		Name:      clusterName,
	}, obcluster)
	if err != nil {
		m.Logger.Error(err, "get obcluster failed", "clusterName", clusterName, "namespaced", m.BackupPolicy.Namespace)
		return nil, errors.Wrap(err, "get obcluster failed")
	}
	return obcluster, nil
}

func (m *ObTenantBackupPolicyManager) getOBTenant() (*v1alpha1.OBTenant, error) {
	tenantName := m.BackupPolicy.Spec.TenantName
	tenant := &v1alpha1.OBTenant{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.BackupPolicy.Namespace,
		Name:      tenantName,
	}, tenant)
	if err != nil {
		if !kubeerrors.IsNotFound(err) {
			m.Logger.Error(err, "get obtenant failed", "tenantName", tenantName, "namespaced", m.BackupPolicy.Namespace)
		}
		return nil, err
	}
	return tenant, nil
}
