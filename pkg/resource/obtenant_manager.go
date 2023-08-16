package resource

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/const/status/obtenant"
	zonestatus "github.com/oceanbase/ob-operator/pkg/const/status/obzone"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	"github.com/oceanbase/ob-operator/pkg/task"
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
	"github.com/pkg/errors"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OBTenantManager struct {
	ResourceManager
	OBTenant *v1alpha1.OBTenant
	Ctx      context.Context
	Client   client.Client
	Recorder record.EventRecorder
	Logger   *logr.Logger
}


func (m *OBTenantManager) getOceanbaseOperationManager() (*operation.OceanbaseOperationManager, error) {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return nil, errors.Wrap(err, "Get obcluster from K8s")
	}
	return GetOceanbaseOperationManagerFromOBCluster(m.Client, obcluster)
}

func (m *OBTenantManager) IsNewResource() bool {
	return m.OBTenant.Status.Status == ""
}

func (m *OBTenantManager) IsDeleting() bool {
	return !m.OBTenant.ObjectMeta.DeletionTimestamp.IsZero()
}


func (m *OBTenantManager) InitStatus() {
	m.Logger.Info("newly created cluster, init status")
	status := v1alpha1.OBTenantStatus{
		Status:           obtenant.Creating,
		Pools:            make([]v1alpha1.ResourcePoolStatus, 0, len(m.OBTenant.Spec.Pools)),
		ConnectWhiteList: m.OBTenant.Status.ConnectWhiteList,
		Charset:          m.OBTenant.Status.Charset,
	}
	m.OBTenant.Status = status
}

func (m *OBTenantManager) SetOperationContext(ctx *v1alpha1.OperationContext) {
	m.OBTenant.Status.OperationContext = ctx
}
func (m *OBTenantManager) ClearTaskInfo() {
	m.OBTenant.Status.Status = obtenant.Running
	m.OBTenant.Status.OperationContext = nil
}

func (m *OBTenantManager) FinishTask() {
	m.OBTenant.Status.Status = m.OBTenant.Status.OperationContext.TargetStatus
	m.OBTenant.Status.OperationContext = nil
}

func (m *OBTenantManager) UpdateStatus() error {
	if m.hasModifiedTenantTask() {
		m.OBTenant.Status.Status = obtenant.Maintaining

		m.Logger.Info("update obtenant status", "status", m.OBTenant.Status)
		m.Logger.Info("update obtenant status", "operation context", m.OBTenant.Status.OperationContext)

		err := m.Client.Status().Update(m.Ctx, m.OBTenant)
		if err != nil {
			m.Logger.Error(err, "Got error when update obtenant status")
			return err
		}
	}
	return nil
}

func (m *OBTenantManager) CheckAndUpdateFinalizers() error {
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
	} else if m.IsDeleting() {
		m.Logger.Info("OBTenant is deleting")
		finalizerFinished = m.OBTenant.Status.Status == zonestatus.FinalizerFinished
	}

	if finalizerFinished {
		m.Logger.Info("Finalizer finished")
		m.OBTenant.ObjectMeta.Finalizers = make([]string, 0)

		err := m.Client.Update(m.Ctx, m.OBTenant)
		if err != nil {
			m.Logger.Error(err, "update obtenant instance failed")
			return errors.Wrapf(err, "Update obtenant %s in K8s failed", m.OBTenant.Name)
		}
	}
	return nil
}

func (m *OBTenantManager) GetTaskFunc(taskName string) (func() error, error) {
	switch taskName {
	case taskname.CreateTenant:
		return m.CreateTenantTask, nil
	case taskname.MaintainTenant:
		return m.MaintainTenantTask, nil
	case taskname.DeleteTenant:
		return m.DeleteTenantTask, nil
	default:
		return nil, errors.Errorf("Can not find an function for task %s", taskName)
	}
}

func (m *OBTenantManager) GetTaskFlow() (*task.TaskFlow, error) {
	// exists unfinished task flow, return the last task flow
	if m.OBTenant.Status.OperationContext != nil {
		m.Logger.Info("get task flow from observer status")
		return task.NewTaskFlow(m.OBTenant.Status.OperationContext), nil
	}

	m.Logger.Info("create task flow according to observer status")

	switch m.OBTenant.Status.Status {
	case obtenant.Creating:
		return task.GetRegistry().Get(flowname.CreateTenant)
	case obtenant.Maintaining:
		m.Logger.Info("Get task flow when obtenant maintaining")
		return task.GetRegistry().Get(flowname.MaintainTenant)
	case obtenant.Deleting:
		m.Logger.Info("Get task flow when obtenant deleting")
		return task.GetRegistry().Get(flowname.DeleteTenant)
	default:
		return nil, nil
	}
}

// ---------- K8S API Helper ----------

func (m *OBTenantManager) generateNamespacedName(name string) types.NamespacedName {
	var namespacedName types.NamespacedName
	namespacedName.Namespace = m.OBTenant.Namespace
	namespacedName.Name = name
	return namespacedName
}


func (m *OBTenantManager) getOBCluster() (*v1alpha1.OBCluster, error) {
	// this label always exists
	clusterName, _ := m.OBTenant.Labels[oceanbaseconst.LabelRefOBCluster]
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(clusterName), obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	return obcluster, nil
}

func (m *OBTenantManager) getObTenant() (*v1alpha1.OBTenant, error) {
	var TenantCurrent *v1alpha1.OBTenant
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(m.OBTenant.Name), TenantCurrent)
	if err != nil {
		klog.Errorln(err)
	}
	return TenantCurrent, err
}

func (m *OBTenantManager) PatchStatus(newTenant, curTenant *v1alpha1.OBTenant) error {
	err := m.Client.Status().Patch(m.Ctx, newTenant, client.MergeFrom(curTenant.DeepCopyObject().(client.Object)))

	if err != nil {
		klog.Errorln(err)
		return err
	}
	return nil
}


