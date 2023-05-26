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
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	clusterstatus "github.com/oceanbase/ob-operator/pkg/const/status/obcluster"
	zonestatus "github.com/oceanbase/ob-operator/pkg/const/status/obzone"
	"github.com/oceanbase/ob-operator/pkg/task"
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
)

type OBZoneManager struct {
	ResourceManager
	Ctx      context.Context
	OBZone   *v1alpha1.OBZone
	Client   client.Client
	Recorder record.EventRecorder
	Logger   *logr.Logger
}

func (m *OBZoneManager) IsNewResource() bool {
	return m.OBZone.Status.Status == ""
}

func (m *OBZoneManager) InitStatus() {
	m.Logger.Info("newly created zone, init status")
	status := v1alpha1.OBZoneStatus{
		Image:          m.OBZone.Spec.OBServerTemplate.Image,
		Status:         zonestatus.New,
		OBServerStatus: make([]v1alpha1.OBServerReplicaStatus, 0, m.OBZone.Spec.Topology.Replica),
	}
	m.OBZone.Status = status
}

func (m *OBZoneManager) SetOperationContext(c *v1alpha1.OperationContext) {
	m.OBZone.Status.OperationContext = c
}

func (m *OBZoneManager) GetTaskFlow() (*task.TaskFlow, error) {
	// exists unfinished task flow, return the last task flow
	if m.OBZone.Status.OperationContext != nil {
		m.Logger.Info("get task flow from obzone status")
		return task.NewTaskFlow(m.OBZone.Status.OperationContext), nil
	}
	// newly created zone
	var taskFlow *task.TaskFlow
	var err error
	var obcluster *v1alpha1.OBCluster

	m.Logger.Info("create task flow according to obzone status")
	if m.OBZone.Status.Status == zonestatus.New {
		obcluster, err = m.getOBCluster()
		if err != nil {
			return nil, errors.Wrap(err, "Get obcluster")
		}
		if obcluster.Status.Status == clusterstatus.New {
			// created when create obcluster
			m.Logger.Info("Create obzone when create obcluster")
			taskFlow, err = task.GetRegistry().Get(flowname.CreateZoneForBootstrap)
		} else {
			// created normally
			m.Logger.Info("Create obzone when obcluster already exists")
			taskFlow, err = task.GetRegistry().Get(flowname.CreateZone)
		}
		if err != nil {
			return nil, errors.Wrap(err, "Get create obzone task flow")
		}
		return taskFlow, nil
	}
	// scale observer
	// upgrade

	// no need to execute task flow
	return nil, nil
}

func (m *OBZoneManager) UpdateStatus() error {
	observerList, err := m.listOBServers()
	if err != nil {
		m.Logger.Error(err, "Got error when list observers")
	}

	observerReplicaStatusList := make([]v1alpha1.OBServerReplicaStatus, 0, len(observerList.Items))
	for _, observer := range observerList.Items {
		observerReplicaStatusList = append(observerReplicaStatusList, v1alpha1.OBServerReplicaStatus{
			Server: observer.Status.PodIp,
			Status: observer.Status.Status,
		})
	}
	m.OBZone.Status.OBServerStatus = observerReplicaStatusList
	m.Logger.Info("update obzone status", "status", m.OBZone.Status)
	m.Logger.Info("update obzone status", "operation context", m.OBZone.Status.OperationContext)
	err = m.Client.Status().Update(m.Ctx, m.OBZone)
	if err != nil {
		m.Logger.Error(err, "Got error when update obzone status")
	}
	return err
}

func (m *OBZoneManager) ClearTaskInfo() {
	m.OBZone.Status.Status = zonestatus.Running
	m.OBZone.Status.OperationContext = nil
}

func (m *OBZoneManager) FinishTask() {
	m.OBZone.Status.Status = m.OBZone.Status.OperationContext.TargetStatus
	m.OBZone.Status.OperationContext = nil
}

func (m *OBZoneManager) GetTaskFunc(name string) (func() error, error) {
	switch name {
	case taskname.CreateOBServer:
		return m.CreateOBServer, nil
	case taskname.WaitOBServerBootstrapReady:
		return m.WaitOBServerBootstrapReady, nil
	case taskname.AddZone:
		return m.AddZone, nil
	default:
		return nil, errors.New(fmt.Sprintf("Can not find an function for %s", name))
	}
}

func (m *OBZoneManager) listOBServers() (*v1alpha1.OBServerList, error) {
	// this label always exists
	observerList := &v1alpha1.OBServerList{}
	err := m.Client.List(m.Ctx, observerList, client.MatchingLabels{
		"reference-zone": m.OBZone.Name,
	}, client.InNamespace(m.OBZone.Namespace))
	if err != nil {
		return nil, errors.Wrap(err, "get observers")
	}
	return observerList, err
}

func (m *OBZoneManager) generateNamespacedName(name string) types.NamespacedName {
	var namespacedName types.NamespacedName
	namespacedName.Namespace = m.OBZone.Namespace
	namespacedName.Name = name
	return namespacedName
}

func (m *OBZoneManager) getOBZone() (*v1alpha1.OBZone, error) {
	// this label always exists
	obzone := &v1alpha1.OBZone{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(m.OBZone.Name), obzone)
	if err != nil {
		return nil, errors.Wrap(err, "get obzone")
	}
	return obzone, nil
}

func (m *OBZoneManager) getOBCluster() (*v1alpha1.OBCluster, error) {
	// this label always exists
	clusterName, _ := m.OBZone.Labels["reference-cluster"]
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(clusterName), obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	return obcluster, nil
}
