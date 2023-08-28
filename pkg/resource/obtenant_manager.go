package resource

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/const/status/tenantstatus"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/status/tenant"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	"github.com/oceanbase/ob-operator/pkg/task"
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/task/status"
	"github.com/oceanbase/ob-operator/pkg/task/fail"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	kuberesource "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
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
		m.Logger.Error(err, "get obcluster from k8s failed",
			"clusterName", m.OBTenant.Spec.ClusterName, "tenantName", m.OBTenant.Spec.TenantName)
		return nil, errors.Wrap(err, "Get obcluster from K8s failed")
	}
	return GetOceanbaseOperationManagerFromOBCluster(m.Client, m.Logger, obcluster)
}

func (m *OBTenantManager) IsNewResource() bool {
	return m.OBTenant.Status.Status == ""
}

func (m *OBTenantManager) IsDeleting() bool {
	return !m.OBTenant.ObjectMeta.DeletionTimestamp.IsZero()
}

func (m *OBTenantManager) InitStatus() {
	m.Logger.Info("newly created obtenant, init status")
	status := v1alpha1.OBTenantStatus{
		Status:           tenantstatus.Creating,
		Pools:            make([]v1alpha1.ResourcePoolStatus, 0, len(m.OBTenant.Spec.Pools)),
	}
	m.OBTenant.Status = status
}

func (m *OBTenantManager) SetOperationContext(ctx *v1alpha1.OperationContext) {
	m.OBTenant.Status.OperationContext = ctx
}

func (m *OBTenantManager) ClearTaskInfo() {
	m.OBTenant.Status.Status = tenantstatus.Running
	m.OBTenant.Status.OperationContext = nil
}

func (m *OBTenantManager) HandleFailure() {
	operationContext := m.OBTenant.Status.OperationContext
	failureRule := operationContext.FailureRule
	switch failureRule.Strategy {
	case fail.RetryTask:
		m.OBTenant.Status.Status = failureRule.NextTryStatus
	case fail.RetryCurrentStep:
		operationContext.TaskStatus = taskstatus.Pending
	case fail.PauseReconcile:
		m.OBTenant.Status.Status = tenantstatus.PausingReconcile
	}

	if m.IsClearOperationContextIfFailed() {
		m.OBTenant.Status.OperationContext = nil
	}

}

func (m *OBTenantManager) IsClearOperationContextIfFailed() bool {
	m.Logger.Info("debug:", "status", m.OBTenant.Status)
	return  m.OBTenant.Status.OperationContext.FailureRule.Strategy != fail.RetryCurrentStep
}

func (m *OBTenantManager) FinishTask() {
	m.OBTenant.Status.Status = m.OBTenant.Status.OperationContext.TargetStatus
	m.OBTenant.Status.OperationContext = nil
}

func (m *OBTenantManager) UpdateStatus() error {
 	obtenantName := m.OBTenant.Spec.TenantName
	var err error
	if m.OBTenant.Status.Status == tenantstatus.FinalizerFinished {
		m.Logger.Info("OBTenant has remove Finalizer", "tenantName", obtenantName)
		return nil
	} else if m.IsDeleting() {
		if m.OBTenant.Status.Status != tenantstatus.Deleting {
			m.Logger.Info("debug: income", "status", m.OBTenant.Status.Status)
			m.OBTenant.Status.Status = tenantstatus.Deleting
			m.OBTenant.Status.OperationContext = nil
			m.Logger.Info("OBTenant prepare deleting", "tenantName", obtenantName)
		}
	} else if m.OBTenant.Status.Status != tenantstatus.Running {
		m.Logger.Info(fmt.Sprintf("OBTenant status is %s (not running), skip compare", m.OBTenant.Status.Status))
	} else {
		// build tenant status from DB
		tenantStatusCurrent, err := m.BuildTenantStatus()
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
	err = m.Client.Status().Update(m.Ctx, m.OBTenant)
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
	case taskname.AddPool:
		return m.AddPoolTask, nil
	case taskname.DeletePool:
		return m.DeletePoolTask, nil
	case taskname.MaintainUnitConfig:
		return m.MaintainUnitConfigTask, nil
	case taskname.DeleteTenant:
		return m.DeleteTenantTask, nil
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
	case tenantstatus.Creating:
		m.Logger.Info("Get task flow when obtenant creating")
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
		taskFlow,err = task.GetRegistry().Get(flowname.MaintainPrimaryZone)
	case tenantstatus.MaintainingLocality:
		m.Logger.Info("Get task flow when obtenant maintaining locality")
		taskFlow, err = task.GetRegistry().Get(flowname.MaintainLocality)
	case tenantstatus.AddingPool:
		m.Logger.Info("Get task flow when obtenant adding pool")
		taskFlow, err = task.GetRegistry().Get(flowname.AddPool)
	case tenantstatus.DeletingPool:
		m.Logger.Info("Get task flow when obtenant deleting list")
		taskFlow, err =  task.GetRegistry().Get(flowname.DeletePool)
	case tenantstatus.MaintainingUnitConfig:
		m.Logger.Info("Get task flow when obtenant maintaining unit config")
		taskFlow,err  = task.GetRegistry().Get(flowname.MaintainUnitConfig)
	case tenantstatus.Deleting:
		m.Logger.Info("Get task flow when obtenant deleting")
		taskFlow,err = task.GetRegistry().Get(flowname.DeleteTenant)
	case tenantstatus.PausingReconcile:
		m.Logger.Error(errors.New("obtenant pause reconcile"),
			"obtenant pause reconcile, please set status to running after manually resolving problem")
		return nil,nil
	default:
		m.Logger.Info("no need to run anything for obtenant")
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if taskFlow.OperationContext.FailureRule.Strategy == "" {
		taskFlow.OperationContext.FailureRule.Strategy = fail.RetryTask
		if taskFlow.OperationContext.FailureRule.NextTryStatus == "" {
			taskFlow.OperationContext.FailureRule.NextTryStatus = tenantstatus.Running
		}
	}

	return taskFlow, nil
}

func (m *OBTenantManager) PrintErrEvent(err error)  {
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
	var TenantCurrent *v1alpha1.OBTenant
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(m.OBTenant.Spec.TenantName), TenantCurrent)
	if err != nil {
		return nil, errors.Wrap(err, "get obtenant")
	}
	return TenantCurrent, nil
}

// --------- compare spec and status ----------
func (m *OBTenantManager) NextStatus() (string, error) {
	tenantName := m.OBTenant.Spec.TenantName

	hasModifiedResourcePool := m.hasToAddPool()
	if hasModifiedResourcePool {
		m.Logger.Info("Maintain Tenant ----- Resource Pool modified", "tenantName", tenantName)
		return tenantstatus.AddingPool, nil
	}
	hasModifiedTenant := m.hasToDeletePool()
	if hasModifiedTenant {
		m.Logger.Info("Maintain Tenant ----- Tenant modified", "tenantName", tenantName)
		return tenantstatus.DeletingPool, nil
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
		m.Logger.Info("debug: TcpInvitedNode changed", "statusWhiteList", statusWhiteList, "statusWhiteList", statusWhiteList)
		return true
	}
	return false
}

func (m *OBTenantManager) hasModifiedUnitConfig() (bool, error) {
	tenantName := m.OBTenant.Spec.TenantName

	version, err := m.GetOBVersion()
	if err != nil {
		m.Logger.Error(err, "maintain tenant failed, check and apply unitConfigV4", "tenantName", tenantName)
		return false, err
	}
	switch string(version[0]) {
	case tenant.Version4:
		return m.hasModifiedUnitConfigV4(), nil
	}
	return false, errors.New("no match version for check and set unit config")
}

func (m *OBTenantManager) hasModifiedUnitConfigV4() bool {
	specUnitConfigMap := GenerateSpecUnitConfigV4Map(m.OBTenant.Spec)
	statusUnitConfigMap := GenerateStatusUnitConfigV4Map(m.OBTenant.Status)
	for _, pool := range m.OBTenant.Spec.Pools {
		specUnitConfig := specUnitConfigMap[pool.ZoneList]
		statusUnitConfig, statusExist := statusUnitConfigMap[pool.ZoneList]

		// If status does not exist, Continue to check UnitConfig of the next ResourcePool
		// while Add and delete a pool in the CheckAndApplyResourcePool
		if !statusExist{
			continue
		}

		if !IsUnitConfigV4Equal(specUnitConfig, statusUnitConfig) {
			m.Logger.Info("debug: unitConfig changed", "specUnitConfig", specUnitConfig, "statusUnitConfig", statusUnitConfig)
			return true
		}
	}
	return false
}

func (m *OBTenantManager) hasToAddPool() bool{
	poolsForAdd := m.GetPoolsForAdd()
	if len(poolsForAdd) > 0 {
		m.Logger.Info("debug: resourcePool for add", "poolsForAdd", poolsForAdd)
		return true
	}
	return false
}

func (m *OBTenantManager) hasToDeletePool() bool{
	poolsForDelete := m.GetPoolsForDelete()
	if len(poolsForDelete) > 0 {
		m.Logger.Info("debug: resourcePool for delete", "poolsForDelete", poolsForDelete)
		return true
	}

	return false
}

func (m *OBTenantManager) hasModifiedUnitNum() bool {
	// handle pool unitNum changed
	if m.OBTenant.Spec.UnitNumber != m.OBTenant.Status.TenantRecordInfo.UnitNumber {
		m.Logger.Info("debug: unitNumber changed", "specUnitNumber", m.OBTenant.Spec.UnitNumber,
			"statusUnitNumber", m.OBTenant.Status.TenantRecordInfo.UnitNumber)
		return true
	}
	return false
}

func (m *OBTenantManager) hasModifiedPrimaryZone() bool {
	specPrimaryZone := GenerateSpecPrimaryZone(m.OBTenant.Spec.Pools)
	statusPrimaryZone := GenerateStatusPrimaryZone(m.OBTenant.Status.Pools)
	specPrimaryZoneMap := GeneratePrimaryZoneMap(specPrimaryZone)
	statusPrimaryZoneMap := GeneratePrimaryZoneMap(statusPrimaryZone)
	if !reflect.DeepEqual(specPrimaryZoneMap, statusPrimaryZoneMap) {
		m.Logger.Info("debug: priority changed", "specPrimaryZoneMap", specPrimaryZoneMap,
			"statusPrimaryZoneMap", statusPrimaryZoneMap)
		return true
	}
	return false
}

func (m *OBTenantManager) hasModifiedLocality() bool {
	specLocalityMap := GenerateSpecLocalityMap(m.OBTenant.Spec.Pools)
	statusLocalityMap := GenerateStatusLocalityMap(m.OBTenant.Status.Pools)
	if !reflect.DeepEqual(specLocalityMap, statusLocalityMap) {
		m.Logger.Info("debug: locality changed", "specLocalityMap", specLocalityMap,
			"statusLocalityMap", statusLocalityMap)
		return true
	}
	return false
}

func (m *OBTenantManager) hasModifiedCharset() bool {
	specCharset := m.OBTenant.Spec.Charset
	if specCharset == "" {
		specCharset = tenant.Charset
	}
	if specCharset != m.OBTenant.Status.TenantRecordInfo.Charset{
		m.Logger.Info("debug: charset changed", "specCharset",specCharset,
			"statusCharset", m.OBTenant.Status.TenantRecordInfo.Charset)
		return true
	}
	return false
}
// ---------- buildTenant function ----------

func (m *OBTenantManager) BuildTenantStatus() (*v1alpha1.OBTenantStatus ,error) {
	tenantName := m.OBTenant.Spec.TenantName
	tenantCurrentStatus := &v1alpha1.OBTenantStatus{}

	tenantExist, err := m.TenantExist(tenantName)
	if err != nil {
		return nil, err
	}
	if !tenantExist {
		return nil, errors.New(fmt.Sprintf("Tenant not exist, Tenant name: %s", tenantName))
	}
	gvTenant, err := m.GetTenantByName(tenantName)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprint("Cannot Get Tenant Failed When Build Tenant Status", tenantName))
	}

	poolStatusList, err := m.BuildPoolStatusList(gvTenant)
	if err != nil {
		return nil, err
	}
	tenantCurrentStatus.Status = m.OBTenant.Status.Status
	tenantCurrentStatus.Pools = poolStatusList
	tenantCurrentStatus.OperationContext = m.OBTenant.Status.OperationContext

	tenantCurrentStatus.TenantRecordInfo = v1alpha1.TenantRecordInfo{}
	tenantCurrentStatus.TenantRecordInfo.TenantID = int(gvTenant.TenantID)
	tenantCurrentStatus.TenantRecordInfo.ConnectWhiteList, err = m.GetVariable(tenant.OBTcpInvitedNodes)
	if err != nil {
		return nil, err
	}
	tenantCurrentStatus.TenantRecordInfo.UnitNumber = poolStatusList[0].UnitNumber
	charset, err := m.GetCharset()
	if err != nil {
		return nil, err
	}
	tenantCurrentStatus.TenantRecordInfo.Charset = charset
	tenantCurrentStatus.TenantRecordInfo.Locality = gvTenant.Locality
	tenantCurrentStatus.TenantRecordInfo.PrimaryZone = gvTenant.PrimaryZone
	poolList := make([]string, 0)
	zoneList := make([]string, 0)
	for _, pool := range tenantCurrentStatus.Pools {
		poolList = append(poolList, m.GeneratePoolName(pool.ZoneList))
		zoneList = append(zoneList, pool.ZoneList)
	}
	tenantCurrentStatus.TenantRecordInfo.PoolList = strings.Join(poolList, ",")
	tenantCurrentStatus.TenantRecordInfo.ZoneList= strings.Join(zoneList, ",")
	tenantCurrentStatus.TenantRecordInfo.Collate = m.OBTenant.Spec.Collate

	return tenantCurrentStatus, nil
}

func (m *OBTenantManager) BuildPoolStatusList(gvTenant *model.Tenant) ([]v1alpha1.ResourcePoolStatus, error) {

	var poolStatusList []v1alpha1.ResourcePoolStatus
	var locality string
	var primaryZone string

	locality = gvTenant.Locality
	primaryZone = gvTenant.PrimaryZone
	typeMap := GenerateTypeMap(locality)
	tenantID := gvTenant.TenantID
	priorityMap := GeneratePriorityMap(primaryZone)
	unitNumMap, err := m.GenerateStatusUnitNumMap(m.OBTenant.Spec.Pools)
	if err != nil {
		return poolStatusList, err
	}
	zoneList, err := m.GenerateStatusZone(tenantID)
	if err != nil {
		return poolStatusList, err
	}
	for _, zone := range zoneList {
		var tenantCurrentStatus v1alpha1.ResourcePoolStatus
		tenantCurrentStatus.ZoneList = zone
		tenantCurrentStatus.Type = typeMap[zone]
		tenantCurrentStatus.UnitNumber = unitNumMap[zone]
		tenantCurrentStatus.Priority = priorityMap[zone]
		tenantCurrentStatus.UnitConfig, err = m.BuildUnitConfigV4FromDB(zone, tenantID)
		if err != nil {
			return poolStatusList, err
		}
		tenantCurrentStatus.Units, err = m.BuildUnitStatusFromDB(zone, tenantID)
		if err != nil {
			return poolStatusList, err
		}
		poolStatusList = append(poolStatusList, tenantCurrentStatus)
	}
	return poolStatusList, nil
}

func (m *OBTenantManager) GenerateStatusZone(tenantID int64) ([]string, error) {
	var zoneList []string
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
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
	for k, _ := range zoneMap {
		zoneList = append(zoneList, k)
	}
	return zoneList, nil
}



func (m *OBTenantManager) GenerateStatusUnitNumMap(zones []v1alpha1.ResourcePoolSpec) (map[string]int, error) {
	unitNumMap := make(map[string]int, 0)
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return unitNumMap, errors.Wrap(err, "Get Sql Operator Error When Building Resource Unit From DB")
	}
	poolList, err := oceanbaseOperationManager.GetPoolList()
	if err != nil {
		return unitNumMap, errors.Wrap(err, "Get sql error when get pool list")
	}
	for _, zone := range zones {
		poolName := m.GeneratePoolName(zone.ZoneList)
		for _, pool := range poolList {
			if pool.Name == poolName {
				unitNumMap[zone.ZoneList] = int(pool.UnitNum)
			}
		}
	}
	return unitNumMap, nil
}

func (m *OBTenantManager) BuildUnitConfigV4FromDB(zone string, tenantID int64) (v1alpha1.UnitConfig, error) {
	var unitConfig v1alpha1.UnitConfig
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return unitConfig, errors.Wrap(err, "Get Sql Operator Error When Building Resource Unit From DB")
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
						unitConfig.MaxCPU = kuberesource.MustParse(strconv.FormatFloat(unitConifg.MaxCPU, 'f', -1, 64))
						unitConfig.MinCPU = kuberesource.MustParse(strconv.FormatFloat(unitConifg.MinCPU, 'f', -1, 64))
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

func (m *OBTenantManager) BuildUnitStatusFromDB(zone string, tenantID int64) ([]v1alpha1.UnitStatus, error) {
	var unitList []v1alpha1.UnitStatus
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
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
