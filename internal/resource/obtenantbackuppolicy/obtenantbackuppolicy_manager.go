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

package obtenantbackuppolicy

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

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

var _ opresource.ResourceManager = &ObTenantBackupPolicyManager{}

type ObTenantBackupPolicyManager struct {
	Ctx          context.Context
	BackupPolicy *v1alpha1.OBTenantBackupPolicy
	Client       client.Client
	Recorder     telemetry.Recorder
	Logger       *logr.Logger
}

func (m *ObTenantBackupPolicyManager) GetMeta() metav1.Object {
	return m.BackupPolicy.GetObjectMeta()
}

func (m *ObTenantBackupPolicyManager) GetStatus() string {
	return string(m.BackupPolicy.Status.Status)
}

func (m *ObTenantBackupPolicyManager) CheckAndUpdateFinalizers() error {
	policy := m.BackupPolicy
	finalizerName := oceanbaseconst.FinalizerBackupPolicy
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
			if m.BackupPolicy.Spec.TenantCRName != "" {
				tenant, err := m.getOBTenantCR()
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
				}
			}
			finalizerFinished = finalizerFinished || m.BackupPolicy.Status.Status == constants.BackupPolicyStatusStopped
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
	m.Recorder.Event(m.BackupPolicy, "Init", "", "Init status")
	err = m.syncTenantInformation()
	if err != nil {
		m.PrintErrEvent(err)
	}
}

func (m *ObTenantBackupPolicyManager) syncTenantInformation() error {
	if m.BackupPolicy.Spec.TenantCRName != "" {
		tenant := &v1alpha1.OBTenant{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.BackupPolicy.Namespace,
			Name:      m.BackupPolicy.Spec.TenantCRName,
		}, tenant)
		if err != nil {
			return err
		}
		m.BackupPolicy.Status.TenantName = tenant.Spec.TenantName
	}

	tenantRecord, err := m.getTenantRecord(false)
	if err != nil {
		return err
	}
	m.BackupPolicy.Status.TenantInfo = tenantRecord
	return nil
}

func (m *ObTenantBackupPolicyManager) SetOperationContext(c *tasktypes.OperationContext) {
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
	} else if m.BackupPolicy.DeletionTimestamp != nil {
		switch m.BackupPolicy.Status.Status {
		case constants.BackupPolicyStatusPaused,
			constants.BackupPolicyStatusRunning,
			constants.BackupPolicyStatusMaintaining,
			constants.BackupPolicyStatusPausing,
			constants.BackupPolicyStatusPrepared,
			constants.BackupPolicyStatusResuming:
			m.BackupPolicy.Status.Status = constants.BackupPolicyStatusDeleting
			m.BackupPolicy.Status.OperationContext = nil
		case constants.BackupPolicyStatusDeleting:
			// do nothing
		default:
			m.BackupPolicy.Status.Status = constants.BackupPolicyStatusStopped
			m.BackupPolicy.Status.OperationContext = nil
		}
	} else if m.BackupPolicy.Status.Status == constants.BackupPolicyStatusRunning {
		if m.BackupPolicy.GetGeneration() > m.BackupPolicy.Status.ObservedGeneration {
			m.BackupPolicy.Status.Status = constants.BackupPolicyStatusMaintaining
		} else {
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
			var backupPath string
			if m.BackupPolicy.Spec.DataBackup.Destination.Type == constants.BackupDestTypeOSS {
				backupPath = m.BackupPolicy.Spec.DataBackup.Destination.Path
			} else {
				backupPath = m.getBackupDestPath()
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
				m.BackupPolicy.Status.Status = constants.BackupPolicyStatusMaintaining
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
					var nextIncrTime time.Time
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

							nextIncrTime = incrCron.Next(lastIncrBackupFinishedAt)
							m.BackupPolicy.Status.NextIncremental = nextIncrTime.Format(time.DateTime)
						} else if latestIncr.Status == "INIT" || latestIncr.Status == "DOING" {
							// do nothing
							_ = latestIncr
						} else {
							m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Incremental BackupJob are in status " + latestIncr.Status)
						}
					} else {
						nextIncrTime = incrCron.Next(lastFullBackupFinishedAt)
						m.BackupPolicy.Status.NextIncremental = nextIncrTime.Format(time.DateTime)
					}
					if !resourceutils.IsZero(nextIncrTime) && nextIncrTime.Before(time.Now()) {
						m.BackupPolicy.Status.Status = constants.BackupPolicyStatusMaintaining
					}
				} else {
					m.BackupPolicy.Status.Status = constants.BackupPolicyStatusMaintaining
				}
			}
		}
	}

	m.BackupPolicy.Status.ObservedGeneration = m.BackupPolicy.GetGeneration()
	return m.retryUpdateStatus()
}

func (m *ObTenantBackupPolicyManager) GetTaskFunc(name tasktypes.TaskName) (tasktypes.TaskFunc, error) {
	return taskMap.GetTask(name, m)
}

func (m *ObTenantBackupPolicyManager) GetTaskFlow() (*tasktypes.TaskFlow, error) {
	// exists unfinished task flow, return the last task flow
	if m.BackupPolicy.Status.OperationContext != nil {
		return tasktypes.NewTaskFlow(m.BackupPolicy.Status.OperationContext), nil
	}
	var taskFlow *tasktypes.TaskFlow
	status := m.BackupPolicy.Status.Status
	// get task flow depending on BackupPolicy status
	switch status {
	case constants.BackupPolicyStatusPreparing:
		taskFlow = genPrepareBackupPolicyFlow(m)
	case constants.BackupPolicyStatusPrepared:
		taskFlow = genStartBackupJobFlow(m)
	case constants.BackupPolicyStatusMaintaining:
		taskFlow = genMaintainRunningPolicyFlow(m)
	case constants.BackupPolicyStatusPausing:
		taskFlow = genPauseBackupFlow(m)
	case constants.BackupPolicyStatusResuming:
		taskFlow = genResumeBackupFlow(m)
	case constants.BackupPolicyStatusDeleting:
		taskFlow = genStopBackupPolicyFlow(m)
	default:
		// Paused, Stopped or Failed
		return nil, nil
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
	if m.BackupPolicy.DeletionTimestamp != nil {
		m.BackupPolicy.Status.OperationContext = nil
	} else {
		operationContext := m.BackupPolicy.Status.OperationContext
		failureRule := operationContext.OnFailure
		switch failureRule.Strategy {
		case "":
			fallthrough
		case strategy.StartOver:
			if m.BackupPolicy.Status.Status != apitypes.BackupPolicyStatusType(failureRule.NextTryStatus) {
				m.BackupPolicy.Status.Status = apitypes.BackupPolicyStatusType(failureRule.NextTryStatus)
				m.BackupPolicy.Status.OperationContext = nil
			} else {
				m.BackupPolicy.Status.OperationContext.Idx = 0
				m.BackupPolicy.Status.OperationContext.TaskStatus = ""
				m.BackupPolicy.Status.OperationContext.TaskId = ""
				m.BackupPolicy.Status.OperationContext.Task = ""
			}
		case strategy.RetryFromCurrent:
			operationContext.TaskStatus = taskstatus.Pending
		case strategy.Pause:
		}
	}
}

func (m *ObTenantBackupPolicyManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.BackupPolicy, corev1.EventTypeWarning, "Task failed", err.Error())
}

func (m *ObTenantBackupPolicyManager) ArchiveResource() {
	m.Logger.Info("Archive obtenant backup policy", "obtenant backup policy", m.BackupPolicy.Name)
	m.Recorder.Event(m.BackupPolicy, "Archive", "", "Archive obtenant backup policy")
	m.BackupPolicy.Status.Status = "Failed"
	m.BackupPolicy.Status.OperationContext = nil
}

func (m *ObTenantBackupPolicyManager) getOBCluster() (*v1alpha1.OBCluster, error) {
	clusterName := m.BackupPolicy.Spec.ObClusterName
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.BackupPolicy.Namespace,
		Name:      clusterName,
	}, obcluster)
	if err != nil {
		m.Logger.Error(err, "Failed to get obcluster", "clusterName", clusterName, "namespaced", m.BackupPolicy.Namespace)
		return nil, errors.Wrap(err, "get obcluster failed")
	}
	return obcluster, nil
}

func (m *ObTenantBackupPolicyManager) getOBTenantCR() (*v1alpha1.OBTenant, error) {
	// Guard that tenantCRName is not empty
	if m.BackupPolicy.Spec.TenantCRName == "" {
		return nil, kubeerrors.NewNotFound(schema.GroupResource{
			Group:    "oceanbase.oceanbase.com",
			Resource: "obtenantbackuppolicies",
		}, m.BackupPolicy.Spec.TenantCRName)
	}
	tenantCRName := m.BackupPolicy.Spec.TenantCRName
	tenant := &v1alpha1.OBTenant{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.BackupPolicy.Namespace,
		Name:      tenantCRName,
	}, tenant)
	if err != nil {
		if !kubeerrors.IsNotFound(err) {
			m.Logger.Error(err, "Failed to get obtenant", "tenantCRName", tenantCRName, "namespaced", m.BackupPolicy.Namespace)
		}
		return nil, err
	}
	return tenant, nil
}
