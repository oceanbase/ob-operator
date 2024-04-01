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

package obparameter

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	parameterstatus "github.com/oceanbase/ob-operator/internal/const/status/obparameter"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	opresource "github.com/oceanbase/ob-operator/pkg/coordinator"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

var _ opresource.ResourceManager = &OBParameterManager{}

type OBParameterManager struct {
	Ctx         context.Context
	OBParameter *v1alpha1.OBParameter
	Client      client.Client
	Recorder    telemetry.Recorder
	Logger      *logr.Logger
}

func (m *OBParameterManager) IsNewResource() bool {
	return m.OBParameter.Status.Status == ""
}

func (m *OBParameterManager) GetStatus() string {
	return m.OBParameter.Status.Status
}

func (m *OBParameterManager) InitStatus() {
	m.Logger.Info("Newly created obparameter, init status")
	status := v1alpha1.OBParameterStatus{
		Status:    parameterstatus.New,
		Parameter: make([]apitypes.ParameterValue, 0),
	}
	m.OBParameter.Status = status
}

func (m *OBParameterManager) SetOperationContext(c *tasktypes.OperationContext) {
	m.OBParameter.Status.OperationContext = c
}

func (m *OBParameterManager) GetTaskFlow() (*tasktypes.TaskFlow, error) {
	// exists unfinished task flow, return the last task flow
	if m.OBParameter.Status.OperationContext != nil {
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Get task flow from obparameter status")
		return tasktypes.NewTaskFlow(m.OBParameter.Status.OperationContext), nil
	}

	// return task flow depends on status

	var taskFlow *tasktypes.TaskFlow
	m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Create task flow according to obparameter status")
	switch m.OBParameter.Status.Status {
	// only need to handle parameter not match
	case parameterstatus.NotMatch:
		taskFlow = genSetOBParameterFlow(m)
	default:
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("No need to run anything for obparameter")
		return nil, nil
	}

	if taskFlow.OperationContext.OnFailure.Strategy == "" {
		taskFlow.OperationContext.OnFailure.Strategy = strategy.StartOver
		if taskFlow.OperationContext.OnFailure.NextTryStatus == "" {
			taskFlow.OperationContext.OnFailure.NextTryStatus = parameterstatus.Matched
		}
	}
	return taskFlow, nil
}

func (m *OBParameterManager) IsDeleting() bool {
	ignoreDel, ok := resourceutils.GetAnnotationField(m.OBParameter, oceanbaseconst.AnnotationsIgnoreDeletion)
	return !m.OBParameter.ObjectMeta.DeletionTimestamp.IsZero() && (!ok || ignoreDel != "true")
}

func (m *OBParameterManager) CheckAndUpdateFinalizers() error {
	return nil
}

func (m *OBParameterManager) UpdateStatus() error {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return errors.Wrap(err, "Get obcluster from K8s")
	}
	operationManager, err := resourceutils.GetSysOperationClient(m.Client, m.Logger, obcluster)
	if err != nil {
		m.Logger.Error(err, "Get operation manager failed")
		return errors.Wrapf(err, "Get operation manager")
	}
	if obcluster.Status.Status != clusterstatus.Running {
		m.OBParameter.Status.Status = parameterstatus.PendingOB
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("OBCluster not in running status, skip compare parameters")
	} else {
		parameterInfoList, err := operationManager.GetParameter(m.OBParameter.Spec.Parameter.Name, nil)
		if err != nil {
			m.Logger.Error(err, "Get parameter info failed")
			return errors.Wrapf(err, "Get parameter info")
		}
		parameterMatched := true
		parameterValues := make([]apitypes.ParameterValue, 0)
		for _, parameterInfo := range parameterInfoList {
			parameterValue := apitypes.ParameterValue{
				Name:   parameterInfo.Name,
				Value:  parameterInfo.Value,
				Zone:   parameterInfo.Zone,
				Server: fmt.Sprintf("%s:%d", parameterInfo.SvrIp, parameterInfo.SvrPort),
			}
			parameterValues = append(parameterValues, parameterValue)
			if !strings.EqualFold(parameterInfo.Value, m.OBParameter.Spec.Parameter.Value) {
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
		if m.OBParameter.Status.Status != failureRule.NextTryStatus {
			m.OBParameter.Status.Status = failureRule.NextTryStatus
			m.OBParameter.Status.OperationContext = nil
		} else {
			m.OBParameter.Status.OperationContext.Idx = 0
			m.OBParameter.Status.OperationContext.TaskStatus = ""
			m.OBParameter.Status.OperationContext.TaskId = ""
			m.OBParameter.Status.OperationContext.Task = ""
		}
	case strategy.RetryFromCurrent:
		operationContext.TaskStatus = taskstatus.Pending
	case strategy.Pause:
	}
}

func (m *OBParameterManager) GetTaskFunc(name tasktypes.TaskName) (tasktypes.TaskFunc, error) {
	return taskMap.GetTask(name, m)
}

func (m *OBParameterManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.OBParameter, corev1.EventTypeWarning, "Task failed", err.Error())
}

func (m *OBParameterManager) ArchiveResource() {
	m.Logger.Info("Archive obparameter", "obparameter", m.OBParameter.Name)
	m.Recorder.Event(m.OBParameter, "Archive", "", "Archive obparameter")
	m.OBParameter.Status.Status = "Failed"
	m.OBParameter.Status.OperationContext = nil
}
