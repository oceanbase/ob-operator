package oceanbase

import (
	"fmt"
	"strings"

	logger "github.com/sirupsen/logrus"

	"github.com/google/uuid"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	clusterstatus "github.com/oceanbase/ob-operator/pkg/const/status/obcluster"
	"github.com/oceanbase/oceanbase-dashboard/internal/business/common"
	"github.com/oceanbase/oceanbase-dashboard/internal/business/constant"
	modelcommon "github.com/oceanbase/oceanbase-dashboard/internal/model/common"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/param"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/response"
	"github.com/oceanbase/oceanbase-dashboard/pkg/oceanbase"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	StatusDeleting  = "deleting"
	StatusRunning   = "running"
	StatusOperating = "operating"
)

func convertStatus(detailedStatus string) string {
	switch detailedStatus {
	case StatusRunning, StatusDeleting:
		return detailedStatus
	default:
		return StatusOperating
	}
}

func getStatisticStatus(obcluster *v1alpha1.OBCluster) string {
	if !obcluster.ObjectMeta.DeletionTimestamp.IsZero() {
		return StatusDeleting
	} else if obcluster.Status.Status == StatusRunning {
		return StatusRunning
	} else {
		return StatusOperating
	}
}

func buildOBClusterResponse(obcluster *v1alpha1.OBCluster) (*response.OBCluster, error) {
	obzoneList, err := oceanbase.ListOBZonesOfOBCluster(obcluster)
	if err != nil {
		return nil, errors.Wrapf(err, "List obzone of obcluster %s %s", obcluster.Namespace, obcluster.Name)
	}
	topology := make([]response.OBZone, 0, len(obzoneList.Items))
	for _, obzone := range obzoneList.Items {
		observers := make([]response.OBServer, 0)
		observerList, err := oceanbase.ListOBServersOfOBZone(&obzone)
		if err != nil {
			return nil, errors.Wrapf(err, "List observers of obzone %s %s", obzone.Namespace, obzone.Name)
		}
		logger.Infof("found %d observer", len(observerList.Items))
		for _, observer := range observerList.Items {
			logger.Infof("add observer %s to result", observer.Name)
			observers = append(observers, response.OBServer{
				Namespace:    observer.Namespace,
				Name:         observer.Name,
				Status:       convertStatus(observer.Status.Status),
				StatusDetail: observer.Status.Status,
				Address:      observer.Status.PodIp,
				// TODO: add metrics
				Metrics: nil,
			})
		}

		nodeSelector := make([]modelcommon.KVPair, 0)
		for k, v := range obzone.Spec.Topology.NodeSelector {
			nodeSelector = append(nodeSelector, modelcommon.KVPair{
				Key:   k,
				Value: v,
			})
		}
		topology = append(topology, response.OBZone{
			Namespace:    obzone.Namespace,
			Name:         obzone.Name,
			Zone:         obzone.Spec.Topology.Zone,
			Replicas:     obzone.Spec.Topology.Replica,
			Status:       convertStatus(obzone.Status.Status),
			StatusDetail: obzone.Status.Status,
			// TODO: query real rs
			RootService:  obzone.Status.OBServerStatus[0].Server,
			OBServers:    observers,
			NodeSelector: nodeSelector,
		})
	}
	return &response.OBCluster{
		Namespace:    obcluster.Namespace,
		Name:         obcluster.Name,
		ClusterName:  obcluster.Spec.ClusterName,
		ClusterId:    obcluster.Spec.ClusterId,
		Status:       getStatisticStatus(obcluster),
		StatusDetail: obcluster.Status.Status,
		CreateTime:   float64(obcluster.ObjectMeta.CreationTimestamp.UnixMilli()) / 1000,
		Image:        obcluster.Status.Image,
		Topology:     topology,
		// TODO: add metrics
		Metrics: nil,
	}, nil
}

func ListOBClusters() ([]response.OBCluster, error) {
	obclusters := make([]response.OBCluster, 0)
	obclusterList, err := oceanbase.ListAllOBClusters()
	if err != nil {
		return obclusters, errors.Wrap(err, "failed to list obclusters")
	}
	for _, obcluster := range obclusterList.Items {
		resp, err := buildOBClusterResponse(&obcluster)
		if err != nil {
			// TODO: add log here
		}
		obclusters = append(obclusters, *resp)
	}
	return obclusters, nil
}

func buildOBServerTemplate(observerSpec *param.OBServerSpec) *v1alpha1.OBServerTemplate {
	if observerSpec == nil {
		return nil
	}
	observerTemplate := &v1alpha1.OBServerTemplate{
		Image: observerSpec.Image,
		Resource: &v1alpha1.ResourceSpec{
			Cpu:    *apiresource.NewQuantity(observerSpec.Resource.Cpu, apiresource.DecimalSI),
			Memory: *apiresource.NewQuantity(observerSpec.Resource.MemoryGB*constant.GB, apiresource.BinarySI),
		},
		Storage: &v1alpha1.OceanbaseStorageSpec{
			DataStorage: &v1alpha1.StorageSpec{
				StorageClass: observerSpec.Storage.Data.StorageClass,
				Size:         *apiresource.NewQuantity(observerSpec.Storage.Data.SizeGB*constant.GB, apiresource.BinarySI),
			},
			RedoLogStorage: &v1alpha1.StorageSpec{
				StorageClass: observerSpec.Storage.RedoLog.StorageClass,
				Size:         *apiresource.NewQuantity(observerSpec.Storage.RedoLog.SizeGB*constant.GB, apiresource.BinarySI),
			},
			LogStorage: &v1alpha1.StorageSpec{
				StorageClass: observerSpec.Storage.Log.StorageClass,
				Size:         *apiresource.NewQuantity(observerSpec.Storage.Log.SizeGB*constant.GB, apiresource.BinarySI),
			},
		},
	}
	return observerTemplate
}

func buildMonitorTemplate(monitorSpec *param.MonitorSpec) *v1alpha1.MonitorTemplate {
	if monitorSpec == nil {
		return nil
	}
	monitorTemplate := &v1alpha1.MonitorTemplate{
		Image: monitorSpec.Image,
		Resource: &v1alpha1.ResourceSpec{
			Cpu:    *apiresource.NewQuantity(monitorSpec.Resource.Cpu, apiresource.DecimalSI),
			Memory: *apiresource.NewQuantity(monitorSpec.Resource.MemoryGB*constant.GB, apiresource.BinarySI),
		},
	}
	return monitorTemplate
}

func buildBackupVolume(nfsVolumeSpec *param.NFSVolumeSpec) *v1alpha1.BackupVolumeSpec {
	if nfsVolumeSpec == nil {
		return nil
	}
	backupVolume := &v1alpha1.BackupVolumeSpec{
		Volume: &corev1.Volume{
			Name: "ob-backup",
			VolumeSource: corev1.VolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server:   nfsVolumeSpec.Address,
					Path:     nfsVolumeSpec.Path,
					ReadOnly: false,
				},
			},
		},
	}
	return backupVolume
}

func buildOBClusterTopology(topology []param.ZoneTopology) []v1alpha1.OBZoneTopology {
	obzoneTopology := make([]v1alpha1.OBZoneTopology, 0)
	for _, zone := range topology {
		obzoneTopology = append(obzoneTopology, v1alpha1.OBZoneTopology{
			Zone:         zone.Zone,
			NodeSelector: common.KVsToMap(zone.NodeSelector),
			Replica:      zone.Replicas,
		})
	}
	return obzoneTopology
}

func buildOBClusterParameters(parameters []modelcommon.KVPair) []v1alpha1.Parameter {
	obparameters := make([]v1alpha1.Parameter, 0)
	for _, parameter := range parameters {
		obparameters = append(obparameters, v1alpha1.Parameter{
			Name:  parameter.Key,
			Value: parameter.Value,
		})
	}
	return obparameters
}

func generateUUID() string {
	parts := strings.Split(uuid.New().String(), "-")
	return parts[len(parts)-1]
}

func generateUserSecrets(clusterName string, clusterId int64) *v1alpha1.OBUserSecrets {
	return &v1alpha1.OBUserSecrets{
		Root:     fmt.Sprintf("%s-%d-root-%s", clusterName, clusterId, generateUUID()),
		ProxyRO:  fmt.Sprintf("%s-%d-proxyro-%s", clusterName, clusterId, generateUUID()),
		Monitor:  fmt.Sprintf("%s-%d-monitor-%s", clusterName, clusterId, generateUUID()),
		Operator: fmt.Sprintf("%s-%d-operator-%s", clusterName, clusterId, generateUUID()),
	}
}

func generateOBClusterInstance(param *param.CreateOBClusterParam) *v1alpha1.OBCluster {
	observerTemplate := buildOBServerTemplate(param.OBServer)
	monitorTemplate := buildMonitorTemplate(param.Monitor)
	backupVolume := buildBackupVolume(param.BackupVolume)
	parameters := buildOBClusterParameters(param.Parameters)
	topology := buildOBClusterTopology(param.Topology)
	obcluster := &v1alpha1.OBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: param.Namespace,
			Name:      param.Name,
		},
		Spec: v1alpha1.OBClusterSpec{
			ClusterName:      param.ClusterName,
			ClusterId:        param.ClusterId,
			OBServerTemplate: observerTemplate,
			MonitorTemplate:  monitorTemplate,
			BackupVolume:     backupVolume,
			Parameters:       parameters,
			Topology:         topology,
			UserSecrets:      generateUserSecrets(param.ClusterName, param.ClusterId),
		},
	}
	return obcluster
}

func CreateOBCluster(param *param.CreateOBClusterParam) error {
	obcluster := generateOBClusterInstance(param)
	err := oceanbase.CreateSecretsForOBCluster(obcluster, param.RootPassword)
	if err != nil {
		return errors.Wrap(err, "Create secrets for obcluster")
	}
	logger.Infof("Generated obcluster instance:%v", obcluster)
	return oceanbase.CreateOBCluster(obcluster)
}

func UpgradeObCluster(obclusterIdentity *param.K8sObjectIdentity, updateParam *param.UpgradeOBClusterParam) error {
	obcluster, err := oceanbase.GetOBCluster(obclusterIdentity.Namespace, obclusterIdentity.Name)
	if err != nil {
		return errors.Wrapf(err, "Get obcluster %s %s", obclusterIdentity.Namespace, obclusterIdentity.Name)
	}
	if obcluster.Status.Status != clusterstatus.Running {
		return errors.Errorf("Obcluster status invalid %s", obcluster.Status.Status)
	}
	obcluster.Spec.OBServerTemplate.Image = updateParam.Image
	return oceanbase.UpdateOBCluster(obcluster)
}

func ScaleOBServer(obzoneIdentity *param.OBZoneIdentity, scaleParam *param.ScaleOBServerParam) error {
	obcluster, err := oceanbase.GetOBCluster(obzoneIdentity.Namespace, obzoneIdentity.Name)
	if err != nil {
		return errors.Wrapf(err, "Get obcluster %s %s", obzoneIdentity.Namespace, obzoneIdentity.Name)
	}
	if obcluster.Status.Status != clusterstatus.Running {
		return errors.Errorf("Obcluster status invalid %s", obcluster.Status.Status)
	}
	found := false
	replicaChanged := false
	for idx, obzone := range obcluster.Spec.Topology {
		if obzone.Zone == obzoneIdentity.OBZoneName {
			found = true
			if obzone.Replica != scaleParam.Replicas {
				replicaChanged = true
				logger.Infof("Scale obzone %s from %d to %d", obzone.Zone, obzone.Replica, scaleParam.Replicas)
				obcluster.Spec.Topology[idx].Replica = scaleParam.Replicas
			}
		}
	}
	if !found {
		return errors.Errorf("obzone %s not found in obcluster %s %s", obzoneIdentity.OBZoneName, obzoneIdentity.Namespace, obzoneIdentity.Name)
	}
	if !replicaChanged {
		return errors.Errorf("obzone %s replica already satisfied in obcluster %s %s", obzoneIdentity.OBZoneName, obzoneIdentity.Namespace, obzoneIdentity.Name)
	}
	return oceanbase.UpdateOBCluster(obcluster)
}

func DeleteOBZone(obzoneIdentity *param.OBZoneIdentity) error {
	obcluster, err := oceanbase.GetOBCluster(obzoneIdentity.Namespace, obzoneIdentity.Name)
	if err != nil {
		return errors.Wrapf(err, "Get obcluster %s %s", obzoneIdentity.Namespace, obzoneIdentity.Name)
	}
	if obcluster.Status.Status != clusterstatus.Running {
		return errors.Errorf("Obcluster status invalid %s", obcluster.Status.Status)
	}
	newTopology := make([]v1alpha1.OBZoneTopology, 0)
	found := false
	for _, obzone := range obcluster.Spec.Topology {
		if obzone.Zone != obzoneIdentity.OBZoneName {
			newTopology = append(newTopology, obzone)
		} else {
			found = true
		}
	}
	if !found {
		return errors.Errorf("obzone %s not found in obcluster %s %s", obzoneIdentity.OBZoneName, obzoneIdentity.Namespace, obzoneIdentity.Name)
	}
	obcluster.Spec.Topology = newTopology
	return oceanbase.UpdateOBCluster(obcluster)
}

func AddOBZone(obclusterIdentity *param.K8sObjectIdentity, zone *param.ZoneTopology) error {
	obcluster, err := oceanbase.GetOBCluster(obclusterIdentity.Namespace, obclusterIdentity.Name)
	if err != nil {
		return errors.Wrapf(err, "Get obcluster %s %s", obclusterIdentity.Namespace, obclusterIdentity.Name)
	}
	if obcluster.Status.Status != clusterstatus.Running {
		return errors.Errorf("Obcluster status invalid %s", obcluster.Status.Status)
	}
	for _, obzone := range obcluster.Spec.Topology {
		if obzone.Zone == zone.Zone {
			return errors.Errorf("obzone %s already exists", zone.Zone)
		}
	}
	obcluster.Spec.Topology = append(obcluster.Spec.Topology, v1alpha1.OBZoneTopology{
		Zone:         zone.Zone,
		NodeSelector: common.KVsToMap(zone.NodeSelector),
		Replica:      zone.Replicas,
	})
	return oceanbase.UpdateOBCluster(obcluster)
}

func GetOBCluster(obclusterIdentity *param.K8sObjectIdentity) (*response.OBCluster, error) {
	obcluster, err := oceanbase.GetOBCluster(obclusterIdentity.Namespace, obclusterIdentity.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "Get obcluster %s %s", obclusterIdentity.Namespace, obclusterIdentity.Name)
	}
	return buildOBClusterResponse(obcluster)
}

func DeleteOBCluster(obclusterIdentity *param.K8sObjectIdentity) error {
	return oceanbase.DeleteOBCluster(obclusterIdentity.Namespace, obclusterIdentity.Name)
}

func GetOBClusterStatistic() ([]response.OBClusterStastistic, error) {
	statisticResult := make([]response.OBClusterStastistic, 0)
	obclusterList, err := oceanbase.ListAllOBClusters()
	if err != nil {
		return statisticResult, errors.Wrap(err, "failed to list obclusters")
	}
	statusMap := make(map[string]int)
	for _, obcluster := range obclusterList.Items {
		statisticStatus := getStatisticStatus(&obcluster)
		cnt, found := statusMap[statisticStatus]
		if found {
			cnt = cnt + 1
		} else {
			cnt = 1
		}
		statusMap[statisticStatus] = cnt
	}
	for status, count := range statusMap {
		statisticResult = append(statisticResult, response.OBClusterStastistic{
			Status: status,
			Count:  count,
		})
	}
	return statisticResult, nil
}
