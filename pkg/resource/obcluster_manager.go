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
	oceanbaseconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	clusterstatus "github.com/oceanbase/ob-operator/pkg/const/status/obcluster"
	zonestatus "github.com/oceanbase/ob-operator/pkg/const/status/obzone"
	"github.com/oceanbase/ob-operator/pkg/task"
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
)

type OBClusterManager struct {
	ResourceManager
	Ctx       context.Context
	OBCluster *v1alpha1.OBCluster
	Client    client.Client
	Recorder  record.EventRecorder
	Logger    *logr.Logger
}

func (m *OBClusterManager) IsNewResource() bool {
	return m.OBCluster.Status.Status == ""
}

func (m *OBClusterManager) InitStatus() {
	m.Logger.Info("newly created cluster, init status")
	status := v1alpha1.OBClusterStatus{
		Image:        m.OBCluster.Spec.OBServerTemplate.Image,
		Status:       clusterstatus.New,
		OBZoneStatus: make([]v1alpha1.OBZoneReplicaStatus, 0, len(m.OBCluster.Spec.Topology)),
	}
	m.OBCluster.Status = status
}

func (m *OBClusterManager) SetOperationContext(c *v1alpha1.OperationContext) {
	m.OBCluster.Status.OperationContext = c
}

func (m *OBClusterManager) GetTaskFlow() (*task.TaskFlow, error) {
	// exists unfinished task flow, return the last task flow
	if m.OBCluster.Status.OperationContext != nil {
		m.Logger.Info("get task flow from obcluster status")
		return task.NewTaskFlow(m.OBCluster.Status.OperationContext), nil
	}
	// return task flow depends on status

	// newly created cluster
	var taskFlow *task.TaskFlow
	var err error
	m.Logger.Info("create task flow according to obcluster status")
	switch m.OBCluster.Status.Status {
	// create obcluster, return taskflow to bootstrap obcluster
	case clusterstatus.New:
		taskFlow, err = task.GetRegistry().Get(flowname.BootstrapOBCluster)
	// after obcluster bootstraped, return taskflow to maintain obcluster after bootstrap
	case clusterstatus.Bootstrapped:
		taskFlow, err = task.GetRegistry().Get(flowname.MaintainOBClusterAfterBootstrap)
	case clusterstatus.AddOBZone:
		taskFlow, err = task.GetRegistry().Get(flowname.AddOBZone)
	case clusterstatus.DeleteOBZone:
		taskFlow, err = task.GetRegistry().Get(flowname.DeleteOBZone)
	case clusterstatus.ModifyOBZoneReplica:
		taskFlow, err = task.GetRegistry().Get(flowname.ModifyOBZoneReplica)
	case clusterstatus.Upgrade:
		taskFlow, err = task.GetRegistry().Get(flowname.Upgrade)
	default:
		m.Logger.Info("no need to run anything for obcluster", "obcluster", m.OBCluster.Name)
	}
	return taskFlow, err
}

func (m *OBClusterManager) IsDeleting() bool {
	return !m.OBCluster.ObjectMeta.DeletionTimestamp.IsZero()
}

func (m *OBClusterManager) CheckAndUpdateFinalizers() error {
	if m.OBCluster.Status.Status == clusterstatus.FinalizerFinished {
		m.OBCluster.ObjectMeta.Finalizers = make([]string, 0)
		return m.Client.Update(m.Ctx, m.OBCluster)
	}
	return nil
}

func (m *OBClusterManager) UpdateStatus() error {
	obzoneList, err := m.listOBZones()
	if err != nil {
		m.Logger.Error(err, "list obzones error")
		return errors.Wrap(err, "list obzones")
	}
	obzoneReplicaStatusList := make([]v1alpha1.OBZoneReplicaStatus, 0, len(obzoneList.Items))
	for _, obzone := range obzoneList.Items {
		obzoneReplicaStatusList = append(obzoneReplicaStatusList, v1alpha1.OBZoneReplicaStatus{
			Zone:   obzone.Name,
			Status: obzone.Status.Status,
		})
	}
	m.OBCluster.Status.OBZoneStatus = obzoneReplicaStatusList
	// compare spec and set status
	if m.OBCluster.Status.Status != clusterstatus.Running {
		m.Logger.Info("OBCluster status is not running, skip compare")
	} else {
		// check topology
		if len(m.OBCluster.Spec.Topology) > len(obzoneList.Items) {
			m.Logger.Info("Compare topology need add zone")
			m.OBCluster.Status.Status = clusterstatus.AddOBZone
		} else if len(m.OBCluster.Spec.Topology) < len(obzoneList.Items) {
			m.Logger.Info("Compare topology need delete zone")
			m.OBCluster.Status.Status = clusterstatus.DeleteOBZone
		} else {
			observerMatch := true
			for _, zone := range m.OBCluster.Spec.Topology {
				if !observerMatch {
					break
				}
				for _, obzone := range obzoneList.Items {
					if zone.Zone == obzone.Spec.Topology.Zone {
						if zone.Replica != len(obzone.Status.OBServerStatus) {
							m.OBCluster.Status.Status = clusterstatus.ModifyOBZoneReplica
							observerMatch = false
						}
						break
					}
				}
			}
		}
		// TODO resource change require pod restart, and since oceanbase is a distributed system, resource can be scaled by add more servers
	}
	m.Logger.Info("update obcluster status", "status", m.OBCluster.Status)
	m.Logger.Info("update obcluster status", "operation context", m.OBCluster.Status.OperationContext)
	err = m.Client.Status().Update(m.Ctx, m.OBCluster)
	if err != nil {
		m.Logger.Error(err, "Got error when update obcluster status")
	}
	return err
}

func (m *OBClusterManager) ClearTaskInfo() {
	m.OBCluster.Status.Status = clusterstatus.Running
	m.OBCluster.Status.OperationContext = nil
}

func (m *OBClusterManager) FinishTask() {
	m.OBCluster.Status.Status = m.OBCluster.Status.OperationContext.TargetStatus
	m.OBCluster.Status.OperationContext = nil
}

func (m *OBClusterManager) GetTaskFunc(name string) (func() error, error) {
	switch name {
	case taskname.CreateOBZone:
		return m.CreateOBZone, nil
	case taskname.DeleteOBZone:
		return m.DeleteOBZone, nil
	case taskname.ModifyOBZoneReplica:
		return m.ModifyOBZoneReplica, nil
	case taskname.WaitOBZoneTopologyMatch:
		return m.WaitOBZoneTopologyMatch, nil
	case taskname.WaitOBZoneBootstrapReady:
		return m.generateWaitOBZoneStatusFunc(zonestatus.BootstrapReady, oceanbaseconst.DefaultStateWaitTimeout), nil
	case taskname.WaitOBZoneRunning:
		return m.generateWaitOBZoneStatusFunc(zonestatus.Running, oceanbaseconst.DefaultStateWaitTimeout), nil
	case taskname.Bootstrap:
		return m.Bootstrap, nil
	case taskname.CreateUsers:
		return m.CreateUsers, nil
	case taskname.CreateOBClusterService:
		return m.CreateService, nil
	case taskname.CreateOBParameter:
		return m.CreateOBParameter, nil
	default:
		return nil, errors.New("Can not find a function for task")
	}
}

func (m *OBClusterManager) listOBZones() (*v1alpha1.OBZoneList, error) {
	// this label always exists
	obzoneList := &v1alpha1.OBZoneList{}
	err := m.Client.List(m.Ctx, obzoneList, client.MatchingLabels{
		oceanbaseconst.LabelRefOBCluster: m.OBCluster.Name,
	}, client.InNamespace(m.OBCluster.Namespace))
	if err != nil {
		return nil, errors.Wrap(err, "get obzone list")
	}
	return obzoneList, nil
}
