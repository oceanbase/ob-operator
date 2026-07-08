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

package oblogservicecluster

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	lsstatus "github.com/oceanbase/ob-operator/internal/const/status/oblogservicecluster"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	opresource "github.com/oceanbase/ob-operator/pkg/coordinator"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

var _ opresource.ResourceManager = &OBLogServiceClusterManager{}

type OBLogServiceClusterManager struct {
	Ctx      context.Context
	Resource *v1alpha1.OBLogServiceCluster
	Client   client.Client
	Recorder telemetry.Recorder
	Logger   *logr.Logger
}

func (m *OBLogServiceClusterManager) GetMeta() metav1.Object {
	return m.Resource.GetObjectMeta()
}

func (m *OBLogServiceClusterManager) GetStatus() string {
	return m.Resource.Status.Status
}

func (m *OBLogServiceClusterManager) InitStatus() {
	m.Logger.Info("Newly created log service cluster, init status")
	m.Recorder.Event(m.Resource, "Init", "", "Newly created log service cluster, init status")
	controllerutil.AddFinalizer(m.Resource, oceanbaseconst.FinalizerOBLogService)
	m.Resource.Status = v1alpha1.OBLogServiceClusterStatus{
		Status: lsstatus.New,
	}
}

func (m *OBLogServiceClusterManager) SetOperationContext(c *tasktypes.OperationContext) {
	m.Resource.Status.OperationContext = c
}

func (m *OBLogServiceClusterManager) GetTaskFlow() (*tasktypes.TaskFlow, error) {
	if m.Resource.Status.OperationContext != nil {
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Get task flow from status")
		return tasktypes.NewTaskFlow(m.Resource.Status.OperationContext), nil
	}

	var taskFlow *tasktypes.TaskFlow
	m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Create task flow according to status")

	switch m.Resource.Status.Status {
	case lsstatus.New, lsstatus.Failed:
		taskFlow = genBootstrapLogServiceFlow(m)
	case lsstatus.ModifyZoneReplica:
		taskFlow = genModifyZoneReplicaFlow(m)
	default:
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("No need to run anything for log service cluster")
		return nil, nil
	}

	if taskFlow.OperationContext.OnFailure.Strategy == "" {
		taskFlow.OperationContext.OnFailure.Strategy = strategy.StartOver
		if taskFlow.OperationContext.OnFailure.NextTryStatus == "" {
			taskFlow.OperationContext.OnFailure.NextTryStatus = lsstatus.Failed
		}
	}
	return taskFlow, nil
}

func (m *OBLogServiceClusterManager) CheckAndUpdateFinalizers() error {
	if m.Resource.Status.Status == lsstatus.FinalizerFinished {
		m.Resource.Finalizers = make([]string, 0)
		return m.Client.Update(m.Ctx, m.Resource)
	}
	// Check if any OBCluster references this LogService
	obclusterList := &v1alpha1.OBClusterList{}
	if err := m.Client.List(m.Ctx, obclusterList, client.InNamespace(m.Resource.Namespace)); err != nil {
		return err
	}
	for _, cluster := range obclusterList.Items {
		if cluster.Spec.LogServiceRef != nil && cluster.Spec.LogServiceRef.Name == m.Resource.Name {
			m.Logger.Info("Cannot delete LogService: still referenced by OBCluster", "obcluster", cluster.Name)
			m.Recorder.Event(m.Resource, corev1.EventTypeWarning, "DeletionBlocked",
				fmt.Sprintf("Cannot delete: still referenced by OBCluster %s", cluster.Name))
			return fmt.Errorf("cannot delete LogService %s: still referenced by OBCluster %s", m.Resource.Name, cluster.Name)
		}
	}
	controllerutil.RemoveFinalizer(m.Resource, oceanbaseconst.FinalizerOBLogService)
	m.Resource.Status.Status = lsstatus.FinalizerFinished
	return m.Client.Update(m.Ctx, m.Resource)
}

func (m *OBLogServiceClusterManager) UpdateStatus() error {
	if m.Resource.DeletionTimestamp != nil {
		return m.Client.Status().Update(m.Ctx, m.Resource)
	}

	zoneList, err := m.listZones()
	if err != nil {
		m.Logger.Error(err, "list zones error")
	}

	if zoneList != nil {
		zoneStatusList := make([]apitypes.LogServiceZoneReplicaStatus, 0, len(zoneList.Items))
		for _, zone := range zoneList.Items {
			zoneStatusList = append(zoneStatusList, apitypes.LogServiceZoneReplicaStatus{
				Zone:   zone.Spec.Topology.Zone,
				Status: zone.Status.Status,
			})
		}
		m.Resource.Status.ZoneStatus = zoneStatusList
	}

	if m.Resource.Status.Status == lsstatus.Running {
		if zoneList != nil {
			for _, topo := range m.Resource.Spec.Topology {
				for _, zone := range zoneList.Items {
					if topo.Zone == zone.Spec.Topology.Zone && topo.Replica != zone.Spec.Topology.Replica {
						m.Resource.Status.Status = lsstatus.ModifyZoneReplica
						break
					}
				}
				if m.Resource.Status.Status != lsstatus.Running {
					break
				}
			}
		}
	}

	return m.Client.Status().Update(m.Ctx, m.Resource)
}

func (m *OBLogServiceClusterManager) ClearTaskInfo() {
	m.Resource.Status.Status = lsstatus.Running
	m.Resource.Status.OperationContext = nil
}

func (m *OBLogServiceClusterManager) FinishTask() {
	m.Resource.Status.Status = m.Resource.Status.OperationContext.TargetStatus
	m.Resource.Status.OperationContext = nil
}

func (m *OBLogServiceClusterManager) HandleFailure() {
	if m.Resource.DeletionTimestamp != nil {
		m.Resource.Status.OperationContext = nil
		return
	}
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

func (m *OBLogServiceClusterManager) GetTaskFunc(name tasktypes.TaskName) (tasktypes.TaskFunc, error) {
	return taskMap.GetTask(name, m)
}

func (m *OBLogServiceClusterManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.Resource, corev1.EventTypeWarning, "Task failed", err.Error())
}

func (m *OBLogServiceClusterManager) ArchiveResource() {
	m.Logger.Info("Archive log service cluster", "name", m.Resource.Name)
	m.Recorder.Event(m.Resource, "Archive", "", "Archive log service cluster")
	m.Resource.Status.Status = lsstatus.Failed
	m.Resource.Status.OperationContext = nil
}

func (m *OBLogServiceClusterManager) listZones() (*v1alpha1.OBLogServiceZoneList, error) {
	zoneList := &v1alpha1.OBLogServiceZoneList{}
	err := m.Client.List(m.Ctx, zoneList, client.MatchingLabels{
		oceanbaseconst.LabelRefOBLogServiceCluster: m.Resource.Name,
	}, client.InNamespace(m.Resource.Namespace))
	if err != nil {
		return nil, err
	}
	return zoneList, nil
}
