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
	"reflect"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	kuberesource "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/const/status/tenantstatus"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/status/tenant"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	"github.com/oceanbase/ob-operator/pkg/task"
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/task/status"
	"github.com/oceanbase/ob-operator/pkg/task/strategy"
)

type OBTenantManager struct {
	ResourceManager
	OBTenant *v1alpha1.OBTenant
	Ctx      context.Context
	Client   client.Client
	Recorder record.EventRecorder
	Logger   *logr.Logger
}

// TODO add lock to be thread safe, and read/write whitelist from/to DB
var GlobalWhiteListMap = make(map[string]string, 0)

func (m *OBTenantManager) getClusterSysClient() (*operation.OceanbaseOperationManager, error) {
	obcluster, err := m.getOBCluster()
	if err != nil {
		m.Logger.Error(err, "get obcluster from k8s failed",
			"clusterName", m.OBTenant.Spec.ClusterName, "tenantName", m.OBTenant.Spec.TenantName)
		return nil, errors.Wrap(err, "Get obcluster from K8s failed")
	}
	return GetOceanbaseOperationManagerFromOBCluster(m.Client, m.Logger, obcluster)
}

func (m *OBTenantManager) getTenantClient() (*operation.OceanbaseOperationManager, error) {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return nil, errors.Wrap(err, "Get obcluster from K8s")
	}
	return GetTenantOperationClient(m.Client, m.Logger, obcluster, m.OBTenant.Spec.TenantName, m.OBTenant.Status.Credentials.Root)
}

func (m *OBTenantManager) IsNewResource() bool {
	return m.OBTenant.Status.Status == ""
}

func (m *OBTenantManager) IsDeleting() bool {
	return !m.OBTenant.ObjectMeta.DeletionTimestamp.IsZero()
}

func (m *OBTenantManager) InitStatus() {
	m.OBTenant.Status = v1alpha1.OBTenantStatus{
		Pools: make([]v1alpha1.ResourcePoolStatus, 0, len(m.OBTenant.Spec.Pools)),
	}
	m.OBTenant.Status.Credentials = v1alpha1.TenantCredentials{
		Root: "obtenant-default-pwd",
	}
	m.OBTenant.Status.TenantRole = m.OBTenant.Spec.TenantRole

	if m.OBTenant.Spec.Source != nil && m.OBTenant.Spec.Source.Restore != nil {
		m.OBTenant.Status.Status = tenantstatus.Restoring
	} else if m.OBTenant.Spec.Source != nil && m.OBTenant.Spec.Source.Tenant != nil {
		m.OBTenant.Status.Status = tenantstatus.CreatingEmptyStandby
	} else {
		m.OBTenant.Status.Status = tenantstatus.CreatingTenant
	}
}

func (m *OBTenantManager) SetOperationContext(ctx *v1alpha1.OperationContext) {
	m.OBTenant.Status.OperationContext = ctx
}

func (m *OBTenantManager) ClearTaskInfo() {
	m.OBTenant.Status.Status = tenantstatus.Running
	m.OBTenant.Status.OperationContext = nil
}

func (m *OBTenantManager) HandleFailure() {
	if m.IsDeleting() {
		m.OBTenant.Status.Status = tenantstatus.DeletingTenant
		m.OBTenant.Status.OperationContext = nil
	} else {
		operationContext := m.OBTenant.Status.OperationContext
		failureRule := operationContext.OnFailure
		switch failureRule.Strategy {
		case strategy.StartOver:
			m.OBTenant.Status.Status = failureRule.NextTryStatus
			m.OBTenant.Status.OperationContext = nil
		case strategy.RetryFromCurrent:
			operationContext.TaskStatus = taskstatus.Pending
		case strategy.Pause:
		}
	}
}

func (m *OBTenantManager) FinishTask() {
	m.OBTenant.Status.Status = m.OBTenant.Status.OperationContext.TargetStatus
	m.OBTenant.Status.OperationContext = nil
}

func (m *OBTenantManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		obtenant := &v1alpha1.OBTenant{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.OBTenant.GetNamespace(),
			Name:      m.OBTenant.GetName(),
		}, obtenant)
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		m.OBTenant.Status.DeepCopyInto(&obtenant.Status)
		return m.Client.Status().Update(m.Ctx, obtenant)
	})
}

func (m *OBTenantManager) UpdateStatus() error {
	obtenantName := m.OBTenant.Spec.TenantName
	var err error
	if m.OBTenant.Status.Status == tenantstatus.FinalizerFinished {
		m.Logger.Info("OBTenant has remove Finalizer", "tenantName", obtenantName)
		return nil
	} else if m.IsDeleting() {
		m.OBTenant.Status.Status = tenantstatus.DeletingTenant
	} else if m.OBTenant.Status.Status == tenantstatus.Restoring &&
		m.OBTenant.Spec.Source != nil &&
		m.OBTenant.Spec.Source.Restore != nil &&
		m.OBTenant.Spec.Source.Restore.Cancel {
		m.OBTenant.Status.OperationContext = nil
		m.OBTenant.Status.Status = tenantstatus.CancelingRestore
	} else if m.OBTenant.Status.Status != tenantstatus.Running {
		m.Logger.Info(fmt.Sprintf("OBTenant status is %s (not running), skip compare", m.OBTenant.Status.Status))
	} else {
		// build tenant status from DB
		tenantStatusCurrent, err := m.buildTenantStatus()
		if err != nil {
			m.Logger.Error(err, "Got error when build obtenant status from DB")
			return err
		}
		m.OBTenant.Status = *tenantStatusCurrent

		nextStatus, err := m.NextStatus()
		if err != nil {
			return err
		}
		m.OBTenant.Status.Status = nextStatus
	}
	m.Logger.Info("update obtenant status", "status", m.OBTenant.Status, "operation context", m.OBTenant.Status.OperationContext)
	err = m.retryUpdateStatus()
	if err != nil {
		m.Logger.Error(err, "Got error when update obtenant status")
		return err
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
		finalizerFinished = m.OBTenant.Status.Status == tenantstatus.FinalizerFinished
	}
	if finalizerFinished {
		m.Logger.Info("Obtenant Finalizer finished")
		m.OBTenant.ObjectMeta.Finalizers = make([]string, 0)
		err := m.Client.Update(m.Ctx, m.OBTenant)
		if err != nil {
			m.Logger.Error(err, "update observer instance failed")
			return errors.Wrapf(err, "Update obtenant %s in K8s failed", m.OBTenant.Spec.TenantName)
		}
	}
	return nil
}

func (m *OBTenantManager) GetTaskFunc(taskName string) (func() error, error) {
	switch taskName {
	case taskname.CheckTenant:
		return m.CheckTenantTask, nil
	case taskname.CheckPoolAndUnitConfig:
		return m.CheckPoolAndConfigTask, nil
	case taskname.CreateTenant:
		return m.CreateTenantTaskWithClear, nil
	case taskname.CreateResourcePoolAndUnitConfig:
		return m.CreateResourcePoolAndConfigTask, nil
	case taskname.MaintainCharset:
		return m.CheckAndApplyCharset, nil
	case taskname.MaintainUnitNum:
		return m.CheckAndApplyUnitNum, nil
	case taskname.MaintainWhiteList:
		return m.CheckAndApplyWhiteList, nil
	case taskname.MaintainPrimaryZone:
		return m.CheckAndApplyPrimaryZone, nil
	case taskname.MaintainLocality:
		return m.CheckAndApplyLocality, nil
	case taskname.AddResourcePool:
		return m.AddPoolTask, nil
	case taskname.DeleteResourcePool:
		return m.DeletePoolTask, nil
	case taskname.MaintainUnitConfig:
		return m.MaintainUnitConfigTask, nil
	case taskname.DeleteTenant:
		return m.DeleteTenantTask, nil
	case taskname.CreateRestoreJobCR:
		return m.CreateTenantRestoreJobCR, nil
	case taskname.WatchRestoreJobToFinish:
		return m.WatchRestoreJobToFinish, nil
	case taskname.CancelRestoreJob:
		return m.CancelTenantRestoreJob, nil
	case taskname.CreateEmptyStandbyTenant:
		return m.CreateEmptyStandbyTenant, nil
	case taskname.CreateUsersByCredentials:
		return m.CreateUserByCredentialSec, nil
	default:
		return nil, errors.Errorf("Can not find an function for task %s", taskName)
	}
}

func (m *OBTenantManager) GetTaskFlow() (*task.TaskFlow, error) {
	// exists unfinished task flow, return the last task flow
	if m.OBTenant.Status.OperationContext != nil {
		m.Logger.Info("get task flow from obtenant status")
		return task.NewTaskFlow(m.OBTenant.Status.OperationContext), nil
	}

	m.Logger.Info("create task flow according to obtenant status")
	var taskFlow *task.TaskFlow
	var err error

	switch m.OBTenant.Status.Status {
	case tenantstatus.CreatingTenant:
		m.Logger.Info("Get task flow when creating tenant")
		taskFlow, err = task.GetRegistry().Get(flowname.CreateTenant)
	case tenantstatus.MaintainingWhiteList:
		m.Logger.Info("Get task flow when obtenant maintaining white list")
		taskFlow, err = task.GetRegistry().Get(flowname.MaintainWhiteList)
	case tenantstatus.MaintainingCharset:
		m.Logger.Info("Get task flow when obtenant maintaining charset")
		taskFlow, err = task.GetRegistry().Get(flowname.MaintainCharset)
	case tenantstatus.MaintainingUnitNum:
		m.Logger.Info("Get task flow when obtenant maintaining unit num")
		taskFlow, err = task.GetRegistry().Get(flowname.MaintainUnitNum)
	case tenantstatus.MaintainingPrimaryZone:
		m.Logger.Info("Get task flow when obtenant maintaining primary zone")
		taskFlow, err = task.GetRegistry().Get(flowname.MaintainPrimaryZone)
	case tenantstatus.MaintainingLocality:
		m.Logger.Info("Get task flow when obtenant maintaining locality")
		taskFlow, err = task.GetRegistry().Get(flowname.MaintainLocality)
	case tenantstatus.AddingResourcePool:
		m.Logger.Info("Get task flow when obtenant adding pool")
		taskFlow, err = task.GetRegistry().Get(flowname.AddPool)
	case tenantstatus.DeletingResourcePool:
		m.Logger.Info("Get task flow when obtenant deleting list")
		taskFlow, err = task.GetRegistry().Get(flowname.DeletePool)
	case tenantstatus.MaintainingUnitConfig:
		m.Logger.Info("Get task flow when obtenant maintaining unit config")
		taskFlow, err = task.GetRegistry().Get(flowname.MaintainUnitConfig)
	case tenantstatus.DeletingTenant:
		m.Logger.Info("Get task flow when deleting tenant")
		taskFlow, err = task.GetRegistry().Get(flowname.DeleteTenant)
	case tenantstatus.PausingReconcile:
		m.Logger.Error(errors.New("obtenant pause reconcile"),
			"obtenant pause reconcile, please set status to running after manually resolving problem")
		return nil, nil
	case tenantstatus.Restoring:
		taskFlow, err = task.GetRegistry().Get(flowname.RestoreTenant)
	case tenantstatus.CancelingRestore:
		taskFlow, err = task.GetRegistry().Get(flowname.CancelRestoreFlow)
	case tenantstatus.CreatingEmptyStandby:
		taskFlow, err = task.GetRegistry().Get(flowname.CreateEmptyStandbyTenant)
	default:
		m.Logger.Info("no need to run anything for obtenant")
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if taskFlow.OperationContext.OnFailure.Strategy == "" {
		taskFlow.OperationContext.OnFailure.Strategy = strategy.StartOver
		if taskFlow.OperationContext.OnFailure.NextTryStatus == "" {
			taskFlow.OperationContext.OnFailure.NextTryStatus = tenantstatus.Running
		}
	}

	return taskFlow, nil
}

func (m *OBTenantManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.OBTenant, corev1.EventTypeWarning, "task exec failed", err.Error())
}

// ---------- K8S API Helper ----------

func (m *OBTenantManager) generateNamespacedName(name string) types.NamespacedName {
	var namespacedName types.NamespacedName
	namespacedName.Namespace = m.OBTenant.Namespace
	namespacedName.Name = name
	return namespacedName
}

func (m *OBTenantManager) getOBCluster() (*v1alpha1.OBCluster, error) {
	clusterName := m.OBTenant.Spec.ClusterName
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(clusterName), obcluster)
	if err != nil {
		m.Logger.Error(err, "get obcluster failed", "clusterName", clusterName, "namespaced", m.OBTenant.Namespace)
		return nil, errors.Wrap(err, "get obcluster failed")
	}
	return obcluster, nil
}

func (m *OBTenantManager) getObTenant() (*v1alpha1.OBTenant, error) {
	var obtenant *v1alpha1.OBTenant
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(m.OBTenant.Spec.TenantName), obtenant)
	if err != nil {
		return nil, errors.Wrap(err, "get obtenant")
	}
	return obtenant, nil
}

// --------- compare spec and status ----------

func (m *OBTenantManager) NextStatus() (string, error) {
	tenantName := m.OBTenant.Spec.TenantName

	// note: change order of state check functions may cause bugs
	hasModifiedResourcePool := m.hasToAddPool()
	if hasModifiedResourcePool {
		m.Logger.Info("Maintain Tenant ----- Resource Pool modified", "tenantName", tenantName)
		return tenantstatus.AddingResourcePool, nil
	}
	hasModifiedTenant := m.hasToDeletePool()
	if hasModifiedTenant {
		m.Logger.Info("Maintain Tenant ----- Tenant modified", "tenantName", tenantName)
		return tenantstatus.DeletingResourcePool, nil
	}
	hasModifiedLocality := m.hasModifiedLocality()
	if hasModifiedLocality {
		m.Logger.Info("tenant locality modified", "tenantName", tenantName)
		return tenantstatus.MaintainingLocality, nil
	}
	hasModifiedPriority := m.hasModifiedPrimaryZone()
	if hasModifiedPriority {
		m.Logger.Info("tenant PrimaryZone modified", "tenantName", tenantName)
		return tenantstatus.MaintainingPrimaryZone, nil
	}
	hasModifiedTcpInvitedNode := m.hasModifiedWhiteList()
	if hasModifiedTcpInvitedNode {
		m.Logger.Info("tenant WhiteList modified", "tenantName", tenantName)
		return tenantstatus.MaintainingWhiteList, nil
	}
	hasModifiedCharset := m.hasModifiedCharset()
	if hasModifiedCharset {
		m.Logger.Info("tenant charset modified", "tenantName", tenantName)
		return tenantstatus.MaintainingCharset, nil
	}
	hasModifiedUnitNum := m.hasModifiedUnitNum()
	if hasModifiedUnitNum {
		m.Logger.Info("tenant unitNum modified", "tenantName", tenantName)
		return tenantstatus.MaintainingUnitNum, nil
	}
	hasModifiedUnitConfig, err := m.hasModifiedUnitConfig()
	if err != nil {
		return tenantstatus.Running, err
	}
	if hasModifiedUnitConfig {
		m.Logger.Info("tenant UnitConfig modified", "tenantName", tenantName)
		return tenantstatus.MaintainingUnitConfig, nil
	}
	return tenantstatus.Running, nil
}

// ---------- Check function ----------

func (m *OBTenantManager) hasModifiedWhiteList() bool {
	specWhiteList := m.OBTenant.Spec.ConnectWhiteList
	statusWhiteList := m.OBTenant.Status.TenantRecordInfo.ConnectWhiteList

	if specWhiteList == "" {
		specWhiteList = tenant.DefaultOBTcpInvitedNodes
	}
	if specWhiteList != statusWhiteList {
		return true
	}
	return false
}

func (m *OBTenantManager) hasModifiedUnitConfig() (bool, error) {
	tenantName := m.OBTenant.Spec.TenantName

	version, err := m.getOBVersion()
	if err != nil {
		m.Logger.Error(err, "maintain tenant failed, check and apply unitConfigV4", "tenantName", tenantName)
		return false, err
	}
	if string(version[0]) == tenant.Version4 {
		return m.hasModifiedUnitConfigV4(), nil
	}
	return false, errors.New("no match version for check and set unit config")
}

func (m *OBTenantManager) hasModifiedUnitConfigV4() bool {
	specUnitConfigMap := m.generateSpecUnitConfigV4Map(m.OBTenant.Spec)
	statusUnitConfigMap := m.GenerateStatusUnitConfigV4Map(m.OBTenant.Status)
	for _, pool := range m.OBTenant.Spec.Pools {
		specUnitConfig := specUnitConfigMap[pool.Zone]
		statusUnitConfig, statusExist := statusUnitConfigMap[pool.Zone]

		// If status does not exist, Continue to check UnitConfig of the next ResourcePool
		// while Add and delete a pool in the CheckAndApplyResourcePool
		if !statusExist {
			continue
		}

		if !IsUnitConfigV4Equal(specUnitConfig, statusUnitConfig) {
			return true
		}
	}
	return false
}

func (m *OBTenantManager) hasToAddPool() bool {
	poolsForAdd := m.getPoolsForAdd()
	if len(poolsForAdd) > 0 {
		return true
	}
	return false
}

func (m *OBTenantManager) hasToDeletePool() bool {
	poolsForDelete := m.getPoolsForDelete()
	if len(poolsForDelete) > 0 {
		return true
	}

	return false
}

func (m *OBTenantManager) hasModifiedUnitNum() bool {
	// handle pool unitNum changed
	if m.OBTenant.Spec.UnitNumber != m.OBTenant.Status.TenantRecordInfo.UnitNumber {
		return true
	}
	return false
}

func (m *OBTenantManager) hasModifiedPrimaryZone() bool {
	specPrimaryZone := m.generateSpecPrimaryZone(m.OBTenant.Spec.Pools)
	statusPrimaryZone := m.generateStatusPrimaryZone(m.OBTenant.Status.Pools)
	specPrimaryZoneMap := m.generatePrimaryZoneMap(specPrimaryZone)
	statusPrimaryZoneMap := m.generatePrimaryZoneMap(statusPrimaryZone)
	if !reflect.DeepEqual(specPrimaryZoneMap, statusPrimaryZoneMap) {
		return true
	}
	return false
}

func (m *OBTenantManager) hasModifiedLocality() bool {
	specLocalityMap := m.generateSpecLocalityMap(m.OBTenant.Spec.Pools)
	statusLocalityMap := m.generateStatusLocalityMap(m.OBTenant.Status.Pools)
	if !reflect.DeepEqual(specLocalityMap, statusLocalityMap) {
		return true
	}
	return false
}

func (m *OBTenantManager) hasModifiedCharset() bool {
	specCharset := m.OBTenant.Spec.Charset
	if specCharset == "" {
		specCharset = tenant.Charset
	}
	if specCharset != m.OBTenant.Status.TenantRecordInfo.Charset {
		return true
	}
	return false
}

// ---------- buildTenant function ----------

func (m *OBTenantManager) buildTenantStatus() (*v1alpha1.OBTenantStatus, error) {
	tenantName := m.OBTenant.Spec.TenantName
	tenantCurrentStatus := &v1alpha1.OBTenantStatus{
		Credentials: m.OBTenant.Status.Credentials,
		TenantRole:  m.OBTenant.Status.TenantRole,
		Source:      m.OBTenant.Status.Source,
	}

	tenantExist, err := m.tenantExist(tenantName)
	if err != nil {
		return nil, err
	}
	if !tenantExist {
		return nil, fmt.Errorf("Tenant not exist, Tenant name: %s", tenantName)
	}
	obtenant, err := m.getTenantByName(tenantName)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprint("Cannot Get Tenant Failed When Build Tenant Status", tenantName))
	}

	poolStatusList, err := m.buildPoolStatusList(obtenant)

	if err != nil {
		return nil, err
	}
	tenantCurrentStatus.Status = m.OBTenant.Status.Status
	tenantCurrentStatus.Pools = poolStatusList
	tenantCurrentStatus.OperationContext = m.OBTenant.Status.OperationContext

	tenantCurrentStatus.TenantRecordInfo = v1alpha1.TenantRecordInfo{}
	tenantCurrentStatus.TenantRecordInfo.TenantID = int(obtenant.TenantID)

	// TODO get whitelist from tenant account
	whitelist, exists := GlobalWhiteListMap[obtenant.TenantName]
	if exists {
		tenantCurrentStatus.TenantRecordInfo.ConnectWhiteList = whitelist
	} else {
		// try update whitelist after the manager restart
		GlobalWhiteListMap[obtenant.TenantName] = tenant.DefaultOBTcpInvitedNodes
		tenantCurrentStatus.TenantRecordInfo.ConnectWhiteList = GlobalWhiteListMap[obtenant.TenantName]
	}

	tenantCurrentStatus.TenantRecordInfo.UnitNumber = poolStatusList[0].UnitNumber
	charset, err := m.getCharset()
	if err != nil {
		return nil, err
	}
	tenantCurrentStatus.TenantRecordInfo.Charset = charset
	tenantCurrentStatus.TenantRecordInfo.Locality = obtenant.Locality
	tenantCurrentStatus.TenantRecordInfo.PrimaryZone = obtenant.PrimaryZone
	poolList := make([]string, 0)
	zoneList := make([]string, 0)
	for _, pool := range tenantCurrentStatus.Pools {
		poolList = append(poolList, m.generatePoolName(pool.ZoneList))
		zoneList = append(zoneList, pool.ZoneList)
	}
	tenantCurrentStatus.TenantRecordInfo.PoolList = strings.Join(poolList, ",")
	tenantCurrentStatus.TenantRecordInfo.ZoneList = strings.Join(zoneList, ",")
	tenantCurrentStatus.TenantRecordInfo.Collate = m.OBTenant.Spec.Collate

	// Root password changed
	if _, err = m.getTenantClient(); err != nil {
		tenantCurrentStatus.Credentials.Root = m.OBTenant.Spec.Credentials.Root
	}

	return tenantCurrentStatus, nil
}

func (m *OBTenantManager) buildPoolStatusList(obTenant *model.OBTenant) ([]v1alpha1.ResourcePoolStatus, error) {
	var poolStatusList []v1alpha1.ResourcePoolStatus
	var locality string
	var primaryZone string

	locality = obTenant.Locality
	primaryZone = obTenant.PrimaryZone
	statusTypeMap := m.generateStatusTypeMapFromLocalityStr(locality)
	specTypeMap := m.generateSpecLocalityMap(m.OBTenant.Spec.Pools)

	tenantID := obTenant.TenantID
	priorityMap := m.generateStatusPriorityMap(primaryZone)
	unitNumMap, err := m.generateStatusUnitNumMap(m.OBTenant.Spec.Pools)
	if err != nil {
		return poolStatusList, err
	}
	poolList, err := m.generateStatusZone(tenantID)
	if err != nil {
		return poolStatusList, err
	}
	for _, zoneList := range poolList {
		var poolCurrentStatus v1alpha1.ResourcePoolStatus
		poolCurrentStatus.ZoneList = zoneList
		localityType, exist := statusTypeMap[zoneList]
		if exist {
			poolCurrentStatus.Type = &localityType
		} else {
			poolCurrentStatus.Type = &v1alpha1.LocalityType{
				Name:     specTypeMap[zoneList].Name,
				Replica:  specTypeMap[zoneList].Replica,
				IsActive: false,
			}
		}
		poolCurrentStatus.UnitNumber = unitNumMap[zoneList]
		poolCurrentStatus.Priority = priorityMap[zoneList]
		poolCurrentStatus.UnitConfig, err = m.buildUnitConfigV4FromDB(zoneList, tenantID)
		if err != nil {
			return poolStatusList, err
		}
		poolCurrentStatus.Units, err = m.buildUnitStatusFromDB(zoneList, tenantID)
		if err != nil {
			return poolStatusList, err
		}
		poolStatusList = append(poolStatusList, poolCurrentStatus)
	}
	return poolStatusList, nil
}

func (m *OBTenantManager) generateStatusZone(tenantID int64) ([]string, error) {
	var zoneList []string
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return zoneList, errors.Wrap(err, "Get Sql Operator Error When Generating Zone For Tenant CR Status")
	}
	poolList, err := oceanbaseOperationManager.GetPoolList()
	if err != nil {
		return nil, errors.Wrap(err, "Get sql error when get pool list")
	}

	poolIdMap := make(map[int64]string, 0)
	for _, pool := range poolList {
		if pool.TenantID.Valid && pool.TenantID.Int64 == tenantID {
			poolIdMap[pool.ResourcePoolID] = pool.Name
		}
	}
	zoneMap := make(map[string]string, 0)
	unitList, err := oceanbaseOperationManager.GetUnitList()
	if err != nil {
		return nil, errors.Wrap(err, "Get sql error when get unit list")
	}
	for _, unit := range unitList {
		if poolIdMap[unit.ResourcePoolID] != "" && zoneMap[unit.Zone] == "" {
			zoneMap[unit.Zone] = unit.Zone
		}
	}
	for k := range zoneMap {
		zoneList = append(zoneList, k)
	}
	return zoneList, nil
}

func (m *OBTenantManager) buildUnitConfigV4FromDB(zone string, tenantID int64) (*v1alpha1.UnitConfig, error) {
	unitConfig := &v1alpha1.UnitConfig{}
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return nil, errors.Wrap(err, "Get Sql Operator Error When Building Resource Unit From DB")
	}
	unitList, err := oceanbaseOperationManager.GetUnitList()
	if err != nil {
		return unitConfig, errors.Wrap(err, "Get sql error when get unit list")
	}
	poolList, err := oceanbaseOperationManager.GetPoolList()
	if err != nil {
		return unitConfig, errors.Wrap(err, "Get sql error when get pool list")
	}
	unitConfigList, err := oceanbaseOperationManager.GetUnitConfigV4List()
	if err != nil {
		return unitConfig, errors.Wrap(err, "Get sql error when get unit config list")
	}
	var resourcePoolIDList []int
	for _, unit := range unitList {
		if unit.Zone == zone {
			resourcePoolIDList = append(resourcePoolIDList, int(unit.ResourcePoolID))
		}
	}
	for _, pool := range poolList {
		for _, resourcePoolID := range resourcePoolIDList {
			if resourcePoolID == int(pool.ResourcePoolID) && pool.TenantID.Valid && pool.TenantID.Int64 == tenantID {
				for _, unitConifg := range unitConfigList {
					if unitConifg.UnitConfigID == pool.UnitConfigID {
						unitConfig.MaxCPU, err = kuberesource.ParseQuantity(strconv.FormatFloat(unitConifg.MaxCPU, 'f', -1, 64))
						if err != nil {
							return nil, err
						}
						unitConfig.MinCPU, err = kuberesource.ParseQuantity(strconv.FormatFloat(unitConifg.MinCPU, 'f', -1, 64))
						if err != nil {
							return nil, err
						}
						unitConfig.MemorySize = *kuberesource.NewQuantity(unitConifg.MemorySize, kuberesource.DecimalSI)
						unitConfig.LogDiskSize = *kuberesource.NewQuantity(unitConifg.LogDiskSize, kuberesource.DecimalSI)
						unitConfig.MaxIops = int(unitConifg.MaxIops)
						unitConfig.MinIops = int(unitConifg.MinIops)
						unitConfig.IopsWeight = int(unitConifg.IopsWeight)
					}
				}
			}
		}
	}
	return unitConfig, nil
}

func (m *OBTenantManager) buildUnitStatusFromDB(zone string, tenantID int64) ([]v1alpha1.UnitStatus, error) {
	var unitList []v1alpha1.UnitStatus
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return unitList, errors.Wrap(err, "Get Sql Operator Error When Building Resource Unit From DB")
	}
	poolList, err := oceanbaseOperationManager.GetPoolList()
	if err != nil {
		return unitList, errors.Wrap(err, "Get sql error when get pool list")
	}
	var resourcePoolIDList []int64
	for _, pool := range poolList {
		if pool.TenantID.Valid && pool.TenantID.Int64 == tenantID {
			resourcePoolIDList = append(resourcePoolIDList, pool.ResourcePoolID)
		}
	}
	units, err := oceanbaseOperationManager.GetUnitList()
	if err != nil {
		return unitList, errors.Wrap(err, "Get Sql Operator Error When Building Resource Unit From DB")
	}
	for _, unit := range units {
		for _, poolId := range resourcePoolIDList {
			if unit.Zone == zone && poolId == unit.ResourcePoolID {
				var res v1alpha1.UnitStatus
				res.UnitId = int(unit.UnitID)
				res.ServerIP = unit.SvrIP
				res.ServerPort = int(unit.SvrPort)
				res.Status = unit.Status
				var migrateServer v1alpha1.MigrateServerStatus
				if unit.MigrateFromSvrIP.Valid {
					migrateServer.ServerIP = unit.MigrateFromSvrIP.String
				}
				if unit.MigrateFromSvrPort.Valid {
					migrateServer.ServerPort = int(unit.MigrateFromSvrPort.Int64)
				}
				res.Migrate = migrateServer
				unitList = append(unitList, res)
			}
		}
	}
	return unitList, nil
}
