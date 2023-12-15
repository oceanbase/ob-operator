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

package obcluster

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/obzone"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	opresource "github.com/oceanbase/ob-operator/pkg/coordinator"
	"github.com/oceanbase/ob-operator/pkg/task"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

type OBClusterManager struct {
	opresource.ResourceManager
	Ctx       context.Context
	OBCluster *v1alpha1.OBCluster
	Client    client.Client
	Recorder  telemetry.Recorder
	Logger    *logr.Logger
}

func (m *OBClusterManager) IsNewResource() bool {
	return m.OBCluster.Status.Status == ""
}

func (m *OBClusterManager) GetStatus() string {
	return m.OBCluster.Status.Status
}

func (m *OBClusterManager) InitStatus() {
	m.Logger.Info("newly created cluster, init status")
	m.Recorder.Event(m.OBCluster, "Init", "", "newly created cluster, init status")
	status := v1alpha1.OBClusterStatus{
		Image:        m.OBCluster.Spec.OBServerTemplate.Image,
		Status:       clusterstatus.New,
		OBZoneStatus: make([]apitypes.OBZoneReplicaStatus, 0, len(m.OBCluster.Spec.Topology)),
		UserSecrets:  m.OBCluster.Spec.UserSecrets,
	}
	if status.UserSecrets != nil {
		if status.UserSecrets.Monitor == "" {
			status.UserSecrets.Monitor = strings.Join([]string{m.OBCluster.Name, "monitor"}, "-")
		}
		if status.UserSecrets.ProxyRO == "" {
			status.UserSecrets.ProxyRO = strings.Join([]string{m.OBCluster.Name, "proxyro"}, "-")
		}
		if status.UserSecrets.Operator == "" {
			status.UserSecrets.Operator = strings.Join([]string{m.OBCluster.Name, "operator"}, "-")
		}
	}
	m.OBCluster.Status = status
}

func (m *OBClusterManager) SetOperationContext(c *tasktypes.OperationContext) {
	m.OBCluster.Status.OperationContext = c
}

func (m *OBClusterManager) GetTaskFlow() (*tasktypes.TaskFlow, error) {
	// exists unfinished task flow, return the last task flow
	if m.OBCluster.Status.OperationContext != nil {
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("get task flow from obcluster status")
		return tasktypes.NewTaskFlow(m.OBCluster.Status.OperationContext), nil
	}
	// return task flow depends on status

	// newly created cluster
	var taskFlow *tasktypes.TaskFlow
	var err error
	m.Logger.V(oceanbaseconst.LogLevelTrace).Info("create task flow according to obcluster status")
	switch m.OBCluster.Status.Status {
	// create obcluster, return taskFlow to bootstrap obcluster
	case clusterstatus.New:
		taskFlow, err = task.GetRegistry().Get(fBootstrapOBCluster)
	// after obcluster bootstraped, return taskFlow to maintain obcluster after bootstrap
	case clusterstatus.Bootstrapped:
		taskFlow, err = task.GetRegistry().Get(fMaintainOBClusterAfterBootstrap)
	case clusterstatus.AddOBZone:
		taskFlow, err = task.GetRegistry().Get(fAddOBZone)
	case clusterstatus.DeleteOBZone:
		taskFlow, err = task.GetRegistry().Get(fDeleteOBZone)
	case clusterstatus.ModifyOBZoneReplica:
		taskFlow, err = task.GetRegistry().Get(fModifyOBZoneReplica)
	case clusterstatus.Upgrade:
		taskFlow, err = task.GetRegistry().Get(fUpgradeOBCluster)
	case clusterstatus.ModifyOBParameter:
		taskFlow, err = task.GetRegistry().Get(fMaintainOBParameter)
	default:
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("no need to run anything for obcluster", "obcluster", m.OBCluster.Name)
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if taskFlow.OperationContext.OnFailure.Strategy == "" {
		taskFlow.OperationContext.OnFailure.Strategy = strategy.StartOver
		if taskFlow.OperationContext.OnFailure.NextTryStatus == "" {
			taskFlow.OperationContext.OnFailure.NextTryStatus = clusterstatus.Running
		}
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

func (m *OBClusterManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		obcluster, err := m.getOBCluster()
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		obcluster.Status = *m.OBCluster.Status.DeepCopy()
		return m.Client.Status().Update(m.Ctx, obcluster)
	})
}

func (m *OBClusterManager) UpdateStatus() error {
	if m.OBCluster.Status.Status == "Failed" {
		return nil
	}
	// update obzone status
	obzoneList, err := m.listOBZones()
	if err != nil {
		m.Logger.Error(err, "list obzones error")
		return errors.Wrap(err, "list obzones")
	}
	obzoneReplicaStatusList := make([]apitypes.OBZoneReplicaStatus, 0, len(obzoneList.Items))
	allZoneVersionSync := true
	for _, obzone := range obzoneList.Items {
		obzoneReplicaStatusList = append(obzoneReplicaStatusList, apitypes.OBZoneReplicaStatus{
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
	obparameterStatusList := make([]apitypes.Parameter, 0)
	for _, obparameter := range obparameterList.Items {
		obparameterStatusList = append(obparameterStatusList, *(obparameter.Spec.Parameter))
	}
	m.OBCluster.Status.Parameters = obparameterStatusList

	// compare spec and set status
	if m.OBCluster.Status.Status != clusterstatus.Running {
		m.Logger.V(oceanbaseconst.LogLevelDebug).Info("OBCluster status is not running, skip compare")
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
			parameterMap := make(map[string]apitypes.Parameter)
			for _, parameter := range m.OBCluster.Status.Parameters {
				m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Build parameter map", "parameter", parameter.Name)
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
	m.Logger.V(oceanbaseconst.LogLevelTrace).Info("update obcluster status", "status", m.OBCluster.Status)
	err = m.retryUpdateStatus()
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

func (m *OBClusterManager) HandleFailure() {
	operationContext := m.OBCluster.Status.OperationContext
	failureRule := operationContext.OnFailure
	switch failureRule.Strategy {
	case strategy.StartOver:
		if m.OBCluster.Status.Status != failureRule.NextTryStatus {
			m.OBCluster.Status.Status = failureRule.NextTryStatus
			m.OBCluster.Status.OperationContext = nil
		} else {
			m.OBCluster.Status.OperationContext.Idx = 0
			m.OBCluster.Status.OperationContext.TaskStatus = ""
			m.OBCluster.Status.OperationContext.TaskId = ""
			m.OBCluster.Status.OperationContext.Task = ""
		}
	case strategy.RetryFromCurrent:
		operationContext.TaskStatus = taskstatus.Pending
	case strategy.Pause:
	}
}

func (m *OBClusterManager) GetTaskFunc(name tasktypes.TaskName) (tasktypes.TaskFunc, error) {
	switch name {
	case tCheckAndCreateUserSecrets:
		return m.CheckAndCreateUserSecrets, nil
	case tCreateOBZone:
		return m.CreateOBZone, nil
	case tDeleteOBZone:
		return m.DeleteOBZone, nil
	case tModifyOBZoneReplica:
		return m.ModifyOBZoneReplica, nil
	case tWaitOBZoneTopologyMatch:
		return m.WaitOBZoneTopologyMatch, nil
	case tWaitOBZoneBootstrapReady:
		return m.generateWaitOBZoneStatusFunc(zonestatus.BootstrapReady, oceanbaseconst.DefaultStateWaitTimeout), nil
	case tWaitOBZoneRunning:
		return m.generateWaitOBZoneStatusFunc(zonestatus.Running, oceanbaseconst.DefaultStateWaitTimeout), nil
	case tWaitOBZoneDeleted:
		return m.WaitOBZoneDeleted, nil
	case tBootstrap:
		return m.Bootstrap, nil
	case tCreateUsers:
		return m.CreateUsers, nil
	case tCreateOBClusterService:
		return m.CreateService, nil
	case tMaintainOBParameter:
		return m.MaintainOBParameter, nil
	case tValidateUpgradeInfo:
		return m.ValidateUpgradeInfo, nil
	case tUpgradeCheck:
		return m.UpgradeCheck, nil
	case tBackupEssentialParameters:
		return m.BackupEssentialParameters, nil
	case tBeginUpgrade:
		return m.BeginUpgrade, nil
	case tRollingUpgradeByZone:
		return m.RollingUpgradeByZone, nil
	case tFinishUpgrade:
		return m.FinishUpgrade, nil
	case tRestoreEssentialParameters:
		return m.RestoreEssentialParameters, nil
	case tCreateServiceForMonitor:
		return m.CreateServiceForMonitor, nil
	case tModifySysTenantReplica:
		return m.ModifySysTenantReplica, nil
	default:
		return nil, errors.New("Can not find a function for task")
	}
}

func (m *OBClusterManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.OBCluster, corev1.EventTypeWarning, "task exec failed", err.Error())
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

func (m *OBClusterManager) ArchiveResource() {
	m.Logger.Info("Archive obcluster", "obcluster", m.OBCluster.Name)
	m.Recorder.Event(m.OBCluster, "Archive", "", "archive obcluster")
	m.OBCluster.Status.Status = "Failed"
	m.OBCluster.Status.OperationContext = nil
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
