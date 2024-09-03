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

package obtenantoperation

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	opresource "github.com/oceanbase/ob-operator/pkg/coordinator"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

var _ opresource.ResourceManager = &ObTenantOperationManager{}

type ObTenantOperationManager struct {
	Ctx      context.Context
	Resource *v1alpha1.OBTenantOperation
	Client   client.Client
	Recorder telemetry.Recorder
	Logger   *logr.Logger
}

func (m *ObTenantOperationManager) GetMeta() metav1.Object {
	return m.Resource.GetObjectMeta()
}

func (m *ObTenantOperationManager) GetStatus() string {
	return string(m.Resource.Status.Status)
}

func (m *ObTenantOperationManager) CheckAndUpdateFinalizers() error {
	return nil
}

func (m *ObTenantOperationManager) InitStatus() {
	var err error
	var tenant *v1alpha1.OBTenant
	switch m.Resource.Spec.Type {
	case constants.TenantOpChangePwd:
		tenant, err = m.getTenantCR(m.Resource.Spec.ChangePwd.Tenant)
		if err != nil {
			m.Logger.Error(err, "Failed to find tenant")
			break
		}
		m.Resource.Status.PrimaryTenant = tenant
	case constants.TenantOpFailover:
		tenant, err = m.getTenantCR(m.Resource.Spec.Failover.StandbyTenant)
		if err != nil {
			m.Logger.Error(err, "Failed to find activating tenant")
			break
		}
		if tenant.Status.TenantRole == constants.TenantRolePrimary {
			err = errors.New("activating tenant is not a standby tenant")
			m.Logger.Error(err, "Failed to find standby tenant")
			break
		}
		m.Resource.Status.PrimaryTenant = tenant
	case constants.TenantOpSwitchover:
		tenant, err = m.getTenantCR(m.Resource.Spec.Switchover.PrimaryTenant)
		if err != nil {
			m.Logger.Error(err, "Failed to find primary tenant")
			break
		}
		standbyTenant, err := m.getTenantCR(m.Resource.Spec.Switchover.StandbyTenant)
		if err != nil {
			m.Logger.Error(err, "Failed to find standby tenant")
			break
		}
		m.Resource.Status.PrimaryTenant = tenant
		m.Resource.Status.SecondaryTenant = standbyTenant
	case constants.TenantOpUpgrade, constants.TenantOpReplayLog:
		tenant, err = m.getTenantCR(*m.Resource.Spec.TargetTenant)
		if err != nil {
			m.Logger.Error(err, "Failed to find target tenant")
			break
		}
		m.Resource.Status.PrimaryTenant = tenant
	}
	if err != nil {
		m.PrintErrEvent(err)
		m.Resource.Status.Status = constants.TenantOpFailed
	} else {
		m.Resource.Status.Status = constants.TenantOpRunning
	}
}

func (m *ObTenantOperationManager) SetOperationContext(c *tasktypes.OperationContext) {
	m.Resource.Status.OperationContext = c
}

func (m *ObTenantOperationManager) ClearTaskInfo() {
	m.Resource.Status.Status = constants.TenantOpRunning
	m.Resource.Status.OperationContext = nil
}

func (m *ObTenantOperationManager) HandleFailure() {
	if m.Resource.DeletionTimestamp != nil {
		m.Resource.Status.OperationContext = nil
	} else {
		operationContext := m.Resource.Status.OperationContext
		failureRule := operationContext.OnFailure
		switch failureRule.Strategy {
		case strategy.StartOver:
			if m.Resource.Status.Status != apitypes.TenantOperationStatus(failureRule.NextTryStatus) {
				m.Resource.Status.Status = apitypes.TenantOperationStatus(failureRule.NextTryStatus)
				m.Resource.Status.OperationContext = nil
			} else {
				m.Resource.Status.OperationContext.Idx = 0
				m.Resource.Status.OperationContext.TaskStatus = ""
				m.Resource.Status.OperationContext.TaskId = ""
				m.Resource.Status.OperationContext.Task = ""
			}
		case strategy.RetryFromCurrent:
			operationContext.TaskStatus = taskstatus.Pending
		case strategy.Pause:
		default:
			m.Resource.Status.OperationContext = nil
			if failureRule.NextTryStatus == "" {
				m.Resource.Status.Status = constants.TenantOpFailed
			} else {
				m.Resource.Status.Status = apitypes.TenantOperationStatus(failureRule.NextTryStatus)
			}
		}
	}
}

func (m *ObTenantOperationManager) FinishTask() {
	m.Resource.Status.Status = apitypes.TenantOperationStatus(m.Resource.Status.OperationContext.TargetStatus)
	m.Resource.Status.OperationContext = nil
}

func (m *ObTenantOperationManager) UpdateStatus() error {
	return m.retryUpdateStatus()
}

func (m *ObTenantOperationManager) ArchiveResource() {
	m.Logger.Info("Archive obtenant operation", "obtenant operation", m.Resource.Name)
	m.Recorder.Event(m.Resource, "Archive", "", "Archive obtenant operation")
	m.Resource.Status.Status = "Failed"
	m.Resource.Status.OperationContext = nil
}

func (m *ObTenantOperationManager) GetTaskFunc(name tasktypes.TaskName) (tasktypes.TaskFunc, error) {
	return taskMap.GetTask(name, m)
}

func (m *ObTenantOperationManager) GetTaskFlow() (*tasktypes.TaskFlow, error) {
	if m.Resource.Status.OperationContext != nil {
		return tasktypes.NewTaskFlow(m.Resource.Status.OperationContext), nil
	}
	var taskFlow *tasktypes.TaskFlow
	var err error
	status := m.Resource.Status.Status
	switch status {
	case constants.TenantOpStarting:
		// taskFlow = low.CheckTenantCRExistenceFlow()
	case constants.TenantOpRunning:
		switch m.Resource.Spec.Type {
		case constants.TenantOpChangePwd:
			taskFlow = genChangeTenantRootPasswordFlow(m)
		case constants.TenantOpFailover:
			taskFlow = genActivateStandbyTenantOpFlow(m)
		case constants.TenantOpSwitchover:
			taskFlow = genSwitchoverTenantsFlow(m)
		case constants.TenantOpUpgrade:
			taskFlow = genUpgradeTenantFlow(m)
		case constants.TenantOpReplayLog:
			taskFlow = genReplayLogOfStandbyFlow(m)
		default:
			taskFlow = genUpdateOBTenantResourceFlow(m)
		}
	case constants.TenantOpReverting:
		switch m.Resource.Spec.Type {
		case constants.TenantOpSwitchover:
			taskFlow = genRevertSwitchoverTenantsFlow(m)
		default:
			err = errors.New("unsupported operation type")
		}
	case constants.TenantOpSuccessful, constants.TenantOpFailed:
		fallthrough
	default:
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if taskFlow.OperationContext.OnFailure.NextTryStatus == "" {
		taskFlow.OperationContext.OnFailure.NextTryStatus = string(constants.TenantOpFailed)
	}
	return taskFlow, nil
}

func (m *ObTenantOperationManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.Resource, corev1.EventTypeWarning, "Task failed", err.Error())
}

func (m *ObTenantOperationManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		resource := &v1alpha1.OBTenantOperation{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.Resource.GetNamespace(),
			Name:      m.Resource.GetName(),
		}, resource)
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		resource.Status = m.Resource.Status
		return m.Client.Status().Update(m.Ctx, m.Resource)
	})
}

func (m *ObTenantOperationManager) getTenantCR(tenantCRName string) (*v1alpha1.OBTenant, error) {
	tenant := &v1alpha1.OBTenant{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.Resource.Namespace,
		Name:      tenantCRName,
	}, tenant)
	if err != nil {
		return nil, errors.Wrap(err, "get tenant")
	}
	return tenant, nil
}

func (m *ObTenantOperationManager) appendOwnerTenantReference(tenant *v1alpha1.OBTenant) {
	meta := tenant.GetObjectMeta()
	m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Append owner tenant reference", "tenant", tenant, "metadata", meta)
	owners := make([]metav1.OwnerReference, 0)
	if m.Resource.OwnerReferences != nil {
		owners = append(owners, m.Resource.OwnerReferences...)
	}
	owners = append(owners, metav1.OwnerReference{
		APIVersion: tenant.APIVersion,
		Kind:       tenant.Kind,
		Name:       meta.GetName(),
		UID:        meta.GetUID(),
	})
	m.Resource.SetOwnerReferences(owners)
}
