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
//go:generate task_register $GOFILE

package obcluster

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	cmdconst "github.com/oceanbase/ob-operator/internal/const/cmd"
	obagentconst "github.com/oceanbase/ob-operator/internal/const/obagent"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/obzone"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/helper"
	"github.com/oceanbase/ob-operator/pkg/helper/converter"
	helpermodel "github.com/oceanbase/ob-operator/pkg/helper/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/param"
	obutil "github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/util"
	"github.com/oceanbase/ob-operator/pkg/task/builder"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

var taskMap = builder.NewTaskHub[*OBClusterManager]()

func WaitOBZoneTopologyMatch(_ *OBClusterManager) tasktypes.TaskError {
	// TODO
	return nil
}

func WaitOBZoneDeleted(m *OBClusterManager) tasktypes.TaskError {
	waitSuccess := false
	for i := 1; i < obcfg.GetConfig().Time.ServerDeleteTimeoutSeconds; i++ {
		obcluster, err := m.getOBCluster()
		if err != nil {
			return errors.Wrap(err, "get obcluster failed")
		}
		zoneDeleted := true
		for _, zoneStatus := range obcluster.Status.OBZoneStatus {
			found := false
			for _, zone := range m.OBCluster.Spec.Topology {
				if zoneStatus.Zone == m.generateZoneName(zone.Zone) {
					found = true
					break
				}
			}
			if !found {
				m.Logger.Info("OBZone not in spec, still not deleted", "zone", zoneStatus.Zone)
				zoneDeleted = false
				break
			}
		}
		if zoneDeleted {
			m.Logger.V(oceanbaseconst.LogLevelTrace).Info("All zone deleted")
			waitSuccess = true
			break
		}
		time.Sleep(time.Second * 1)
	}
	if waitSuccess {
		return nil
	}
	return errors.Errorf("OBCluster %s zone still not deleted when timeout", m.OBCluster.Name)
}

func ModifyOBZoneReplica(m *OBClusterManager) tasktypes.TaskError {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		obzoneList, err := m.listOBZones()
		if err != nil {
			m.Logger.Error(err, "List obzone failed")
			return errors.Wrapf(err, "List obzone of obcluster %s failed", m.OBCluster.Name)
		}
		for _, zone := range m.OBCluster.Spec.Topology {
			for _, obzone := range obzoneList.Items {
				if zone.Zone == obzone.Spec.Topology.Zone && zone.Replica != obzone.Spec.Topology.Replica {
					m.Logger.Info("Modify obzone replica", "obzone", zone.Zone)
					obzone.Spec.Topology.Replica = zone.Replica
					err = m.Client.Update(m.Ctx, &obzone)
					if err != nil {
						return errors.Wrapf(err, "Modify obzone %s replica failed", zone.Zone)
					}
				}
			}
		}
		return nil
	})
}

func DeleteOBZone(m *OBClusterManager) tasktypes.TaskError {
	zonesToDelete, err := m.getZonesToDelete()
	if err != nil {
		return errors.Wrap(err, "Failed to get obzones to delete")
	}
	for _, zone := range zonesToDelete {
		err = m.Client.Delete(m.Ctx, &zone)
		if err != nil {
			return errors.Wrapf(err, "Delete obzone %s", zone.Name)
		}
		m.Recorder.Event(m.OBCluster, "DeleteOBZone", "", fmt.Sprintf("Delete obzone %s successfully", zone.Name))
	}
	return nil
}

func CreateOBZone(m *OBClusterManager) tasktypes.TaskError {
	m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Create obzones")
	blockOwnerDeletion := true
	ownerReferenceList := make([]metav1.OwnerReference, 0)
	ownerReference := metav1.OwnerReference{
		APIVersion:         m.OBCluster.APIVersion,
		Kind:               m.OBCluster.Kind,
		Name:               m.OBCluster.Name,
		UID:                m.OBCluster.GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
	}
	ownerReferenceList = append(ownerReferenceList, ownerReference)
	independentVolumeAnnoVal, independentVolumeAnnoExist := resourceutils.GetAnnotationField(m.OBCluster, oceanbaseconst.AnnotationsIndependentPVCLifecycle)
	singlePVCAnnoVal, singlePVCAnnoExist := resourceutils.GetAnnotationField(m.OBCluster, oceanbaseconst.AnnotationsSinglePVC)
	modeAnnoVal, modeAnnoExist := resourceutils.GetAnnotationField(m.OBCluster, oceanbaseconst.AnnotationsMode)
	migrateAnnoVal, migrateAnnoExist := resourceutils.GetAnnotationField(m.OBCluster, oceanbaseconst.AnnotationsSourceClusterAddress)
	for _, zone := range m.OBCluster.Spec.Topology {
		zoneName := m.generateZoneName(zone.Zone)
		zoneExists := false
		for _, zoneStatus := range m.OBCluster.Status.OBZoneStatus {
			if zoneName == zoneStatus.Zone {
				zoneExists = true
				break
			}
		}
		if zoneExists {
			m.Logger.Info("Zone already exists", "zone", zoneName)
			continue
		}
		labels := make(map[string]string)
		labels[oceanbaseconst.LabelRefUID] = string(m.OBCluster.GetUID())
		labels[oceanbaseconst.LabelRefOBCluster] = m.OBCluster.Name
		finalizerName := oceanbaseconst.FinalizerDeleteOBZone
		finalizers := []string{finalizerName}
		obzone := &v1alpha1.OBZone{
			ObjectMeta: metav1.ObjectMeta{
				Name:            zoneName,
				Namespace:       m.OBCluster.Namespace,
				OwnerReferences: ownerReferenceList,
				Labels:          labels,
				Finalizers:      finalizers,
			},
			Spec: v1alpha1.OBZoneSpec{
				ClusterName:      m.OBCluster.Spec.ClusterName,
				ClusterId:        m.OBCluster.Spec.ClusterId,
				OBServerTemplate: m.OBCluster.Spec.OBServerTemplate,
				MonitorTemplate:  m.OBCluster.Spec.MonitorTemplate,
				BackupVolume:     m.OBCluster.Spec.BackupVolume,
				Topology:         zone,
				ServiceAccount:   m.OBCluster.Spec.ServiceAccount,
			},
		}
		obzone.ObjectMeta.Annotations = make(map[string]string)
		if independentVolumeAnnoExist {
			obzone.ObjectMeta.Annotations[oceanbaseconst.AnnotationsIndependentPVCLifecycle] = independentVolumeAnnoVal
		}
		if singlePVCAnnoExist {
			obzone.ObjectMeta.Annotations[oceanbaseconst.AnnotationsSinglePVC] = singlePVCAnnoVal
		}
		if modeAnnoExist {
			obzone.ObjectMeta.Annotations[oceanbaseconst.AnnotationsMode] = modeAnnoVal
		}
		if migrateAnnoExist {
			obzone.ObjectMeta.Annotations[oceanbaseconst.AnnotationsSourceClusterAddress] = migrateAnnoVal
		}
		m.Logger.Info("Create obzone", "zone", zoneName)
		err := m.Client.Create(m.Ctx, obzone)
		if err != nil {
			m.Logger.Error(err, "Failed to create obzone", "zone", zone.Zone)
			return errors.Wrap(err, "create obzone")
		}
		m.Recorder.Event(m.OBCluster, "CreateOBZone", "", fmt.Sprintf("Create obzone %s successfully", zoneName))
	}
	return nil
}

func Bootstrap(m *OBClusterManager) tasktypes.TaskError {
	obzoneList, err := m.listOBZones()
	if err != nil {
		m.Logger.Error(err, "list obzones failed")
		return errors.Wrap(err, "list obzones")
	}
	m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Successfully get obzone list", "obzone list", obzoneList)
	if len(obzoneList.Items) == 0 {
		return errors.Wrap(err, "no obzone belongs to this cluster")
	}
	var manager *operation.OceanbaseOperationManager
	for i := 0; i < obcfg.GetConfig().Time.GetConnectionMaxRetries; i++ {
		manager, err = m.getOceanbaseOperationManager()
		if err != nil || manager == nil {
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Get oceanbase operation manager failed")
			time.Sleep(time.Second * time.Duration(obcfg.GetConfig().Time.CheckConnectionInterval))
		} else {
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Successfully got oceanbase operation manager")
			break
		}
	}
	if err != nil {
		m.Logger.Error(err, "get oceanbase operation manager failed")
		return errors.Wrap(err, "get oceanbase operation manager")
	}

	modeAnnoVal, modeAnnoExist := resourceutils.GetAnnotationField(m.OBCluster, oceanbaseconst.AnnotationsMode)

	bootstrapServers := make([]model.BootstrapServerInfo, 0, len(m.OBCluster.Spec.Topology))
	if modeAnnoExist && modeAnnoVal == oceanbaseconst.ModeStandalone {
		m.Logger.Info("Bootstrap as standalone mode")
		bootstrapServers = append(bootstrapServers, model.BootstrapServerInfo{
			Zone: m.OBCluster.Spec.Topology[0].Zone,
			Server: &model.ServerInfo{
				Ip:   "127.0.0.1",
				Port: oceanbaseconst.RpcPort,
			},
		})
	} else {
		connectAddress := manager.Connector.DataSource().GetAddress()
		for _, zone := range obzoneList.Items {
			serverIp := zone.Status.OBServerStatus[0].GetConnectAddr()
			// Notes: If the addr of the db connector is in this obzone, use it as the bootstrap server instead of the first one
			for _, serverInfo := range zone.Status.OBServerStatus {
				if serverInfo.Server == connectAddress || serverInfo.ServiceIP == connectAddress {
					serverIp = serverInfo.GetConnectAddr()
				}
			}
			serverInfo := &model.ServerInfo{
				Ip:   serverIp,
				Port: oceanbaseconst.RpcPort,
			}
			bootstrapServers = append(bootstrapServers, model.BootstrapServerInfo{
				Zone:   zone.Spec.Topology.Zone,
				Server: serverInfo,
			})
		}
	}

	err = manager.Bootstrap(m.Ctx, bootstrapServers)
	if err != nil {
		m.Logger.Error(err, "bootstrap failed")
	} else {
		m.Recorder.Event(m.OBCluster, "Bootstrap", "", "Bootstrap successfully")
	}
	return err
}

// Use Or for compatibility
func CreateUsers(m *OBClusterManager) tasktypes.TaskError {
	err := m.createUser(oceanbaseconst.RootUser, m.OBCluster.Spec.UserSecrets.Root, oceanbaseconst.AllPrivilege)
	if err != nil {
		return errors.Wrap(err, "Create root user")
	}
	err = m.createUser(oceanbaseconst.OperatorUser, m.OBCluster.Spec.UserSecrets.Operator, oceanbaseconst.AllPrivilege)
	if err != nil {
		return errors.Wrap(err, "Create operator user")
	}
	err = m.createUser(obagentconst.MonitorUser, m.OBCluster.Spec.UserSecrets.Monitor, oceanbaseconst.SelectPrivilege)
	if err != nil {
		return errors.Wrap(err, "Create monitor user")
	}
	err = m.createUser(oceanbaseconst.ProxyUser, m.OBCluster.Spec.UserSecrets.ProxyRO, oceanbaseconst.SelectPrivilege)
	if err != nil {
		return errors.Wrap(err, "Create proxyro user")
	}
	return nil
}

func MaintainOBParameter(m *OBClusterManager) tasktypes.TaskError {
	parameterMap := make(map[string]apitypes.Parameter)
	for _, parameter := range m.OBCluster.Status.Parameters {
		m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Build parameter map", "parameter", parameter.Name)
		parameterMap[parameter.Name] = parameter
	}
	for _, parameter := range m.OBCluster.Spec.Parameters {
		parameterStatus, parameterExists := parameterMap[parameter.Name]
		if !parameterExists {
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Parameter not exists, need create", "param", parameter.Name)
			err := m.CreateOBParameter(&parameter)
			if err != nil {
				// since parameter is not a big problem, just log the error
				m.Logger.Error(err, "Create obparameter failed", "param", parameter.Name)
			}
		} else if parameterStatus.Value != parameter.Value {
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Parameter value not matched, need update", "param", parameter.Name)
			err := m.UpdateOBParameter(&parameter)
			if err != nil {
				// since parameter is not a big problem, just log the error
				m.Logger.Error(err, "Update obparameter failed", "param", parameter.Name)
			}
		}
		m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Remove parameter from map", "parameter", parameter.Name)
		delete(parameterMap, parameter.Name)
	}

	// delete parameters that not in spec definition
	for _, parameter := range parameterMap {
		m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Delete parameter", "parameter", parameter.Name)
		err := m.DeleteOBParameter(&parameter)
		if err != nil {
			m.Logger.Error(err, "Failed to delete parameter")
		}
	}
	return nil
}

func ValidateUpgradeInfo(m *OBClusterManager) tasktypes.TaskError {
	// Get current obcluster version
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrapf(err, "Failed to get operation manager of obcluster %s", m.OBCluster.Name)
	}
	// version, err := oceanbaseOperationManager.GetVersion()
	version, err := oceanbaseOperationManager.GetVersion(m.Ctx)
	if err != nil {
		return errors.Wrapf(err, "Failed to get version of obcluster %s", m.OBCluster.Name)
	}

	_, _, err = resourceutils.RunJob(m.Ctx, m.Client, m.Logger, m.OBCluster.Namespace,
		fmt.Sprintf("%s-upgrade-validate", m.OBCluster.Name),
		m.OBCluster.Spec.OBServerTemplate.Image,
		m.OBCluster.Spec.OBServerTemplate.PodFields,
		fmt.Sprintf(oceanbaseconst.CmdUpgradeValidateTemplate, version.String()))
	if err != nil {
		return errors.Wrap(err, "Upgrade is unsupported")
	}
	return nil
}

func UpgradeCheck(m *OBClusterManager) tasktypes.TaskError {
	return resourceutils.ExecuteUpgradeScript(m.Ctx, m.Client, m.Logger, m.OBCluster, oceanbaseconst.UpgradeCheckerScriptPath, "")
}

func BackupEssentialParameters(m *OBClusterManager) tasktypes.TaskError {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrapf(err, "Failed to get operation manager of obcluster %s", m.OBCluster.Name)
	}
	essentialParameters := make([]model.Parameter, 0)
	for _, parameter := range oceanbaseconst.UpgradeEssentialParameters {
		parameterValues, err := oceanbaseOperationManager.GetParameter(m.Ctx, parameter, nil)
		if err != nil {
			return errors.Wrapf(err, "Failed to get parameter %s", parameter)
		}
		essentialParameters = append(essentialParameters, parameterValues...)
	}

	contextMap := make(map[string]string)
	jsonContent, err := json.Marshal(essentialParameters)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal essential parameters")
	}
	contextMap[oceanbaseconst.EssentialParametersKey] = string(jsonContent)
	contextObjectName := fmt.Sprintf("%s-%d-%s", m.OBCluster.Spec.ClusterName, m.OBCluster.Spec.ClusterId, oceanbaseconst.EssentialParametersKey)
	contextSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      contextObjectName,
			Namespace: m.OBCluster.Namespace,
		},
		Type:       "Opaque",
		StringData: contextMap,
	}
	err = m.Client.Create(m.Ctx, contextSecret)
	if err != nil {
		return errors.Wrap(err, "Create context secret object")
	}
	return nil
}

func BeginUpgrade(m *OBClusterManager) tasktypes.TaskError {
	return resourceutils.ExecuteUpgradeScript(m.Ctx, m.Client, m.Logger, m.OBCluster, oceanbaseconst.UpgradePreScriptPath, "")
}

func RollingUpgradeByZone(m *OBClusterManager) tasktypes.TaskError {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		zones, err := m.listOBZones()
		if err != nil {
			return errors.Wrap(err, "Failed to get obzone list")
		}
		for _, zone := range zones.Items {
			// update image and tag
			zone.Spec.OBServerTemplate.Image = m.OBCluster.Spec.OBServerTemplate.Image
			err = m.Client.Update(m.Ctx, &zone)
			if err != nil {
				return errors.Wrap(err, "Failed to update obzone image")
			}
			err = m.WaitOBZoneUpgradeFinished(zone.Name)
			if err != nil {
				return errors.Wrapf(err, "Wait obzone %s upgrade finish failed", zone.Name)
			}
		}
		return nil
	})
}

func FinishUpgrade(m *OBClusterManager) tasktypes.TaskError {
	return resourceutils.ExecuteUpgradeScript(m.Ctx, m.Client, m.Logger, m.OBCluster, oceanbaseconst.UpgradePostScriptPath, "")
}

func ModifySysTenantReplica(m *OBClusterManager) tasktypes.TaskError {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrapf(err, "Failed to get operation manager of obcluster %s", m.OBCluster.Name)
	}
	desiredZones := make([]string, 0)
	for _, obzone := range m.OBCluster.Spec.Topology {
		desiredZones = append(desiredZones, obzone.Zone)
	}
	// add zone to pool zone list
	sysPool, err := oceanbaseOperationManager.GetPoolByName(m.Ctx, oceanbaseconst.SysTenantPool)
	if err != nil {
		return errors.Wrap(err, "Failed to get sys pool info")
	}
	zoneList := strings.Split(sysPool.ZoneList, ";")
	for _, zone := range desiredZones {
		found := false
		for _, z := range zoneList {
			if zone == z {
				found = true
				break
			}
		}
		if !found {
			zoneList = append(zoneList, zone)
		}
	}
	m.Logger.Info("Modify sys pool's zone list", "zone list", zoneList)
	err = oceanbaseOperationManager.AlterPool(m.Ctx, &model.PoolParam{
		PoolName: oceanbaseconst.SysTenantPool,
		ZoneList: zoneList,
	})
	if err != nil {
		return errors.Wrapf(err, "Failed to modify sys pool's zone list to  %v", zoneList)
	}
	// add locality one by one
	sysTenant, err := oceanbaseOperationManager.GetTenantByName(m.Ctx, oceanbaseconst.SysTenant)
	if err != nil {
		return errors.Wrap(err, "Failed to get sys tenant info")
	}
	locality := sysTenant.Locality
	replicas := obutil.ConvertFromLocalityStr(locality)
	for _, zone := range desiredZones {
		found := false
		for _, r := range replicas {
			if zone == r.Zone {
				found = true
				break
			}
		}
		if !found {
			replicas = append(replicas, model.Replica{
				Type: oceanbaseconst.FullType,
				Num:  1,
				Zone: zone,
			})
			locality = obutil.ConvertToLocalityStr(replicas)
			m.Logger.Info("Modify sys tenant's locality when add zone", "locality", locality)
			err = oceanbaseOperationManager.SetTenant(m.Ctx, model.TenantSQLParam{
				TenantName: oceanbaseconst.SysTenant,
				Locality:   locality,
			})
			if err != nil {
				return errors.Wrapf(err, "Failed to set sys locality to %s", locality)
			}
			err = oceanbaseOperationManager.WaitTenantLocalityChangeFinished(m.Ctx, oceanbaseconst.SysTenant, obcfg.GetConfig().Time.LocalityChangeTimeoutSeconds)
			if err != nil {
				return errors.Wrapf(err, "Locality change to %s not finished after timeout", locality)
			}
		}
	}
	// delete locality one by one
	for _, r := range replicas {
		found := false
		for _, zone := range desiredZones {
			if zone == r.Zone {
				found = true
				break
			}
		}
		if !found {
			newReplicas := obutil.OmitZoneFromReplicas(replicas, r.Zone)
			locality = obutil.ConvertToLocalityStr(newReplicas)
			m.Logger.Info("Modify sys tenant's locality when deleting zone", "locality", locality)
			err = oceanbaseOperationManager.SetTenant(m.Ctx, model.TenantSQLParam{
				TenantName: oceanbaseconst.SysTenant,
				Locality:   locality,
			})
			if err != nil {
				return errors.Wrapf(err, "Failed to set sys locality to %s", locality)
			}
			err = oceanbaseOperationManager.WaitTenantLocalityChangeFinished(m.Ctx, oceanbaseconst.SysTenant, obcfg.GetConfig().Time.LocalityChangeTimeoutSeconds)
			if err != nil {
				return errors.Wrapf(err, "Locality change to %s not finished after timeout", locality)
			}
		}
	}
	// delete zone from pool zone list
	newZoneList := make([]string, 0)
	for _, zone := range zoneList {
		found := false
		for _, z := range desiredZones {
			if zone == z {
				found = true
				break
			}
		}
		if found {
			newZoneList = append(newZoneList, zone)
		}
	}
	m.Logger.Info("Modify sys pool's zone list when delete zone", "zone list", newZoneList)
	return oceanbaseOperationManager.AlterPool(m.Ctx, &model.PoolParam{
		PoolName: oceanbaseconst.SysTenantPool,
		ZoneList: newZoneList,
	})
}

func CreateServiceForMonitor(m *OBClusterManager) tasktypes.TaskError {
	ownerReferenceList := make([]metav1.OwnerReference, 0)
	ownerReference := metav1.OwnerReference{
		APIVersion: m.OBCluster.APIVersion,
		Kind:       m.OBCluster.Kind,
		Name:       m.OBCluster.Name,
		UID:        m.OBCluster.GetUID(),
	}
	ownerReferenceList = append(ownerReferenceList, ownerReference)
	selector := make(map[string]string)
	selector[oceanbaseconst.LabelRefOBCluster] = m.OBCluster.Name
	monitorService := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       m.OBCluster.Namespace,
			Name:            fmt.Sprintf("svc-monitor-%s-%s", m.OBCluster.Name, rand.String(6)),
			OwnerReferences: ownerReferenceList,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       obagentconst.HttpPortName,
					Protocol:   corev1.ProtocolTCP,
					Port:       obagentconst.HttpPort,
					TargetPort: intstr.FromInt(obagentconst.HttpPort),
				},
			},
			Selector: selector,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
	err := m.Client.Create(m.Ctx, &monitorService)
	if err != nil {
		return errors.Wrap(err, "Create monitor service")
	}
	m.Recorder.Event(m.OBCluster, "MaintainedAfterBootstrap", "", "Create monitor service successfully")
	return nil
}

func RestoreEssentialParameters(m *OBClusterManager) tasktypes.TaskError {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrapf(err, "Failed to get operation manager of obcluster %s", m.OBCluster.Name)
	}
	essentialParameters := make([]model.Parameter, 0)

	contextObjectName := fmt.Sprintf("%s-%d-%s", m.OBCluster.Spec.ClusterName, m.OBCluster.Spec.ClusterId, oceanbaseconst.EssentialParametersKey)
	contextSecret := &corev1.Secret{}
	err = m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.OBCluster.Namespace,
		Name:      contextObjectName,
	}, contextSecret)
	if err != nil {
		m.Logger.Error(err, "Failed to get context secret")
		// parameter can be set manually, just return here and emit an event
		m.Recorder.Event(m.OBCluster, "Warning", "Restore essential parameters failed", err.Error())
		return nil
	}

	encodedParameters := string(contextSecret.Data[oceanbaseconst.EssentialParametersKey])
	m.Logger.Info("Get encoded parameters", "parameters", encodedParameters)
	err = json.Unmarshal([]byte(encodedParameters), &essentialParameters)
	if err != nil {
		return errors.New("Parse encoded parameters failed")
	}

	for _, parameter := range essentialParameters {
		err = oceanbaseOperationManager.SetParameter(m.Ctx, parameter.Name, parameter.Value, &param.Scope{
			Name:  "server",
			Value: fmt.Sprintf("%s:%d", parameter.SvrIp, parameter.SvrPort),
		})
		if err != nil {
			return errors.Wrapf(err, "Failed to set parameter %s to %s:%d", parameter.Name, parameter.SvrIp, parameter.SvrPort)
		}
	}
	_ = m.Client.Delete(m.Ctx, contextSecret)
	m.Recorder.Event(m.OBCluster, "Upgrade", "", "Restore essential parameters successfully")
	return nil
}

func CheckAndCreateUserSecrets(m *OBClusterManager) tasktypes.TaskError {
	secretList := []string{
		m.OBCluster.Spec.UserSecrets.Operator,
		m.OBCluster.Spec.UserSecrets.Monitor,
		m.OBCluster.Spec.UserSecrets.ProxyRO,
	}
	for _, secret := range secretList {
		fetchedSec := &corev1.Secret{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.OBCluster.Namespace,
			Name:      secret,
		}, fetchedSec)
		if err != nil {
			if kubeerrors.IsNotFound(err) {
				err := m.Client.Create(m.Ctx, &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      secret,
						Namespace: m.OBCluster.Namespace,
					},
					StringData: map[string]string{
						"password": rand.String(16),
					},
				})
				if err != nil {
					return errors.Wrap(err, "Create secret "+secret)
				}
			}
		}
	}
	return nil
}

func CreateOBClusterService(m *OBClusterManager) tasktypes.TaskError {
	modeAnnoVal, modeAnnoExist := resourceutils.GetAnnotationField(m.OBCluster, oceanbaseconst.AnnotationsMode)
	if modeAnnoExist && modeAnnoVal == oceanbaseconst.ModeStandalone {
		err := m.Client.Create(m.Ctx, &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      m.OBCluster.GetName() + "-standalone-svc",
				Namespace: m.OBCluster.GetNamespace(),
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: m.OBCluster.APIVersion,
					Kind:       m.OBCluster.Kind,
					Name:       m.OBCluster.GetName(),
					UID:        m.OBCluster.GetUID(),
				}},
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{
					Name:       "sql",
					Protocol:   corev1.ProtocolTCP,
					Port:       2881,
					TargetPort: intstr.IntOrString{IntVal: 2881},
				}},
				Selector: map[string]string{
					oceanbaseconst.LabelRefOBCluster: m.OBCluster.GetName(),
				},
				Type: corev1.ServiceTypeNodePort,
			},
		})
		if err != nil {
			m.Recorder.Event(m.OBCluster, "Warning", "Create standalone service failed", err.Error())
			return errors.Wrap(err, "Create service")
		}
	}
	return nil
}

func CheckImageReady(m *OBClusterManager) tasktypes.TaskError {
	jobName := "image-pull-ready-" + rand.String(8)
	var ttl int32 = 120
	var backoffLimit int32 = 32
	checkImagePullJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: m.OBCluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{{
				Kind:       m.OBCluster.Kind,
				APIVersion: m.OBCluster.APIVersion,
				Name:       m.OBCluster.Name,
				UID:        m.OBCluster.UID,
			}},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            "helper-check-image-pull-ready",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Image:           m.OBCluster.Spec.OBServerTemplate.Image,
						Command:         []string{"bash", "-c", "/home/admin/oceanbase/bin/oceanbase-helper help"},
					}},
					RestartPolicy: corev1.RestartPolicyNever,
					SchedulerName: resourceutils.GetSchedulerName(m.OBCluster.Spec.OBServerTemplate.PodFields),
				},
			},
			TTLSecondsAfterFinished: &ttl,
			BackoffLimit:            &backoffLimit,
		},
	}
	err := m.Client.Create(m.Ctx, checkImagePullJob)
	if err != nil {
		return errors.Wrap(err, "Create check image pull job")
	}
	m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Create check image pull job", "job", jobName)

	imagePullReady := false
	var checkImagePullReadyMaxTimes = 8000
	labelSelector := metav1.FormatLabelSelector(checkImagePullJob.Spec.Selector)
	selector, err := labels.Parse(labelSelector)
	if err != nil {
		return errors.Wrap(err, "Parse label selector")
	}
outerLoop:
	for i := 0; i < checkImagePullReadyMaxTimes; i++ {
		podList := &corev1.PodList{}
		err = m.Client.List(m.Ctx, podList, &client.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			return errors.Wrap(err, "List pods")
		}
		if len(podList.Items) == 0 {
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("No pod found for check image pull job")
			time.Sleep(time.Second * time.Duration(obcfg.GetConfig().Time.CheckJobInterval))
			continue
		}
		pod := podList.Items[0]
		switch pod.Status.Phase {
		case corev1.PodFailed:
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Check image pull job failed")
			return errors.New("Check image pull job failed")
		case corev1.PodSucceeded, corev1.PodRunning:
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Check image pull job finished")
			imagePullReady = true
			break outerLoop
		case corev1.PodPending, corev1.PodUnknown:
			// if every container has pulled its image, break outer loop
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.State.Waiting != nil {
					switch containerStatus.State.Waiting.Reason {
					case "ErrImagePull", "ImagePullBackOff":
						m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Wait to pull image", "reason", containerStatus.State.Waiting.Reason, "message", containerStatus.State.Waiting.Message)
					default:
						m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Container is waiting", "reason", containerStatus.State.Waiting.Reason, "message", containerStatus.State.Waiting.Message)
					}
					time.Sleep(time.Second * time.Duration(obcfg.GetConfig().Time.CheckJobInterval))
					continue outerLoop
				} else if containerStatus.State.Running != nil || containerStatus.State.Terminated != nil {
					m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Container is running or terminated")
				}
			}
			imagePullReady = true
			break outerLoop
		}
	}
	if !imagePullReady {
		return errors.New("Image pull not ready")
	}
	return nil
}

func CheckClusterMode(m *OBClusterManager) tasktypes.TaskError {
	modeAnnoVal, modeAnnoExist := resourceutils.GetAnnotationField(m.OBCluster, oceanbaseconst.AnnotationsMode)
	if modeAnnoExist {
		versionOutput, _, err := resourceutils.RunJob(m.Ctx, m.Client, m.Logger, m.OBCluster.Namespace,
			m.OBCluster.Name+"-get-version",
			m.OBCluster.Spec.OBServerTemplate.Image,
			m.OBCluster.Spec.OBServerTemplate.PodFields,
			"/home/admin/oceanbase/bin/oceanbase-helper version",
		)
		if err != nil {
			// Make compatible with legacy version in which there is no `version` command
			var code int32
			switch modeAnnoVal {
			case oceanbaseconst.ModeStandalone:
				_, code, err = resourceutils.RunJob(m.Ctx, m.Client, m.Logger, m.OBCluster.Namespace,
					m.OBCluster.Name+"-standalone-validate",
					m.OBCluster.Spec.OBServerTemplate.Image,
					m.OBCluster.Spec.OBServerTemplate.PodFields,
					"/home/admin/oceanbase/bin/oceanbase-helper standalone validate",
				)
			case oceanbaseconst.ModeService:
				_, code, err = resourceutils.RunJob(m.Ctx, m.Client, m.Logger, m.OBCluster.Namespace,
					m.OBCluster.Name+"-service-validate",
					m.OBCluster.Spec.OBServerTemplate.Image,
					m.OBCluster.Spec.OBServerTemplate.PodFields,
					"/home/admin/oceanbase/bin/oceanbase-helper service validate",
				)
			}
			if err != nil && code > 1 {
				return errors.Wrap(err, "Failed to run service mode validate job")
			}
			return nil
		}
		var version string
		lines := strings.Split(versionOutput, "\n")
		if len(lines) == 0 {
			m.Logger.Info("Get version failed")
			return nil
		}
		if len(lines) > 3 {
			versionStr := strings.Split(lines[1], " ")
			semVer := versionStr[len(versionStr)-1]
			releaseStr := strings.Split(strings.Split(lines[3], " ")[1], "-")[0]
			version = fmt.Sprintf("%s-%s", semVer[0:len(semVer)-1], releaseStr)
		} else {
			version = strings.TrimSpace(lines[0])
		}
		currentVersion, err := helper.ParseOceanBaseVersion(version)
		if err != nil {
			m.Logger.WithValues("version", version, "err", err.Error()).Info("Failed to parse current version")
			return nil
		}
		switch modeAnnoVal {
		case oceanbaseconst.ModeStandalone:
			standaloneVersion, _ := helper.ParseOceanBaseVersion(oceanbaseconst.StandaloneMinVersion)
			if currentVersion.Cmp(standaloneVersion) < 0 {
				return errors.Errorf("Current version is lower than %s, does not support standalone mode", oceanbaseconst.StandaloneMinVersion)
			}
		case oceanbaseconst.ModeService:
			if strings.HasPrefix(version, oceanbaseconst.ServiceExcludeVersion) {
				return errors.Errorf("Current version is %s. 4.2.2.x does not support service mode", version)
			}
			requiredVersion, _ := helper.ParseOceanBaseVersion(oceanbaseconst.ServiceMinVersion)
			if currentVersion.Cmp(requiredVersion) < 0 {
				return errors.Errorf("Current version is lower than %s, does not support service mode", oceanbaseconst.ServiceMinVersion)
			}
		}
		m.Logger.Info("Run service mode validate job successfully", "version", version, "mode", modeAnnoVal)
	}
	return nil
}

func CheckMigration(m *OBClusterManager) tasktypes.TaskError {
	m.Logger.Info("Check before migration")
	manager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, "get operation manager")
	}

	// check version strictly matches
	targetVersionStr, _, err := resourceutils.RunJob(m.Ctx, m.Client, m.Logger, m.OBCluster.Namespace,
		fmt.Sprintf("%s-version", m.OBCluster.Name),
		m.OBCluster.Spec.OBServerTemplate.Image,
		m.OBCluster.Spec.OBServerTemplate.PodFields,
		oceanbaseconst.CmdVersion)
	if err != nil {
		return errors.Wrap(err, "get target oceanbase version")
	}

	sourceVersion, err := manager.GetVersion(m.Ctx)
	if err != nil {
		return errors.Wrap(err, "get source oceanbase version")
	}

	if sourceVersion.String() != targetVersionStr {
		return errors.Errorf("version mismatch source cluster: %s, target cluster: %s", sourceVersion.String(), targetVersionStr)
	}

	// check obzone matches topology
	obzoneList, err := manager.ListZones(m.Ctx)
	if err != nil {
		return errors.Wrap(err, "list obzones")
	}
	zoneMap := make(map[string]struct{})
	for _, zone := range obzoneList {
		zoneMap[zone.Name] = struct{}{}
	}

	extraZones := make([]string, 0)
	for _, obzone := range m.OBCluster.Spec.Topology {
		_, found := zoneMap[obzone.Zone]
		if !found {
			extraZones = append(extraZones, obzone.Zone)
		} else {
			delete(zoneMap, obzone.Zone)
		}
	}
	if len(extraZones) > 0 {
		return errors.Errorf("obzone %s defined but not in source cluster", strings.Join(extraZones, ","))
	}

	undefinedZones := make([]string, 0)
	for zone := range zoneMap {
		undefinedZones = append(undefinedZones, zone)
	}
	if len(undefinedZones) > 0 {
		return errors.Errorf("obzone %s not defined in obcluster's topology", strings.Join(undefinedZones, ","))
	}

	// check obcluster name and id
	obclusterNameParamList, err := manager.GetParameter(m.Ctx, oceanbaseconst.ClusterNameParam, nil)
	if err != nil {
		return errors.Wrap(err, "get obcluster name failed")
	}
	obclusterName := obclusterNameParamList[0].Value
	obclusterIdParamList, err := manager.GetParameter(m.Ctx, oceanbaseconst.ClusterIdParam, nil)
	if err != nil {
		return errors.Wrap(err, "get obcluster id failed")
	}
	obclusterId := obclusterIdParamList[0].Value
	if obclusterName != m.OBCluster.Spec.ClusterName {
		return errors.Errorf("Cluster name mismatch, source cluster: %s, current: %s", obclusterName, m.OBCluster.Spec.ClusterName)
	}
	if obclusterId != fmt.Sprintf("%d", m.OBCluster.Spec.ClusterId) {
		return errors.Errorf("Cluster id mismatch, source cluster: %s, current: %d", obclusterId, m.OBCluster.Spec.ClusterId)
	}
	return nil
}

func ScaleOBZonesVertically(m *OBClusterManager) tasktypes.TaskError {
	return m.rollingUpdateZones(m.changeZonesWhenScaling, zonestatus.ScaleVertically, zonestatus.Running, obcfg.GetConfig().Time.DefaultStateWaitTimeout)()
}

func ExpandPVC(m *OBClusterManager) tasktypes.TaskError {
	return m.modifyOBZonesAndCheckStatus(m.changeZonesWhenExpandingPVC, zonestatus.ExpandPVC, obcfg.GetConfig().Time.DefaultStateWaitTimeout)()
}

func ModifyServerTemplate(m *OBClusterManager) tasktypes.TaskError {
	return m.rollingUpdateZones(m.changeZonesWhenModifyingServerTemplate, zonestatus.ModifyServerTemplate, zonestatus.Running, obcfg.GetConfig().Time.DefaultStateWaitTimeout)()
}

func WaitOBZoneBootstrapReady(m *OBClusterManager) tasktypes.TaskError {
	return m.generateWaitOBZoneStatusFunc(zonestatus.BootstrapReady, obcfg.GetConfig().Time.DefaultStateWaitTimeout)()
}

func WaitOBZoneRunning(m *OBClusterManager) tasktypes.TaskError {
	return m.generateWaitOBZoneStatusFunc(zonestatus.Running, obcfg.GetConfig().Time.DefaultStateWaitTimeout)()
}

func RollingUpdateOBZones(m *OBClusterManager) tasktypes.TaskError {
	return m.rollingUpdateZones(m.changeZonesWhenUpdatingOBServers, zonestatus.RollingUpdateServers, zonestatus.Running, obcfg.GetConfig().Time.ServerDeleteTimeoutSeconds)()
}

func CheckEnvironment(m *OBClusterManager) tasktypes.TaskError {
	volumeName := m.OBCluster.Name + "check-clog-volume-" + rand.String(6)
	claimName := m.OBCluster.Name + "check-clog-claim-" + rand.String(6)
	jobName := m.OBCluster.Name + "-check-fs-" + rand.String(6)
	// Create PVC
	storageSpec := m.OBCluster.Spec.OBServerTemplate.Storage.RedoLogStorage
	requestsResources := corev1.ResourceList{}
	// Try fallocate to check if the filesystem meet the requirement.
	// The checker requires 4Mi space, we set the request to 64Mi for safety.
	requestsResources["storage"] = storageSpec.Size
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      claimName,
			Namespace: m.OBCluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: m.OBCluster.APIVersion,
				Kind:       m.OBCluster.Kind,
				Name:       m.OBCluster.Name,
				UID:        m.OBCluster.UID,
			}},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: requestsResources,
			},
			StorageClassName: &storageSpec.StorageClass,
		},
	}
	err := m.Client.Create(m.Ctx, pvc)
	if err != nil {
		return errors.Wrap(err, "Create pvc for checking storage")
	}
	defer func() {
		err = m.Client.Delete(m.Ctx, pvc)
		if err != nil {
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Failed to delete pvc for checking storage", "err", err)
		}
	}()
	// Assemble volumeConfigs
	volumeConfigs := resourceutils.JobContainerVolumes{
		Volumes: []corev1.Volume{{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: claimName,
				},
			},
		}},
		VolumeMounts: []corev1.VolumeMount{{
			Name:      volumeName,
			MountPath: oceanbaseconst.ClogPath,
		}},
	}
	_, exitCode, err := resourceutils.RunJob(
		m.Ctx, m.Client, m.Logger, m.OBCluster.Namespace,
		jobName,
		m.OBCluster.Spec.OBServerTemplate.Image,
		m.OBCluster.Spec.OBServerTemplate.PodFields,
		"/home/admin/oceanbase/bin/oceanbase-helper env-check storage "+oceanbaseconst.ClogPath,
		volumeConfigs,
	)
	// exit code 1 means the image version does not support the env-check command, just ignore it and try
	if err != nil && exitCode != 1 {
		return errors.Wrap(err, "Check filesystem")
	}
	return nil
}

func AnnotateOBCluster(m *OBClusterManager) tasktypes.TaskError {
	// Annotate obcluster with mode
	supportStaticIP := false
	mode := m.OBCluster.Annotations[oceanbaseconst.AnnotationsMode]
	withMode := mode == oceanbaseconst.ModeService || mode == oceanbaseconst.ModeStandalone

	if withMode {
		supportStaticIP = true
	} else {
		serverList := &v1alpha1.OBServerList{}
		err := m.Client.List(m.Ctx, serverList, client.MatchingLabels{oceanbaseconst.LabelRefOBCluster: m.OBCluster.Name}, client.InNamespace(m.OBCluster.Namespace))
		if err != nil {
			return errors.Wrap(err, "List servers of obcluster")
		}
		if len(serverList.Items) == 0 {
			return errors.New("No server found for obcluster")
		}
		for _, server := range serverList.Items {
			if server.Status.CNI != oceanbaseconst.CNIUnknown {
				supportStaticIP = true
				break
			}
		}
	}

	if supportStaticIP {
		copied := m.OBCluster.DeepCopy()
		if copied.Annotations == nil {
			copied.Annotations = make(map[string]string)
		}
		copied.Annotations[oceanbaseconst.AnnotationsSupportStaticIP] = "true"
		err := m.Client.Patch(m.Ctx, copied, client.MergeFrom(m.OBCluster))
		if err != nil {
			return errors.Wrap(err, "Patch obcluster")
		}
		zones, err := m.listOBZones()
		if err != nil {
			return errors.Wrap(err, "List obzones")
		}
		for _, zone := range zones.Items {
			copiedZone := zone.DeepCopy()
			if copiedZone.Annotations == nil {
				copiedZone.Annotations = make(map[string]string)
			}
			copiedZone.Annotations[oceanbaseconst.AnnotationsSupportStaticIP] = "true"
			err = m.Client.Patch(m.Ctx, copiedZone, client.MergeFrom(&zone))
			if err != nil {
				return errors.Wrap(err, "Patch obzone")
			}
		}
	}
	return nil
}

func OptimizeClusterByScenario(m *OBClusterManager) tasktypes.TaskError {
	// start a job to read optimize parameters, ignore errors, only proceed with valid outputs and ignore the errors
	m.Logger.Info("Start to optimize obcluster parameters")
	jobName := fmt.Sprintf("optimize-cluster-%s-%s", m.OBCluster.Name, rand.String(6))
	output, code, _ := resourceutils.RunJob(
		m.Ctx, m.Client, m.Logger, m.OBCluster.Namespace,
		jobName,
		m.OBCluster.Spec.OBServerTemplate.Image,
		m.OBCluster.Spec.OBServerTemplate.PodFields,
		fmt.Sprintf("bin/oceanbase-helper optimize cluster %s", m.OBCluster.Spec.Scenario))
	if code == int32(cmdconst.ExitCodeOK) || code == int32(cmdconst.ExitCodeIgnorableErr) {
		optimizeConfig := &helpermodel.OptimizationResponse{}
		err := json.Unmarshal([]byte(output), optimizeConfig)
		if err != nil {
			m.Logger.Error(err, "Failed to parse optimization config")
		}
		conn, err := m.getOceanbaseOperationManager()
		if err != nil {
			m.Logger.Error(err, "Get operation manager failed")
		}
		// obcluster only need to set parameters
		for _, parameter := range optimizeConfig.Parameters {
			m.Logger.Info("Set parameter %s to %s", parameter.Name, converter.ConvertFloat(parameter.Value))
			err := conn.SetParameter(m.Ctx, parameter.Name, converter.ConvertFloat(parameter.Value), nil)
			if err != nil {
				m.Logger.Error(err, "Failed to set parameter")
			}
		}
	}
	return nil
}

func AdjustParameters(m *OBClusterManager) tasktypes.TaskError {
	conn, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, "Get operation manager")
	}
	gvservers, err := conn.ListGVServers(m.Ctx)
	if err != nil {
		return errors.Wrap(err, "List gv servers")
	}
	zones, err := m.listOBZones()
	if err != nil {
		return errors.Wrap(err, "List obzones")
	}
	if len(zones.Items) == 0 {
		return errors.New("No obzone found")
	}

	oldResource := zones.Items[0].Spec.OBServerTemplate.Resource

	var maxAssignedCpu int64
	var maxAssignedMem int64
	var memoryLimitPercent float64
	for _, gvserver := range gvservers {
		if gvserver.MemAssigned > maxAssignedMem {
			if gvserver.MemoryLimit > oldResource.Memory.Value() {
				memoryLimitPercent = 0.9
			} else {
				memoryLimitPercent = float64(gvserver.MemoryLimit) / oldResource.Memory.AsApproximateFloat64()
			}
			maxAssignedMem = gvserver.MemAssigned
		}
		if gvserver.CPUAssigned > maxAssignedCpu {
			maxAssignedCpu = gvserver.CPUAssigned
		}
	}
	newResource := m.OBCluster.Spec.OBServerTemplate.Resource
	specMem := newResource.Memory.AsApproximateFloat64()
	specMemoryLimit := int64(specMem * memoryLimitPercent)

	targetMemoryLimit := max(specMemoryLimit, maxAssignedMem)
	m.Logger.V(oceanbaseconst.LogLevelDebug).
		Info("Adjust memory limit",
			"maxAssignedMem", maxAssignedMem,
			"specMem", specMem,
			"targetMemoryLimit", targetMemoryLimit,
			"percent", memoryLimitPercent,
		)

	copiedCluster := m.OBCluster.DeepCopy()

	foundMemoryLimit := false
	if newResource.Memory.Cmp(oldResource.Memory) != 0 {
		for i, p := range copiedCluster.Spec.Parameters {
			if p.Name == "memory_limit" {
				copiedCluster.Spec.Parameters[i].Value = fmt.Sprintf("%dM", targetMemoryLimit>>20)
				foundMemoryLimit = true
				break
			}
		}
		if !foundMemoryLimit {
			copiedCluster.Spec.Parameters = append(copiedCluster.Spec.Parameters, apitypes.Parameter{
				Name:  "memory_limit",
				Value: fmt.Sprintf("%dM", targetMemoryLimit>>20),
			})
		}
	}

	if oldResource.Cpu.Cmp(newResource.Cpu) != 0 {
		targetCpuCount := "16"
		if newResource.Cpu.Value() > 16 {
			targetCpuCount = newResource.Cpu.String()
		}
		foundCpuCount := false
		for i, p := range copiedCluster.Spec.Parameters {
			if p.Name == "cpu_count" {
				copiedCluster.Spec.Parameters[i].Value = targetCpuCount
				foundCpuCount = true
				break
			}
		}
		if !foundCpuCount {
			copiedCluster.Spec.Parameters = append(copiedCluster.Spec.Parameters, apitypes.Parameter{
				Name:  "cpu_count",
				Value: targetCpuCount,
			})
		}
	}

	err = m.Client.Patch(m.Ctx, copiedCluster, client.MergeFrom(m.OBCluster))
	if err != nil {
		m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Patch obcluster", "obcluster", m.OBCluster.Name)
		return errors.Wrap(err, "Patch obcluster")
	}
	return nil
}
