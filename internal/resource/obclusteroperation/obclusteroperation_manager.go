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

package obclusteroperation

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	opresource "github.com/oceanbase/ob-operator/pkg/coordinator"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

var _ opresource.ResourceManager = &OBClusterOperationManager{}

type OBClusterOperationManager struct {
	Ctx      context.Context
	Resource *v1alpha1.OBClusterOperation
	Client   client.Client
	Recorder telemetry.Recorder
	Logger   *logr.Logger
}

func (m *OBClusterOperationManager) GetMeta() metav1.Object {
	return m.Resource.GetObjectMeta()
}

func (m *OBClusterOperationManager) GetStatus() string {
	return string(m.Resource.Status.Status)
}

func (m *OBClusterOperationManager) CheckAndUpdateFinalizers() error {
	return nil
}

func (m *OBClusterOperationManager) InitStatus() {
	m.Resource.Status.Status = constants.ClusterOpStatusRunning
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.Resource.Namespace,
		Name:      m.Resource.Spec.OBCluster,
	}, obcluster)
	if err != nil {
		m.Logger.V(oceanbaseconst.LogLevelDebug).WithValues("err", err).Info("Failed to find obcluster")
		return
	}
	m.Resource.Status.ClusterSnapshot = &v1alpha1.OBClusterSnapshot{
		Spec:   &obcluster.Spec,
		Status: &obcluster.Status,
	}
}

func (m *OBClusterOperationManager) SetOperationContext(c *tasktypes.OperationContext) {
	m.Resource.Status.OperationContext = c
}

func (m *OBClusterOperationManager) ClearTaskInfo() {
	m.Resource.Status.Status = constants.ClusterOpStatusRunning
	m.Resource.Status.OperationContext = nil
}

func (m *OBClusterOperationManager) HandleFailure() {
	if m.Resource.DeletionTimestamp != nil {
		m.Resource.Status.OperationContext = nil
	} else {
		operationContext := m.Resource.Status.OperationContext
		failureRule := operationContext.OnFailure
		switch failureRule.Strategy {
		case strategy.StartOver:
			if m.Resource.Status.Status != apitypes.ClusterOperationStatus(failureRule.NextTryStatus) {
				m.Resource.Status.Status = apitypes.ClusterOperationStatus(failureRule.NextTryStatus)
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
		default:
			m.Resource.Status.OperationContext = nil
			if failureRule.NextTryStatus == "" {
				m.Resource.Status.Status = constants.ClusterOpStatusFailed
			} else {
				m.Resource.Status.Status = apitypes.ClusterOperationStatus(failureRule.NextTryStatus)
			}
		}
	}
}

func (m *OBClusterOperationManager) FinishTask() {
	m.Resource.Status.Status = apitypes.ClusterOperationStatus(m.Resource.Status.OperationContext.TargetStatus)
	m.Resource.Status.OperationContext = nil
}

func (m *OBClusterOperationManager) UpdateStatus() error {
	return m.retryUpdateStatus()
}

func (m *OBClusterOperationManager) ArchiveResource() {
	m.Logger.Info("Archive obcluster operation", "obcluster operation", m.Resource.Name)
	m.Recorder.Event(m.Resource, "Archive", "", "Archive obcluster operation")
	m.Resource.Status.Status = constants.ClusterOpStatusFailed
	m.Resource.Status.OperationContext = nil
}

func (m *OBClusterOperationManager) GetTaskFunc(name tasktypes.TaskName) (tasktypes.TaskFunc, error) {
	return taskMap.GetTask(name, m)
}

func (m *OBClusterOperationManager) GetTaskFlow() (*tasktypes.TaskFlow, error) {
	if m.Resource.Status.OperationContext != nil {
		return tasktypes.NewTaskFlow(m.Resource.Status.OperationContext), nil
	}
	var taskFlow *tasktypes.TaskFlow
	status := m.Resource.Status.Status
	switch status {
	case constants.ClusterOpStatusRunning:
		if strings.EqualFold(string(m.Resource.Spec.Type), string(constants.ClusterOpTypeRestartOBServers)) &&
			m.Resource.Spec.RestartOBServers != nil && m.Resource.Spec.RestartOBServers.RestartOnly {
			taskFlow = genRestartOBServersOnlyFlow(m)
		} else {
			taskFlow = genModifySpecAndWatchFlow(m)
		}
	case constants.ClusterOpStatusPending,
		constants.ClusterOpStatusSucceeded,
		constants.ClusterOpStatusFailed:
		fallthrough
	default:
		return nil, nil
	}
	if taskFlow.OperationContext.OnFailure.NextTryStatus == "" {
		taskFlow.OperationContext.OnFailure.NextTryStatus = string(constants.TenantOpFailed)
	}
	return taskFlow, nil
}

func (m *OBClusterOperationManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.Resource, corev1.EventTypeWarning, "Task failed", err.Error())
}

func (m *OBClusterOperationManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		resource := &v1alpha1.OBClusterOperation{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.Resource.GetNamespace(),
			Name:      m.Resource.GetName(),
		}, resource)
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		resource.Status = m.Resource.Status
		return m.Client.Status().Update(m.Ctx, m.Resource)
	})
}
