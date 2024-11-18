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

package obtenantvariable

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	obtenantvariablestatus "github.com/oceanbase/ob-operator/internal/const/status/obtenantvariable"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	opresource "github.com/oceanbase/ob-operator/pkg/coordinator"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

var _ opresource.ResourceManager = &OBTenantVariableManager{}

type OBTenantVariableManager struct {
	Ctx              context.Context
	OBTenantVariable *v1alpha1.OBTenantVariable
	Client           client.Client
	Recorder         telemetry.Recorder
	Logger           *logr.Logger
}

func (m *OBTenantVariableManager) GetMeta() metav1.Object {
	return m.OBTenantVariable.GetObjectMeta()
}

func (m *OBTenantVariableManager) GetStatus() string {
	return m.OBTenantVariable.Status.Status
}

func (m *OBTenantVariableManager) InitStatus() {
	m.Logger.Info("Newly created obtenantvariable, init status")
	status := v1alpha1.OBTenantVariableStatus{
		Status: obtenantvariablestatus.New,
		Variable: apitypes.Variable{
			Name:  m.OBTenantVariable.Spec.Variable.Name,
			Value: "",
		},
	}
	m.OBTenantVariable.Status = status
}

func (m *OBTenantVariableManager) SetOperationContext(c *tasktypes.OperationContext) {
	m.OBTenantVariable.Status.OperationContext = c
}

func (m *OBTenantVariableManager) GetTaskFlow() (*tasktypes.TaskFlow, error) {
	// exists unfinished task flow, return the last task flow
	if m.OBTenantVariable.Status.OperationContext != nil {
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Get task flow from obtenantvariable status")
		return tasktypes.NewTaskFlow(m.OBTenantVariable.Status.OperationContext), nil
	}

	// return task flow depends on status

	var taskFlow *tasktypes.TaskFlow
	m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Create task flow according to obtenantvariable status")
	switch m.OBTenantVariable.Status.Status {
	// only need to handle variable not match
	case obtenantvariablestatus.NotMatch:
		taskFlow = genSetOBTenantVariableFlow(m)
	default:
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("No need to run anything for obtenantvariable")
		return nil, nil
	}

	if taskFlow.OperationContext.OnFailure.Strategy == "" {
		taskFlow.OperationContext.OnFailure.Strategy = strategy.StartOver
		if taskFlow.OperationContext.OnFailure.NextTryStatus == "" {
			taskFlow.OperationContext.OnFailure.NextTryStatus = obtenantvariablestatus.Matched
		}
	}
	return taskFlow, nil
}

func (m *OBTenantVariableManager) CheckAndUpdateFinalizers() error {
	return nil
}

func (m *OBTenantVariableManager) UpdateStatus() error {
	operationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		m.Logger.Error(err, "Get operation manager failed")
		return errors.Wrapf(err, "Get operation manager")
	}
	variable, err := operationManager.GetGlobalVariable(m.Ctx, m.OBTenantVariable.Spec.Variable.Name)
	if err != nil {
		m.Logger.Error(err, "Get tenant variable info failed")
		return errors.Wrapf(err, "Get tenant variable info")
	}
	m.OBTenantVariable.Status.Variable = apitypes.Variable{
		Name:  variable.Name,
		Value: variable.Value,
	}
	if m.OBTenantVariable.Status.Status != obtenantvariablestatus.NotMatch {
		if variable.Value != m.OBTenantVariable.Spec.Variable.Value {
			m.OBTenantVariable.Status.Status = obtenantvariablestatus.NotMatch
		} else {
			m.OBTenantVariable.Status.Status = obtenantvariablestatus.Matched
		}
	}
	err = m.retryUpdateStatus()
	if err != nil {
		m.Logger.Error(err, "Got error when update obtenantvariable status")
	}
	return nil
}

func (m *OBTenantVariableManager) ClearTaskInfo() {
	m.OBTenantVariable.Status.Status = obtenantvariablestatus.Matched
	m.OBTenantVariable.Status.OperationContext = nil
}

func (m *OBTenantVariableManager) FinishTask() {
	m.OBTenantVariable.Status.Status = m.OBTenantVariable.Status.OperationContext.TargetStatus
	m.OBTenantVariable.Status.OperationContext = nil
}

func (m *OBTenantVariableManager) HandleFailure() {
	operationContext := m.OBTenantVariable.Status.OperationContext
	failureRule := operationContext.OnFailure
	switch failureRule.Strategy {
	case strategy.StartOver:
		if m.OBTenantVariable.Status.Status != failureRule.NextTryStatus {
			m.OBTenantVariable.Status.Status = failureRule.NextTryStatus
			m.OBTenantVariable.Status.OperationContext = nil
		} else {
			m.OBTenantVariable.Status.OperationContext.Idx = 0
			m.OBTenantVariable.Status.OperationContext.TaskStatus = ""
			m.OBTenantVariable.Status.OperationContext.TaskId = ""
			m.OBTenantVariable.Status.OperationContext.Task = ""
		}
	case strategy.RetryFromCurrent:
		operationContext.TaskStatus = taskstatus.Pending
	case strategy.Pause:
	}
}

func (m *OBTenantVariableManager) GetTaskFunc(name tasktypes.TaskName) (tasktypes.TaskFunc, error) {
	return taskMap.GetTask(name, m)
}

func (m *OBTenantVariableManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.OBTenantVariable, corev1.EventTypeWarning, "Task failed", err.Error())
}

func (m *OBTenantVariableManager) ArchiveResource() {
	m.Logger.Info("Archive obtenantvariable", "obtenantvariable", m.OBTenantVariable.Name)
	m.Recorder.Event(m.OBTenantVariable, "Archive", "", "Archive obtenantvariable")
	m.OBTenantVariable.Status.Status = "Failed"
	m.OBTenantVariable.Status.OperationContext = nil
}
