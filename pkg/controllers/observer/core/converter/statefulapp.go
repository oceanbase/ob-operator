/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package converter

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	observerutil "github.com/oceanbase/ob-operator/pkg/controllers/observer/core/util"
	statefulappCore "github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/const"
)

func GenerateStatefulAppName(name string) string {
	statefulAppName := fmt.Sprintf("sapp-%s", name)
	return statefulAppName
}

func GetObClusterName(name string) string {
	return name[5:]
}

func GenerateStatefulAppBootStrapZoneSpec(zone []cloudv1.Subset) []cloudv1.Subset {
	res := make([]cloudv1.Subset, 0)
	for _, subset := range zone {
		tmp := subset
		tmp.Replicas = 1
		res = append(res, tmp)
	}
	return res
}

func GenerateObContainer(obClusterSpec cloudv1.OBClusterSpec) corev1.Container {
	port := make([]corev1.ContainerPort, 0)
	cablePort := corev1.ContainerPort{}
	cablePort.Name = observerconst.CablePortName
	cablePort.ContainerPort = observerconst.CablePort
	cablePort.Protocol = corev1.ProtocolTCP
	port = append(port, cablePort)
	mysqlPort := corev1.ContainerPort{}
	mysqlPort.Name = observerconst.MysqlPortName
	mysqlPort.ContainerPort = observerconst.MysqlPort
	mysqlPort.Protocol = corev1.ProtocolTCP
	port = append(port, mysqlPort)
	rpcPort := corev1.ContainerPort{}
	rpcPort.Name = observerconst.RpcPortName
	rpcPort.ContainerPort = observerconst.RpcPort
	rpcPort.Protocol = corev1.ProtocolTCP
	port = append(port, rpcPort)

	requestsResources := corev1.ResourceList{}
	requestsResources["cpu"] = obClusterSpec.Resources.CPU
	requestsResources["memory"] = obClusterSpec.Resources.Memory
	limitResources := corev1.ResourceList{}
	limitResources["cpu"] = obClusterSpec.Resources.CPU
	limitResources["memory"] = obClusterSpec.Resources.Memory
	resources := corev1.ResourceRequirements{
		Requests: requestsResources,
		Limits:   limitResources,
	}

	volumeMountDataFile := corev1.VolumeMount{}
	volumeMountDataFile.Name = observerconst.DatafileStorageName
	volumeMountDataFile.MountPath = observerconst.DatafileStoragePath
	volumeMountDataLog := corev1.VolumeMount{}
	volumeMountDataLog.Name = observerconst.DatalogStorageName
	volumeMountDataLog.MountPath = observerconst.DatalogStoragePath
	volumeMountLog := corev1.VolumeMount{}
	volumeMountLog.Name = observerconst.LogStorageName
	volumeMountLog.MountPath = observerconst.LogStoragePath
	volumeMountBackup := corev1.VolumeMount{}
	volumeMountBackup.Name = observerconst.BackupName
	volumeMountBackup.MountPath = observerconst.BackupPath

	volumeMounts := make([]corev1.VolumeMount, 0)
	volumeMounts = append(volumeMounts, volumeMountDataFile)
	volumeMounts = append(volumeMounts, volumeMountDataLog)
	volumeMounts = append(volumeMounts, volumeMountLog)

	backupVolumeSpec := obClusterSpec.Resources.Volume
	if backupVolumeSpec.Name != "" {
		volumeMounts = append(volumeMounts, volumeMountBackup)
	}
	readinessProbeHTTP := corev1.HTTPGetAction{}
	readinessProbeHTTP.Port = intstr.FromInt(observerconst.CablePort)
	readinessProbeHTTP.Path = observerconst.CableReadinessUrl
	readinessProbe := corev1.Probe{}
	readinessProbe.Handler.HTTPGet = &readinessProbeHTTP
	readinessProbe.PeriodSeconds = observerconst.CableReadinessPeriod
	container := corev1.Container{
		Name:            observerconst.ImgOb,
		Image:           fmt.Sprintf("%s:%s", obClusterSpec.ImageRepo, obClusterSpec.Tag),
		ImagePullPolicy: observerconst.ImgPullPolicy,
		Ports:           port,
		Resources:       resources,
		VolumeMounts:    volumeMounts,
		ReadinessProbe:  &readinessProbe,
	}
	return container
}

func GenerateObagentContainer(obClusterSpec cloudv1.OBClusterSpec) corev1.Container {

	ports := make([]corev1.ContainerPort, 0)
	monagentPort := corev1.ContainerPort{}
	monagentPort.Name = observerconst.MonagentPortName
	monagentPort.ContainerPort = observerconst.MonagentPort
	monagentPort.Protocol = corev1.ProtocolTCP
	ports = append(ports, monagentPort)

	volumeMountConfFile := corev1.VolumeMount{}
	volumeMountConfFile.Name = observerconst.ConfFileStorageName
	volumeMountConfFile.MountPath = observerconst.ConfFileStoragePath
	volumeMounts := make([]corev1.VolumeMount, 0)
	volumeMounts = append(volumeMounts, volumeMountConfFile)

	readinessProbeHTTP := corev1.HTTPGetAction{}
	readinessProbeHTTP.Port = intstr.FromInt(observerconst.MonagentPort)
	readinessProbeHTTP.Path = observerconst.MonagentReadinessUrl
	readinessProbe := corev1.Probe{}
	readinessProbe.Handler.HTTPGet = &readinessProbeHTTP
	readinessProbe.PeriodSeconds = observerconst.MonagentConfigPeriod
	container := corev1.Container{
		Name:            observerconst.ImgObagent,
		Image:           obClusterSpec.ImageObagent,
		ImagePullPolicy: observerconst.ImgPullPolicy,
		Ports:           ports,
		VolumeMounts:    volumeMounts,
		ReadinessProbe:  &readinessProbe,
	}
	return container
}

func GeneratePodSpec(obClusterSpec cloudv1.OBClusterSpec) corev1.PodSpec {

	container := GenerateObContainer(obClusterSpec)
	containers := make([]corev1.Container, 0)
	containers = append(containers, container)

	container = GenerateObagentContainer(obClusterSpec)
	containers = append(containers, container)

	volumeDataFile := corev1.Volume{}
	volumeDataFile.Name = observerconst.DatafileStorageName
	volumeDataFileSource := &corev1.PersistentVolumeClaimVolumeSource{
		ClaimName: observerconst.DatafileStorageName,
	}
	volumeDataFile.VolumeSource.PersistentVolumeClaim = volumeDataFileSource
	volumeDataLog := corev1.Volume{}
	volumeDataLog.Name = observerconst.DatalogStorageName
	volumeDataLogSource := &corev1.PersistentVolumeClaimVolumeSource{
		ClaimName: observerconst.DatalogStorageName,
	}
	volumeDataLog.VolumeSource.PersistentVolumeClaim = volumeDataLogSource
	volumeLog := corev1.Volume{}
	volumeLog.Name = observerconst.LogStorageName
	volumeLogSource := &corev1.PersistentVolumeClaimVolumeSource{
		ClaimName: observerconst.LogStorageName,
	}
	volumeLog.VolumeSource.PersistentVolumeClaim = volumeLogSource
	volumeObagentConfFile := corev1.Volume{}
	volumeObagentConfFile.Name = observerconst.ConfFileStorageName
	volumeObagentConfFileSource := &corev1.PersistentVolumeClaimVolumeSource{
		ClaimName: observerconst.ConfFileStorageName,
	}
	volumeObagentConfFile.VolumeSource.PersistentVolumeClaim = volumeObagentConfFileSource
	volumes := make([]corev1.Volume, 0)
	volumes = append(volumes, volumeDataFile)
	volumes = append(volumes, volumeDataLog)
	volumes = append(volumes, volumeLog)
	volumes = append(volumes, volumeObagentConfFile)

	backupVolumeSpec := obClusterSpec.Resources.Volume
	if backupVolumeSpec.Name != "" {
		backupHostPathType := corev1.HostPathType(corev1.HostPathDirectory)
		backupHostPath := corev1.HostPathVolumeSource{
			Path: backupVolumeSpec.Path,
			Type: &backupHostPathType,
		}
		backupVolume := corev1.Volume{
			Name: backupVolumeSpec.Name,
		}
		backupVolume.HostPath = &backupHostPath
		volumes = append(volumes, backupVolume)
	}

	podSpec := corev1.PodSpec{
		Volumes:    volumes,
		Containers: containers,
	}
	return podSpec
}

func GenerateStorageSpec(obClusterSpec cloudv1.OBClusterSpec) []cloudv1.StorageTemplate {
	storageTemplates := make([]cloudv1.StorageTemplate, 0)
	for _, storageSpec := range obClusterSpec.Resources.Storage {
		storageTemplate := cloudv1.StorageTemplate{}
		storageTemplate.Name = storageSpec.Name
		requestsResources := corev1.ResourceList{}
		requestsResources["storage"] = storageSpec.Size
		storageTemplate.PVC.StorageClassName = &(storageSpec.StorageClassName)
		accessModes := make([]corev1.PersistentVolumeAccessMode, 0)
		accessModes = append(accessModes, corev1.ReadWriteOnce)
		storageTemplate.PVC.AccessModes = accessModes
		storageTemplate.PVC.Resources.Requests = requestsResources
		storageTemplates = append(storageTemplates, storageTemplate)
	}
	return storageTemplates
}

func GenerateStatefulAppSpec(cluster cloudv1.Cluster, obCluster cloudv1.OBCluster) cloudv1.StatefulAppSpec {
	subset := GenerateStatefulAppBootStrapZoneSpec(cluster.Zone)
	podSpec := GeneratePodSpec(obCluster.Spec)
	storage := GenerateStorageSpec(obCluster.Spec)
	spec := cloudv1.StatefulAppSpec{
		Cluster:          cluster.Cluster,
		Subsets:          subset,
		PodTemplate:      podSpec,
		StorageTemplates: storage,
	}
	return spec
}

func GenerateStatefulAppObject(cluster cloudv1.Cluster, obCluster cloudv1.OBCluster) cloudv1.StatefulApp {
	name := GenerateStatefulAppName(obCluster.Name)
	objectMeta := observerutil.GenerateObjectMeta(obCluster, name)
	spec := GenerateStatefulAppSpec(cluster, obCluster)
	statefulApp := cloudv1.StatefulApp{
		ObjectMeta: objectMeta,
		Spec:       spec,
	}
	return statefulApp
}

func GetSubsetStatusFromStatefulApp(zoneName string, statefulApp cloudv1.StatefulApp) cloudv1.SubsetStatus {
	var res cloudv1.SubsetStatus
	for _, subset := range statefulApp.Status.Subsets {
		if subset.Name == zoneName {
			res = subset
			break
		}
	}
	return res
}

func UpdateSubsetReplicaForStatefulApp(subset cloudv1.Subset, statefulApp cloudv1.StatefulApp) cloudv1.StatefulApp {
	zoneList := make([]cloudv1.Subset, 0)

	isExist := false
	for _, zone := range statefulApp.Spec.Subsets {
		if zone.Name == subset.Name && zone.Replicas != 0 {
			isExist = true
		}
	}

	if !isExist {
		newZone := subset
		newZone.Replicas = 1
		zoneList = append(zoneList, newZone)
	}

	for _, zone := range statefulApp.Spec.Subsets {
		if zone.Name == subset.Name {
			// one by one
			if zone.Replicas < subset.Replicas {
				zone.Replicas = zone.Replicas + 1
			}
			if zone.Replicas > subset.Replicas {
				zone.Replicas = zone.Replicas - 1
			}
		}
		if zone.Replicas > 0 {
			zoneList = append(zoneList, zone)
		}

	}
	statefulApp.Spec.Subsets = zoneList
	return statefulApp
}

func CheckStatefulAppStatus(statefulApp cloudv1.StatefulApp) bool {
	if statefulApp.Status.ClusterStatus != statefulappCore.Ready {
		return false
	}
	return true
}

func UpdateZoneForStatefulApp(clusterList []cloudv1.Cluster, statefulApp cloudv1.StatefulApp) cloudv1.StatefulApp {
	cluster := GetClusterSpecFromOBTopology(clusterList)
	zoneList := cluster.Zone
	statefulApp.Spec.Subsets = zoneList
	return statefulApp
}
