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

	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/task/status"
	"github.com/oceanbase/ob-operator/pkg/task/strategy"

	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	apipod "k8s.io/kubernetes/pkg/api/v1/pod"
	"sigs.k8s.io/controller-runtime/pkg/client"

	oceanbaseconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	clusterstatus "github.com/oceanbase/ob-operator/pkg/const/status/obcluster"
	serverstatus "github.com/oceanbase/ob-operator/pkg/const/status/observer"
	"github.com/oceanbase/ob-operator/pkg/task"
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
)

type OBServerManager struct {
	ResourceManager
	Ctx      context.Context
	OBServer *v1alpha1.OBServer
	Client   client.Client
	Recorder record.EventRecorder
	Logger   *logr.Logger
}

func (m *OBServerManager) GetTaskFunc(name string) (func() error, error) {
	switch name {
	case taskname.CreateOBPVC:
		return m.CreateOBPVC, nil
	case taskname.CreateOBPod:
		return m.CreateOBPod, nil
	case taskname.WaitOBServerReady:
		return m.WaitOBServerReady, nil
	case taskname.WaitOBClusterBootstrapped:
		return m.WaitOBClusterBootstrapped, nil
	case taskname.AddServer:
		return m.AddServer, nil
	case taskname.DeleteOBServerInCluster:
		return m.DeleteOBServerInCluster, nil
	case taskname.WaitOBServerDeletedInCluster:
		return m.WaitOBServerDeletedInCluster, nil
	case taskname.WaitOBServerPodReady:
		return m.WaitOBServerPodReady, nil
	case taskname.WaitOBServerActiveInCluster:
		return m.WaitOBServerActiveInCluster, nil
	case taskname.UpgradeOBServerImage:
		return m.UpgradeOBServerImage, nil
	case taskname.AnnotateOBServerPod:
		return m.AnnotateOBServerPod, nil
	default:
		return nil, errors.Errorf("Can not find an function for task %s", name)
	}
}

func (m *OBServerManager) IsNewResource() bool {
	return m.OBServer.Status.Status == ""
}

func (m *OBServerManager) InitStatus() {
	m.Logger.Info("newly created server, init status")
	status := v1alpha1.OBServerStatus{
		Image:  m.OBServer.Spec.OBServerTemplate.Image,
		Status: serverstatus.New,
	}
	m.OBServer.Status = status
}

func (m *OBServerManager) SetOperationContext(c *v1alpha1.OperationContext) {
	m.OBServer.Status.OperationContext = c
}

func (m *OBServerManager) SupportStaticIp() bool {
	switch m.OBServer.Status.CNI {
	case oceanbaseconst.CNICalico:
		return true
	default:
		return false
	}
}

func (m *OBServerManager) getCurrentOBServerFromOB() (*model.OBServer, error) {
	if m.OBServer.Status.PodIp == "" {
		err := errors.New("pod ip is empty")
		m.Logger.Error(err, "unable to get observer info")
		return nil, err
	}
	observerInfo := &model.ServerInfo{
		Ip:   m.OBServer.Status.PodIp,
		Port: oceanbaseconst.RpcPort,
	}
	operationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return nil, errors.Wrapf(err, "Get oceanbase operation manager failed")
	}
	return operationManager.GetServer(observerInfo)
}

func (m *OBServerManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		observer, err := m.getOBServer()
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		observer.Status = *m.OBServer.Status.DeepCopy()
		return m.Client.Status().Update(m.Ctx, observer)
	})
}

func (m *OBServerManager) UpdateStatus() error {
	// update deleting status when object is deleting
	if m.IsDeleting() {
		m.OBServer.Status.Status = serverstatus.Deleting
	} else {
		// get Pod status and update
		pod, err := m.getPod()
		if err != nil {
			if kubeerrors.IsNotFound(err) {
				m.Logger.Info("pod not found")
				if m.OBServer.Status.Status == serverstatus.Running {
					if m.SupportStaticIp() {
						m.Logger.Info("current cni supports specific static ip address, recover by recreate pod")
						m.OBServer.Status.Status = serverstatus.Recover
					} else {
						m.Logger.Info("observer not recoverable, delete current observer and wait recreate")
						m.OBServer.Status.Status = serverstatus.Unrecoverable
					}
				}
			} else {
				m.Logger.Info("observer status is not running, wait task finish")
				return errors.Wrap(err, "get pod when update status")
			}
		} else {
			m.OBServer.Status.Ready = apipod.IsPodReady(pod)
			m.OBServer.Status.PodPhase = pod.Status.Phase
			m.OBServer.Status.PodIp = pod.Status.PodIP
			m.OBServer.Status.NodeIp = pod.Status.HostIP
			// TODO update from obcluster
			m.OBServer.Status.CNI = GetCNIFromAnnotation(pod)
		}
		if m.OBServer.Status.Status == serverstatus.Running {
			m.Logger.Info("check observer in obcluster")
			observer, err := m.getCurrentOBServerFromOB()
			if err != nil {
				m.Logger.Info("Get observer failed, check next time")
			} else if observer == nil {
				m.OBServer.Status.Status = serverstatus.AddServer
			}
		}
		if m.OBServer.Status.Status == serverstatus.Running {
			if NeedAnnotation(pod, m.OBServer.Status.CNI) {
				m.OBServer.Status.Status = serverstatus.Annotate
			} else {
				for _, container := range pod.Spec.Containers {
					if container.Name == oceanbaseconst.ContainerName {
						m.OBServer.Status.Image = container.Image
						break
					}
				}
				if m.OBServer.Spec.OBServerTemplate.Image != m.OBServer.Status.Image {
					m.Logger.Info("Found image changed, begin upgrade")
					m.OBServer.Status.Status = serverstatus.Upgrade
				}
			}
		}

		m.Logger.Info("update observer status", "status", m.OBServer.Status)
		m.Logger.Info("update observer status", "operation context", m.OBServer.Status.OperationContext)
	}

	err := m.retryUpdateStatus()
	if err != nil {
		m.Logger.Error(err, "Got error when update observer status")
	}
	m.Logger.Info("update status finished")
	return err
}

func (m *OBServerManager) IsDeleting() bool {
	return !m.OBServer.ObjectMeta.DeletionTimestamp.IsZero()
}

func (m *OBServerManager) CheckAndUpdateFinalizers() error {
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
		finalizerFinished = m.OBServer.Status.Status == serverstatus.FinalizerFinished
	}
	if finalizerFinished {
		m.Logger.Info("Finalizer finished")
		m.OBServer.ObjectMeta.Finalizers = make([]string, 0)
		err := m.Client.Update(m.Ctx, m.OBServer)
		if err != nil {
			m.Logger.Error(err, "update observer instance failed")
			return errors.Wrapf(err, "Update observer %s in K8s failed", m.OBServer.Name)
		}
	}
	return nil
}

func (m *OBServerManager) GetTaskFlow() (*task.TaskFlow, error) {
	// exists unfinished task flow, return the last task flow
	if m.OBServer.Status.OperationContext != nil {
		m.Logger.Info("get task flow from observer status")
		return task.NewTaskFlow(m.OBServer.Status.OperationContext), nil
	}
	// newly created observer
	var taskFlow *task.TaskFlow
	var err error
	var obcluster *v1alpha1.OBCluster

	m.Logger.Info("create task flow according to observer status")
	switch m.OBServer.Status.Status {
	case serverstatus.New:
		obcluster, err = m.getOBCluster()
		if err != nil {
			return nil, errors.Wrap(err, "Get obcluster")
		}
		if obcluster.Status.Status == clusterstatus.New {
			// created when create obcluster
			m.Logger.Info("Create observer when create obcluster")
			taskFlow, err = task.GetRegistry().Get(flowname.PrepareOBServerForBootstrap)
		} else {
			// created normally
			m.Logger.Info("Create observer when obcluster already exists")
			taskFlow, err = task.GetRegistry().Get(flowname.CreateOBServer)
		}
		if err != nil {
			return nil, errors.Wrap(err, "Get create observer task flow")
		}
	case serverstatus.BootstrapReady:
		m.Logger.Info("Get task flow when bootstrap ready")
		taskFlow, err = task.GetRegistry().Get(flowname.MaintainOBServerAfterBootstrap)
	case serverstatus.Deleting:
		m.Logger.Info("Get task flow when observer deleting")
		taskFlow, err = task.GetRegistry().Get(flowname.DeleteOBServerFinalizer)
	case serverstatus.Upgrade:
		m.Logger.Info("Get task flow when observer upgrade")
		taskFlow, err = task.GetRegistry().Get(flowname.UpgradeOBServer)
	case serverstatus.Recover:
		m.Logger.Info("Get task flow when observer need recover")
		taskFlow, err = task.GetRegistry().Get(flowname.RecoverOBServer)
	case serverstatus.Annotate:
		m.Logger.Info("Get task flow when observer need set annotation")
		taskFlow, err = task.GetRegistry().Get(flowname.AnnotateOBServerPod)
	case serverstatus.AddServer:
		m.Logger.Info("Get task flow when observer need to be added to obcluster")
		taskFlow, err = task.GetRegistry().Get(flowname.AddServerInOB)
	default:
		m.Logger.Info("no need to run anything for observer")
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if taskFlow.OperationContext.OnFailure.Strategy == "" {
		taskFlow.OperationContext.OnFailure.Strategy = strategy.StartOver
		if taskFlow.OperationContext.OnFailure.NextTryStatus == "" {
			taskFlow.OperationContext.OnFailure.NextTryStatus = serverstatus.Running
		}
	}
	return taskFlow, nil
}

func (m *OBServerManager) ClearTaskInfo() {
	m.OBServer.Status.Status = serverstatus.Running
	m.OBServer.Status.OperationContext = nil
}

func (m *OBServerManager) FinishTask() {
	m.OBServer.Status.Status = m.OBServer.Status.OperationContext.TargetStatus
	m.OBServer.Status.OperationContext = nil
}

func (m *OBServerManager) HandleFailure() {
	if m.IsDeleting() {
		m.OBServer.Status.Status = serverstatus.Deleting
		m.OBServer.Status.OperationContext = nil
	} else {
		operationContext := m.OBServer.Status.OperationContext
		failureRule := operationContext.OnFailure
		switch failureRule.Strategy {
		case strategy.StartOver:
			m.OBServer.Status.Status = failureRule.NextTryStatus
			m.OBServer.Status.OperationContext = nil
		case strategy.RetryFromCurrent:
			operationContext.TaskStatus = taskstatus.Pending
		case strategy.Pause:
		}
	}
}

func (m *OBServerManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.OBServer, corev1.EventTypeWarning, "task exec failed", err.Error())
}

func (m *OBServerManager) generateNamespacedName(name string) types.NamespacedName {
	var namespacedName types.NamespacedName
	namespacedName.Namespace = m.OBServer.Namespace
	namespacedName.Name = name
	return namespacedName
}

func (m *OBServerManager) getPod() (*corev1.Pod, error) {
	// this label always exists
	pod := &corev1.Pod{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(m.OBServer.Name), pod)
	if err != nil {
		return nil, errors.Wrap(err, "get pod")
	}
	return pod, nil
}

func (m *OBServerManager) getOBCluster() (*v1alpha1.OBCluster, error) {
	// this label always exists
	clusterName, _ := m.OBServer.Labels[oceanbaseconst.LabelRefOBCluster]
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(clusterName), obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	return obcluster, nil
}

// get observer from K8s api server
func (m *OBServerManager) getOBServer() (*v1alpha1.OBServer, error) {
	// this label always exists
	observer := &v1alpha1.OBServer{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(m.OBServer.Name), observer)
	if err != nil {
		return nil, errors.Wrap(err, "get observer")
	}
	return observer, nil
}

func (m *OBServerManager) getOBZone() (*v1alpha1.OBZone, error) {
	// this label always exists
	zoneName, _ := m.OBServer.Labels[oceanbaseconst.LabelRefOBZone]
	obzone := &v1alpha1.OBZone{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(zoneName), obzone)
	if err != nil {
		return nil, errors.Wrap(err, "get obzone")
	}
	return obzone, nil
}
