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

package obzone

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"

	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/obzone"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	opresource "github.com/oceanbase/ob-operator/pkg/coordinator"
	"github.com/oceanbase/ob-operator/pkg/task"
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/task/status"
	"github.com/oceanbase/ob-operator/pkg/task/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

type OBZoneManager struct {
	opresource.ResourceManager
	Ctx      context.Context
	OBZone   *v1alpha1.OBZone
	Client   client.Client
	Recorder telemetry.Recorder
	Logger   *logr.Logger
}

func (m *OBZoneManager) IsNewResource() bool {
	return m.OBZone.Status.Status == ""
}

func (m *OBZoneManager) GetStatus() string {
	return m.OBZone.Status.Status
}

func (m *OBZoneManager) InitStatus() {
	m.Logger.Info("newly created zone, init status")
	status := v1alpha1.OBZoneStatus{
		Image:          m.OBZone.Spec.OBServerTemplate.Image,
		Status:         zonestatus.New,
		OBServerStatus: make([]apitypes.OBServerReplicaStatus, 0, m.OBZone.Spec.Topology.Replica),
	}
	m.OBZone.Status = status
}

func (m *OBZoneManager) SetOperationContext(c *tasktypes.OperationContext) {
	m.OBZone.Status.OperationContext = c
}

func (m *OBZoneManager) GetTaskFlow() (*task.TaskFlow, error) {
	// exists unfinished task flow, return the last task flow
	if m.OBZone.Status.OperationContext != nil {
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("get task flow from obzone status")
		return task.NewTaskFlow(m.OBZone.Status.OperationContext), nil
	}
	// newly created zone
	var taskFlow *task.TaskFlow
	var err error
	var obcluster *v1alpha1.OBCluster

	switch m.OBZone.Status.Status {
	case zonestatus.New:
		obcluster, err = m.getOBCluster()
		if err != nil {
			return nil, errors.Wrap(err, "Get obcluster")
		}
		if obcluster.Status.Status == clusterstatus.New {
			// created when create obcluster
			m.Logger.Info("Create obzone when create obcluster")
			taskFlow, err = task.GetRegistry().Get(flowname.PrepareOBZoneForBootstrap)
		} else {
			// created normally
			m.Logger.Info("Create obzone when obcluster already exists")
			taskFlow, err = task.GetRegistry().Get(flowname.CreateOBZone)
		}
		if err != nil {
			return nil, errors.Wrap(err, "Get create obzone task flow")
		}
	case zonestatus.BootstrapReady:
		taskFlow, err = task.GetRegistry().Get(flowname.MaintainOBZoneAfterBootstrap)
	case zonestatus.AddOBServer:
		taskFlow, err = task.GetRegistry().Get(flowname.AddOBServer)
	case zonestatus.DeleteOBServer:
		taskFlow, err = task.GetRegistry().Get(flowname.DeleteOBServer)
	case zonestatus.Deleting:
		taskFlow, err = task.GetRegistry().Get(flowname.DeleteOBZoneFinalizer)
	case zonestatus.Upgrade:
		obcluster, err = m.getOBCluster()
		if err != nil {
			return nil, errors.Wrap(err, "Get obcluster")
		}
		if len(obcluster.Status.OBZoneStatus) >= 3 {
			return task.GetRegistry().Get(flowname.UpgradeOBZone)
		}
		return task.GetRegistry().Get(flowname.ForceUpgradeOBZone)
		// TODO upgrade
	default:
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("no need to run anything for obzone")
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if taskFlow.OperationContext.OnFailure.Strategy == "" {
		taskFlow.OperationContext.OnFailure.Strategy = strategy.StartOver
		if taskFlow.OperationContext.OnFailure.NextTryStatus == "" {
			taskFlow.OperationContext.OnFailure.NextTryStatus = zonestatus.Running
		}
	}
	return taskFlow, nil
}

func (m *OBZoneManager) IsDeleting() bool {
	return !m.OBZone.ObjectMeta.DeletionTimestamp.IsZero()
}

func (m *OBZoneManager) CheckAndUpdateFinalizers() error {
	finalizerFinished := false
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
	} else {
		finalizerFinished = m.OBZone.Status.Status == zonestatus.FinalizerFinished
	}
	if finalizerFinished {
		m.Logger.Info("Finalizer finished")
		m.OBZone.ObjectMeta.Finalizers = make([]string, 0)
		err := m.Client.Update(m.Ctx, m.OBZone)
		if err != nil {
			m.Logger.Error(err, "update obzone instance failed")
			return errors.Wrapf(err, "Update obzone %s in K8s failed", m.OBZone.Name)
		}
	}
	return nil
}

func (m *OBZoneManager) ArchiveResource() {
	m.Logger.Info("Archive obzone", "obzone", m.OBZone.Name)
	m.Recorder.Event(m.OBZone, "Archive", "", "archive obzone")
	m.OBZone.Status.Status = "Failed"
	m.OBZone.Status.OperationContext = nil
}

func (m *OBZoneManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		obzone, err := m.getOBZone()
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		obzone.Status = *m.OBZone.Status.DeepCopy()
		return m.Client.Status().Update(m.Ctx, obzone)
	})
}

func (m *OBZoneManager) UpdateStatus() error {
	if m.OBZone.Status.Status == "Failed" {
		return nil
	}
	observerList, err := m.listOBServers()
	if err != nil {
		m.Logger.Error(err, "Got error when list observers")
	}
	observerReplicaStatusList := make([]apitypes.OBServerReplicaStatus, 0, len(observerList.Items))
	availableReplica := 0
	// handle upgrade
	allServerVersionSync := true
	for _, observer := range observerList.Items {
		observerReplicaStatusList = append(observerReplicaStatusList, apitypes.OBServerReplicaStatus{
			Server: observer.Status.PodIp,
			Status: observer.Status.Status,
		})
		if observer.Status.Status != serverstatus.Unrecoverable {
			availableReplica++
		}
		if observer.Status.Image != m.OBZone.Spec.OBServerTemplate.Image {
			m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Found observer image not match")
			allServerVersionSync = false
		}
	}
	m.OBZone.Status.OBServerStatus = observerReplicaStatusList
	if m.IsDeleting() {
		m.OBZone.Status.Status = zonestatus.Deleting
	}
	if m.OBZone.Status.Status != zonestatus.Running {
		m.Logger.V(oceanbaseconst.LogLevelDebug).Info("OBZone status is not running, skip compare")
	} else {
		// set status image
		if allServerVersionSync {
			m.OBZone.Status.Image = m.OBZone.Spec.OBServerTemplate.Image
		}
		// check topology
		if m.OBZone.Spec.Topology.Replica > availableReplica {
			m.Logger.Info("Compare topology need add observer")
			m.OBZone.Status.Status = zonestatus.AddOBServer
		} else if m.OBZone.Spec.Topology.Replica < len(m.OBZone.Status.OBServerStatus) {
			m.Logger.Info("Compare topology need delete observer")
			m.OBZone.Status.Status = zonestatus.DeleteOBServer
		}
		// do nothing when observer match topology replica

		// TODO resource change require pod restart, and since oceanbase is a distributed system, resource can be scaled by add more servers
		if m.OBZone.Status.Status == zonestatus.Running {
			if m.OBZone.Status.Image != m.OBZone.Spec.OBServerTemplate.Image {
				m.Logger.Info("Found image changed, need upgrade")
				m.OBZone.Status.Status = zonestatus.Upgrade
			}
		}
	}
	m.Logger.V(oceanbaseconst.LogLevelTrace).Info("update obzone status", "status", m.OBZone.Status)
	err = m.retryUpdateStatus()
	if err != nil {
		m.Logger.Error(err, "Got error when update obzone status")
	}
	return err
}

func (m *OBZoneManager) ClearTaskInfo() {
	m.OBZone.Status.Status = zonestatus.Running
	m.OBZone.Status.OperationContext = nil
}

func (m *OBZoneManager) HandleFailure() {
	if m.IsDeleting() {
		m.OBZone.Status.Status = zonestatus.Deleting
		m.OBZone.Status.OperationContext = nil
	} else {
		operationContext := m.OBZone.Status.OperationContext
		failureRule := operationContext.OnFailure
		switch failureRule.Strategy {
		case strategy.StartOver:
			m.OBZone.Status.Status = failureRule.NextTryStatus
			m.OBZone.Status.OperationContext.Idx = 0
			m.OBZone.Status.OperationContext.TaskStatus = ""
			m.OBZone.Status.OperationContext.TaskId = ""
			m.OBZone.Status.OperationContext.Task = ""
		case strategy.RetryFromCurrent:
			operationContext.TaskStatus = taskstatus.Pending
		case strategy.Pause:
		}
	}
}

func (m *OBZoneManager) FinishTask() {
	m.OBZone.Status.Status = m.OBZone.Status.OperationContext.TargetStatus
	m.OBZone.Status.OperationContext = nil
}

func (m *OBZoneManager) GetTaskFunc(name string) (tasktypes.TaskFunc, error) {
	switch name {
	case taskname.CreateOBServer:
		return m.CreateOBServer, nil
	case taskname.WaitOBServerBootstrapReady:
		return m.generateWaitOBServerStatusFunc(serverstatus.BootstrapReady, oceanbaseconst.DefaultStateWaitTimeout), nil
	case taskname.WaitOBServerRunning:
		return m.generateWaitOBServerStatusFunc(serverstatus.Running, oceanbaseconst.DefaultStateWaitTimeout), nil
	case taskname.AddZone:
		return m.AddZone, nil
	case taskname.StartOBZone:
		return m.StartOBZone, nil
	case taskname.DeleteOBServer:
		return m.DeleteOBServer, nil
	case taskname.DeleteAllOBServer:
		return m.DeleteAllOBServer, nil
	case taskname.WaitReplicaMatch:
		return m.WaitReplicaMatch, nil
	case taskname.WaitOBServerDeleted:
		return m.WaitOBServerDeleted, nil
	case taskname.StopOBZone:
		return m.StopOBZone, nil
	case taskname.DeleteOBZoneInCluster:
		return m.DeleteOBZoneInCluster, nil
	case taskname.OBClusterHealthCheck:
		return m.OBClusterHealthCheck, nil
	case taskname.OBZoneHealthCheck:
		return m.OBZoneHealthCheck, nil
	case taskname.UpgradeOBServer:
		return m.UpgradeOBServer, nil
	case taskname.WaitOBServerUpgraded:
		return m.WaitOBServerUpgraded, nil
	default:
		return nil, errors.Errorf("Can not find an function for %s", name)
	}
}

func (m *OBZoneManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.OBZone, corev1.EventTypeWarning, "task exec failed", err.Error())
}

func (m *OBZoneManager) listOBServers() (*v1alpha1.OBServerList, error) {
	// this label always exists
	observerList := &v1alpha1.OBServerList{}
	err := m.Client.List(m.Ctx, observerList, client.MatchingLabels{
		oceanbaseconst.LabelRefOBZone: m.OBZone.Name,
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
	clusterName, _ := m.OBZone.Labels[oceanbaseconst.LabelRefOBCluster]
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(clusterName), obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	return obcluster, nil
}
