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
	m.Logger.Info("Set operation context", "current", m.OBCluster.Status.OperationContext, "new", c)
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
		taskFlow, err = task.GetRegistry().Get(flowname.UpgradeOBCluster)
	case clusterstatus.ModifyOBParameter:
		taskFlow, err = task.GetRegistry().Get(flowname.MaintainOBParameter)
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
	// update obzone status
	obzoneList, err := m.listOBZones()
	if err != nil {
		m.Logger.Error(err, "list obzones error")
		return errors.Wrap(err, "list obzones")
	}
	obzoneReplicaStatusList := make([]v1alpha1.OBZoneReplicaStatus, 0, len(obzoneList.Items))
	allZoneVersionSync := true
	for _, obzone := range obzoneList.Items {
		obzoneReplicaStatusList = append(obzoneReplicaStatusList, v1alpha1.OBZoneReplicaStatus{
			Zone:   obzone.Name,
			Status: obzone.Status.Status,
		})
		if obzone.Status.Image != m.OBCluster.Spec.OBServerTemplate.Image {
			m.Logger.Info("obzone still not sync")
			allZoneVersionSync = false
		}
	}
	m.OBCluster.Status.OBZoneStatus = obzoneReplicaStatusList

	// update parameter, only need to record and compare parameter spec
	obparameterList, err := m.listOBParameters()
	if err != nil {
		m.Logger.Error(err, "list obparameters error")
		return errors.Wrap(err, "list obparameters")
	}
	obparameterStatusList := make([]v1alpha1.Parameter, 0)
	for _, obparameter := range obparameterList.Items {
		obparameterStatusList = append(obparameterStatusList, *(obparameter.Spec.Parameter))
	}
	m.OBCluster.Status.Parameters = obparameterStatusList

	// compare spec and set status
	if m.OBCluster.Status.Status != clusterstatus.Running {
		m.Logger.Info("OBCluster status is not running, skip compare")
	} else {
		if allZoneVersionSync {
			m.OBCluster.Status.Image = m.OBCluster.Spec.OBServerTemplate.Image
		}
		// TODO: refactor this part of code
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

		// check for upgrade
		if m.OBCluster.Status.Status == clusterstatus.Running {
			if m.OBCluster.Spec.OBServerTemplate.Image != m.OBCluster.Status.Image {
				m.Logger.Info("Check obcluster image not match, need upgrade")
				m.OBCluster.Status.Status = clusterstatus.Upgrade
			}
		}

		// only do this when obzone matched, thus obcluster status is running after obzone check
		if m.OBCluster.Status.Status == clusterstatus.Running {
			parameterMap := make(map[string]v1alpha1.Parameter)
			for _, parameter := range m.OBCluster.Status.Parameters {
				m.Logger.Info("Build parameter map", "parameter", parameter.Name)
				parameterMap[parameter.Name] = parameter
			}
			for _, parameter := range m.OBCluster.Spec.Parameters {
				parameterStatus, parameterExists := parameterMap[parameter.Name]
				// need create or update parameter
				if !parameterExists || parameterStatus.Value != parameter.Value {
					m.OBCluster.Status.Status = clusterstatus.ModifyOBParameter
					break
				}
				delete(parameterMap, parameter.Name)
			}

			// need delete parameter
			if len(parameterMap) > 0 {
				m.OBCluster.Status.Status = clusterstatus.ModifyOBParameter
			}
		}
	}
	m.Logger.Info("update obcluster status", "status", m.OBCluster.Status)
	m.Logger.Info("update obcluster status", "operation context", m.OBCluster.Status.OperationContext)
	err = m.Client.Status().Update(m.Ctx, m.OBCluster.DeepCopy())
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
	case taskname.WaitOBZoneDeleted:
		return m.WaitOBZoneDeleted, nil
	case taskname.Bootstrap:
		return m.Bootstrap, nil
	case taskname.CreateUsers:
		return m.CreateUsers, nil
	case taskname.CreateOBClusterService:
		return m.CreateService, nil
	case taskname.MaintainOBParameter:
		return m.MaintainOBParameter, nil
	case taskname.ValidateUpgradeInfo:
		return m.ValidateUpgradeInfo, nil
	case taskname.UpgradeCheck:
		return m.UpgradeCheck, nil
	case taskname.BackupEssentialParameters:
		return m.BackupEssentialParameters, nil
	case taskname.BeginUpgrade:
		return m.BeginUpgrade, nil
	case taskname.RollingUpgradeByZone:
		return m.RollingUpgradeByZone, nil
	case taskname.FinishUpgrade:
		return m.FinishUpgrade, nil
	case taskname.RestoreEssentialParameters:
		return m.RestoreEssentialParameters, nil
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

func (m *OBClusterManager) listOBParameters() (*v1alpha1.OBParameterList, error) {
	// this label always exists
	obparameterList := &v1alpha1.OBParameterList{}
	err := m.Client.List(m.Ctx, obparameterList, client.MatchingLabels{
		oceanbaseconst.LabelRefOBCluster: m.OBCluster.Name,
	}, client.InNamespace(m.OBCluster.Namespace))
	if err != nil {
		return nil, errors.Wrap(err, "get obzone list")
	}
	return obparameterList, nil
}
