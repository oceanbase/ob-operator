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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/task/status"
	"github.com/oceanbase/ob-operator/pkg/task/strategy"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"
	clusterstatus "github.com/oceanbase/ob-operator/pkg/const/status/obcluster"
	parameterstatus "github.com/oceanbase/ob-operator/pkg/const/status/obparameter"
	"github.com/oceanbase/ob-operator/pkg/task"
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
)

type OBParameterManager struct {
	ResourceManager
	Ctx         context.Context
	OBParameter *v1alpha1.OBParameter
	Client      client.Client
	Recorder    record.EventRecorder
	Logger      *logr.Logger
}

func (m *OBParameterManager) IsNewResource() bool {
	return m.OBParameter.Status.Status == ""
}

func (m *OBParameterManager) InitStatus() {
	m.Logger.Info("newly created obparameter, init status")
	status := v1alpha1.OBParameterStatus{
		Status:    parameterstatus.New,
		Parameter: make([]v1alpha1.ParameterValue, 0),
	}
	m.OBParameter.Status = status
}

func (m *OBParameterManager) SetOperationContext(c *v1alpha1.OperationContext) {
	m.OBParameter.Status.OperationContext = c
}

func (m *OBParameterManager) GetTaskFlow() (*task.TaskFlow, error) {
	// exists unfinished task flow, return the last task flow
	if m.OBParameter.Status.OperationContext != nil {
		m.Logger.Info("get task flow from obparameter status")
		return task.NewTaskFlow(m.OBParameter.Status.OperationContext), nil
	}

	// return task flow depends on status

	var taskFlow *task.TaskFlow
	var err error
	m.Logger.Info("create task flow according to obparameter status")
	switch m.OBParameter.Status.Status {
	// only need to handle parameter not match
	case parameterstatus.NotMatch:
		taskFlow, err = task.GetRegistry().Get(flowname.SetOBParameter)
	default:
		m.Logger.Info("no need to run anything for obparameter")
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if taskFlow.OperationContext.OnFailure.Strategy == "" {
		taskFlow.OperationContext.OnFailure.Strategy = strategy.StartOver
		if taskFlow.OperationContext.OnFailure.NextTryStatus == "" {
			taskFlow.OperationContext.OnFailure.NextTryStatus = parameterstatus.Matched
		}
	}
	return taskFlow, err
}

func (m *OBParameterManager) IsDeleting() bool {
	return !m.OBParameter.ObjectMeta.DeletionTimestamp.IsZero()
}

func (m *OBParameterManager) CheckAndUpdateFinalizers() error {
	return nil
}

func (m *OBParameterManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		parameter := &v1alpha1.OBParameter{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.OBParameter.GetNamespace(),
			Name:      m.OBParameter.GetName(),
		}, parameter)
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		parameter.Status = *m.OBParameter.Status.DeepCopy()
		return m.Client.Status().Update(m.Ctx, parameter)
	})
}

func (m *OBParameterManager) UpdateStatus() error {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return errors.Wrap(err, "Get obcluster from K8s")
	}
	operationManager, err := GetSysOperationClient(m.Client, m.Logger, obcluster)
	if err != nil {
		m.Logger.Error(err, "Get operation manager failed")
		return errors.Wrapf(err, "Get operation manager")
	}
	if obcluster.Status.Status != clusterstatus.Running {
		m.OBParameter.Status.Status = parameterstatus.PendingOB
		m.Logger.Info("obcluster not in running status, skip compare parameters")
	} else {
		parameterInfoList, err := operationManager.GetParameter(m.OBParameter.Spec.Parameter.Name, nil)
		if err != nil {
			m.Logger.Error(err, "Get parameter info failed")
			return errors.Wrapf(err, "Get parameter info")
		}
		parameterMatched := true
		parameterValues := make([]v1alpha1.ParameterValue, 0)
		for _, parameterInfo := range parameterInfoList {
			parameterValue := v1alpha1.ParameterValue{
				Name:   parameterInfo.Name,
				Value:  parameterInfo.Value,
				Zone:   parameterInfo.Zone,
				Server: fmt.Sprintf("%s:%d", parameterInfo.SvrIp, parameterInfo.SvrPort),
			}
			parameterValues = append(parameterValues, parameterValue)
			if parameterInfo.Value != m.OBParameter.Spec.Parameter.Value {
				parameterMatched = false
			}
		}
		m.OBParameter.Status.Parameter = parameterValues
		if m.OBParameter.Status.Status != parameterstatus.NotMatch {
			if !parameterMatched {
				m.OBParameter.Status.Status = parameterstatus.NotMatch
			} else {
				m.OBParameter.Status.Status = parameterstatus.Matched
			}
		}
	}
	err = m.retryUpdateStatus()
	if err != nil {
		m.Logger.Error(err, "Got error when update obparameter status")
	}
	return nil
}

func (m *OBParameterManager) ClearTaskInfo() {
	m.OBParameter.Status.Status = parameterstatus.Matched
	m.OBParameter.Status.OperationContext = nil
}

func (m *OBParameterManager) FinishTask() {
	m.OBParameter.Status.Status = m.OBParameter.Status.OperationContext.TargetStatus
	m.OBParameter.Status.OperationContext = nil
}

func (m *OBParameterManager) HandleFailure() {
	operationContext := m.OBParameter.Status.OperationContext
	failureRule := operationContext.OnFailure
	switch failureRule.Strategy {
	case strategy.StartOver:
		m.OBParameter.Status.Status = failureRule.NextTryStatus
		m.OBParameter.Status.OperationContext = nil
	case strategy.RetryFromCurrent:
		operationContext.TaskStatus = taskstatus.Pending
	case strategy.Pause:
	}
}

func (m *OBParameterManager) GetTaskFunc(name string) (func() error, error) {
	switch name {
	case taskname.SetOBParameter:
		return m.SetOBParameter, nil
	default:
		return nil, errors.New("Can not find a function for task")
	}
}

func (m *OBParameterManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.OBParameter, corev1.EventTypeWarning, "task exec failed", err.Error())
}

func (m *OBParameterManager) SetOBParameter() error {
	operationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		m.Logger.Error(err, "Get operation manager failed")
		return errors.Wrapf(err, "Get operation manager")
	}
	err = operationManager.SetParameter(m.OBParameter.Spec.Parameter.Name, m.OBParameter.Spec.Parameter.Value, nil)
	if err != nil {
		m.Logger.Error(err, "Set parameter failed")
		return errors.Wrapf(err, "Set parameter")
	}
	return nil
}

func (m *OBParameterManager) generateNamespacedName(name string) types.NamespacedName {
	var namespacedName types.NamespacedName
	namespacedName.Namespace = m.OBParameter.Namespace
	namespacedName.Name = name
	return namespacedName
}

func (m *OBParameterManager) getOBCluster() (*v1alpha1.OBCluster, error) {
	// this label always exists
	clusterName, _ := m.OBParameter.Labels[oceanbaseconst.LabelRefOBCluster]
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(clusterName), obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	return obcluster, nil
}

func (m *OBParameterManager) getOceanbaseOperationManager() (*operation.OceanbaseOperationManager, error) {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return nil, errors.Wrap(err, "Get obcluster from K8s")
	}
	return GetSysOperationClient(m.Client, m.Logger, obcluster)
}
