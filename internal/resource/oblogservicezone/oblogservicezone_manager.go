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

package oblogservicezone

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	lsstatus "github.com/oceanbase/ob-operator/internal/const/status/oblogservicecluster"
	nodestatus "github.com/oceanbase/ob-operator/internal/const/status/oblogservicenode"
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/oblogservicezone"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	opresource "github.com/oceanbase/ob-operator/pkg/coordinator"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

var _ opresource.ResourceManager = &OBLogServiceZoneManager{}

type OBLogServiceZoneManager struct {
	Ctx      context.Context
	Resource *v1alpha1.OBLogServiceZone
	Client   client.Client
	Recorder telemetry.Recorder
	Logger   *logr.Logger
}

func (m *OBLogServiceZoneManager) GetMeta() metav1.Object {
	return m.Resource.GetObjectMeta()
}

func (m *OBLogServiceZoneManager) GetStatus() string {
	return m.Resource.Status.Status
}

func (m *OBLogServiceZoneManager) InitStatus() {
	m.Logger.Info("Newly created log service zone, init status")
	controllerutil.AddFinalizer(m.Resource, oceanbaseconst.FinalizerOBLogServiceZone)
	m.Resource.Status = v1alpha1.OBLogServiceZoneStatus{
		Status: zonestatus.New,
	}
}

func (m *OBLogServiceZoneManager) SetOperationContext(c *tasktypes.OperationContext) {
	m.Resource.Status.OperationContext = c
}

func (m *OBLogServiceZoneManager) GetTaskFlow() (*tasktypes.TaskFlow, error) {
	if m.Resource.Status.OperationContext != nil {
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Get task flow from status")
		return tasktypes.NewTaskFlow(m.Resource.Status.OperationContext), nil
	}

	var taskFlow *tasktypes.TaskFlow
	switch m.Resource.Status.Status {
	case zonestatus.New:
		clusterName := m.Resource.Labels[oceanbaseconst.LabelRefOBLogServiceCluster]
		lsCluster := &v1alpha1.OBLogServiceCluster{}
		err := m.Client.Get(m.Ctx, client.ObjectKey{
			Namespace: m.Resource.Namespace,
			Name:      clusterName,
		}, lsCluster)
		if err != nil {
			if kubeerrors.IsNotFound(err) {
				taskFlow = genCreateNodesFlow(m)
			} else {
				return nil, err
			}
		} else if lsCluster.Status.Status == lsstatus.New {
			taskFlow = genPrepareZoneForBootstrapFlow(m)
		} else {
			taskFlow = genCreateNodesFlow(m)
		}
	case zonestatus.BootstrapReady:
		taskFlow = genMaintainZoneAfterBootstrapFlow(m)
	case zonestatus.AddNode:
		taskFlow = genAddNodeFlow(m)
	case zonestatus.DeleteNode:
		taskFlow = genDeleteNodeFlow(m)
	case zonestatus.Deleting:
		taskFlow = genDeleteZoneFlow(m)
	default:
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("No need to run anything for log service zone")
		return nil, nil
	}

	if taskFlow.OperationContext.OnFailure.Strategy == "" {
		taskFlow.OperationContext.OnFailure.Strategy = strategy.StartOver
		if taskFlow.OperationContext.OnFailure.NextTryStatus == "" {
			taskFlow.OperationContext.OnFailure.NextTryStatus = zonestatus.Failed
		}
	}
	return taskFlow, nil
}

func (m *OBLogServiceZoneManager) CheckAndUpdateFinalizers() error {
	if m.Resource.Status.Status == zonestatus.FinalizerFinished {
		m.Resource.ObjectMeta.Finalizers = make([]string, 0)
		return m.Client.Update(m.Ctx, m.Resource)
	}
	// Transition to Deleting to run the delete flow instead of skipping cleanup
	m.Resource.Status.Status = zonestatus.Deleting
	return m.Client.Status().Update(m.Ctx, m.Resource)
}

func (m *OBLogServiceZoneManager) UpdateStatus() error {
	if m.Resource.DeletionTimestamp != nil {
		m.Resource.Status.Status = zonestatus.Deleting
		return m.Client.Status().Update(m.Ctx, m.Resource)
	}

	nodeList, err := m.listNodes()
	if err != nil {
		m.Logger.Error(err, "list nodes error")
	}

	if nodeList == nil {
		return m.Client.Status().Update(m.Ctx, m.Resource)
	}
	nodeStatusList := make([]apitypes.LogServiceNodeReplicaStatus, 0, len(nodeList.Items))
	availableNodes := 0
	for _, node := range nodeList.Items {
		nodeStatusList = append(nodeStatusList, apitypes.LogServiceNodeReplicaStatus{
			NodeName: node.Name,
			Status:   node.Status.Status,
		})
		if node.Status.Status == nodestatus.Running {
			availableNodes++
		}
	}
	m.Resource.Status.NodeStatus = nodeStatusList

	if m.Resource.Status.Status == zonestatus.Running {
		expectedReplica := m.Resource.Spec.Topology.Replica
		if expectedReplica > availableNodes {
			m.Resource.Status.Status = zonestatus.AddNode
		} else if len(nodeList.Items) > expectedReplica {
			m.Resource.Status.Status = zonestatus.DeleteNode
		}
	}

	return m.Client.Status().Update(m.Ctx, m.Resource)
}

func (m *OBLogServiceZoneManager) ClearTaskInfo() {
	m.Resource.Status.Status = zonestatus.Running
	m.Resource.Status.OperationContext = nil
}

func (m *OBLogServiceZoneManager) FinishTask() {
	m.Resource.Status.Status = m.Resource.Status.OperationContext.TargetStatus
	m.Resource.Status.OperationContext = nil
}

func (m *OBLogServiceZoneManager) HandleFailure() {
	if m.Resource.DeletionTimestamp != nil {
		m.Resource.Status.Status = zonestatus.Deleting
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
		case strategy.Pause:
		}
	}
}

func (m *OBLogServiceZoneManager) GetTaskFunc(name tasktypes.TaskName) (tasktypes.TaskFunc, error) {
	return taskMap.GetTask(name, m)
}

func (m *OBLogServiceZoneManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.Resource, corev1.EventTypeWarning, "Task failed", err.Error())
}

func (m *OBLogServiceZoneManager) ArchiveResource() {
	m.Logger.Info("Archive log service zone", "name", m.Resource.Name)
	m.Recorder.Event(m.Resource, "Archive", "", "Archive log service zone")
	m.Resource.Status.Status = zonestatus.Failed
	m.Resource.Status.OperationContext = nil
}

func (m *OBLogServiceZoneManager) listNodes() (*v1alpha1.OBLogServiceNodeList, error) {
	nodeList := &v1alpha1.OBLogServiceNodeList{}
	err := m.Client.List(m.Ctx, nodeList, client.MatchingLabels{
		oceanbaseconst.LabelRefOBLogServiceZone: m.Resource.Name,
	}, client.InNamespace(m.Resource.Namespace))
	if err != nil {
		return nil, err
	}
	return nodeList, nil
}
