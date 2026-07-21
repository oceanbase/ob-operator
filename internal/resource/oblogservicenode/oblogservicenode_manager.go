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

package oblogservicenode

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	lsstatus "github.com/oceanbase/ob-operator/internal/const/status/oblogservicecluster"
	nodestatus "github.com/oceanbase/ob-operator/internal/const/status/oblogservicenode"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	opresource "github.com/oceanbase/ob-operator/pkg/coordinator"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

var _ opresource.ResourceManager = &OBLogServiceNodeManager{}

type OBLogServiceNodeManager struct {
	Ctx       context.Context
	Resource  *v1alpha1.OBLogServiceNode
	Client    client.Client
	APIReader client.Reader
	Recorder  telemetry.Recorder
	Logger    *logr.Logger
}

func (m *OBLogServiceNodeManager) GetMeta() metav1.Object {
	return m.Resource.GetObjectMeta()
}

func (m *OBLogServiceNodeManager) GetStatus() string {
	return m.Resource.Status.Status
}

func (m *OBLogServiceNodeManager) InitStatus() {
	m.Logger.Info("Newly created log service node, init status")
	controllerutil.AddFinalizer(m.Resource, oceanbaseconst.FinalizerOBLogServiceNode)
	m.Resource.Status = v1alpha1.OBLogServiceNodeStatus{
		Status:  nodestatus.New,
		PodName: m.Resource.Name,
	}
}

func (m *OBLogServiceNodeManager) SetOperationContext(c *tasktypes.OperationContext) {
	m.Resource.Status.OperationContext = c
}

func (m *OBLogServiceNodeManager) GetTaskFlow() (*tasktypes.TaskFlow, error) {
	if m.Resource.Status.OperationContext != nil {
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Get task flow from status")
		return tasktypes.NewTaskFlow(m.Resource.Status.OperationContext), nil
	}

	var taskFlow *tasktypes.TaskFlow
	switch m.Resource.Status.Status {
	case nodestatus.New:
		clusterName := m.Resource.Spec.ClusterName
		lsCluster := &v1alpha1.OBLogServiceCluster{}
		err := m.Client.Get(m.Ctx, client.ObjectKey{
			Namespace: m.Resource.Namespace,
			Name:      clusterName,
		}, lsCluster)
		if err != nil {
			if kubeerrors.IsNotFound(err) {
				taskFlow = genCreateNodeFlow(m)
			} else {
				return nil, err
			}
		} else if lsCluster.Status.Status == lsstatus.New {
			m.Logger.Info("Prepare log service node for bootstrap")
			taskFlow = genPrepareNodeForBootstrapFlow(m)
		} else {
			m.Logger.Info("Create log service node (cluster already bootstrapped)")
			taskFlow = genCreateNodeFlow(m)
		}
	case nodestatus.BootstrapReady:
		taskFlow = genMaintainNodeAfterBootstrapFlow(m)
	case nodestatus.Recover:
		taskFlow = genRecoverNodeFlow(m)
	case nodestatus.Deleting:
		taskFlow = genDeleteNodeFlow(m)
	default:
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("No need to run anything for log service node")
		return nil, nil
	}

	if taskFlow.OperationContext.OnFailure.Strategy == "" {
		taskFlow.OperationContext.OnFailure.Strategy = strategy.StartOver
		if taskFlow.OperationContext.OnFailure.NextTryStatus == "" {
			taskFlow.OperationContext.OnFailure.NextTryStatus = nodestatus.Failed
		}
	}
	return taskFlow, nil
}

func (m *OBLogServiceNodeManager) CheckAndUpdateFinalizers() error {
	if m.Resource.Status.Status == nodestatus.FinalizerFinished {
		m.Resource.Finalizers = make([]string, 0)
		return m.Client.Update(m.Ctx, m.Resource)
	}
	// Transition to Deleting to run the delete flow instead of skipping cleanup
	m.Resource.Status.Status = nodestatus.Deleting
	return m.Client.Status().Update(m.Ctx, m.Resource)
}

func (m *OBLogServiceNodeManager) UpdateStatus() error {
	if m.Resource.DeletionTimestamp != nil {
		m.Resource.Status.Status = nodestatus.Deleting
		return m.Client.Status().Update(m.Ctx, m.Resource)
	}

	// Populate ServiceIP from existing Service if not yet set (service mode only)
	if m.Resource.Status.ServiceIP == "" {
		mode, modeAnnoExist := resourceutils.GetAnnotationField(m.Resource, oceanbaseconst.AnnotationsMode)
		if modeAnnoExist && mode == oceanbaseconst.ModeService {
			svcName := fmt.Sprintf("%s-svc", m.Resource.Name)
			svc := &corev1.Service{}
			if err := m.Client.Get(m.Ctx, types.NamespacedName{
				Namespace: m.Resource.Namespace,
				Name:      svcName,
			}, svc); err == nil {
				m.Resource.Status.ServiceIP = svc.Spec.ClusterIP
			}
		}
	}

	// Refresh PodIP/phase from the current pod whenever it exists, so that
	// bootstrap (which needs the node connect address) can read the PodIP even
	// before the node reaches Running. This mirrors the unconditional ServiceIP
	// refresh above and is required for Pod-IP mode where the connect address
	// is the PodIP itself.
	pod := &corev1.Pod{}
	if m.Resource.Status.PodName != "" {
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.Resource.Namespace,
			Name:      m.Resource.Status.PodName,
		}, pod)
		if err != nil && !kubeerrors.IsNotFound(err) {
			return err
		}
		if err == nil {
			m.Resource.Status.PodPhase = pod.Status.Phase
			// A failed/evicted pod may report an empty PodIP; keep the last
			// known IP so recovery can pin the recreated pod to it.
			if pod.Status.PodIP != "" {
				m.Resource.Status.PodIP = pod.Status.PodIP
			}
		}
	}

	if m.Resource.Status.Status == nodestatus.Running {
		if m.Resource.Status.PodName != "" && pod.Name == "" {
			// Pod lookup above returned NotFound; need recovery.
			m.Logger.Info("LogService node pod not found, need recovery", "pod", m.Resource.Status.PodName)
			m.setRecoveryStatus()
		} else if pod.Name != "" {
			m.Resource.Status.Ready = pod.Status.Phase == corev1.PodRunning
			if pod.Status.Phase == corev1.PodFailed {
				m.Logger.Info("LogService node pod in Failed phase, need recovery", "pod", m.Resource.Status.PodName)
				m.setRecoveryStatus()
			}
		}
	}
	return m.Client.Status().Update(m.Ctx, m.Resource)
}

func (m *OBLogServiceNodeManager) ClearTaskInfo() {
	m.Resource.Status.Status = nodestatus.Running
	m.Resource.Status.OperationContext = nil
}

func (m *OBLogServiceNodeManager) FinishTask() {
	m.Resource.Status.Status = m.Resource.Status.OperationContext.TargetStatus
	m.Resource.Status.OperationContext = nil
}

func (m *OBLogServiceNodeManager) HandleFailure() {
	if m.Resource.DeletionTimestamp != nil {
		m.Resource.Status.Status = nodestatus.Deleting
		m.Resource.Status.OperationContext = nil
	} else {
		operationContext := m.Resource.Status.OperationContext
		failureRule := operationContext.OnFailure
		switch failureRule.Strategy {
		case strategy.StartOver:
			if m.Resource.Status.Status != failureRule.NextTryStatus {
				m.Resource.Status.Status = failureRule.NextTryStatus
				m.Resource.Status.OperationContext = nil
			} else {
				m.Resource.Status.OperationContext.Idx = 0
				m.Resource.Status.OperationContext.TaskStatus = ""
				m.Resource.Status.OperationContext.TaskId = ""
				m.Resource.Status.OperationContext.Task = ""
			}
		case strategy.RetryFromCurrent:
			operationContext.TaskStatus = taskstatus.Pending
		default:
			// strategy.Pause intentionally does nothing
		}
	}
}

func (m *OBLogServiceNodeManager) setRecoveryStatus() {
	if m.Resource.SupportStaticIP() {
		m.Logger.Info("LogService node supports static IP, recovering by recreating pod")
		m.Resource.Status.Status = nodestatus.Recover
	} else {
		m.Logger.Info("LogService node does not support static IP, marking unrecoverable")
		m.Resource.Status.Status = nodestatus.Unrecoverable
	}
}

func (m *OBLogServiceNodeManager) GetTaskFunc(name tasktypes.TaskName) (tasktypes.TaskFunc, error) {
	return taskMap.GetTask(name, m)
}

func (m *OBLogServiceNodeManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.Resource, corev1.EventTypeWarning, "Task failed", err.Error())
}

func (m *OBLogServiceNodeManager) ArchiveResource() {
	m.Logger.Info("Archive log service node", "name", m.Resource.Name)
	m.Recorder.Event(m.Resource, "Archive", "", "Archive log service node")
	m.Resource.Status.Status = nodestatus.Failed
	m.Resource.Status.OperationContext = nil
}
