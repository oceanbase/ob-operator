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

func GenerateStatefulAppBootStrapZoneSpec(zone []cloudv1.Subset) []cloudv1.Subset {
	res := make([]cloudv1.Subset, 0)
	for _, subset := range zone {
		tmp := subset
		tmp.Replicas = 1
		res = append(res, tmp)
	}
	return res
}

func GenerateImage(version string) string {
	image := fmt.Sprintf("%s-%s", observerconst.ImgProfix, version)
	return image
}

func GeneratePodSpec(obClusterSpec cloudv1.OBClusterSpec) corev1.PodSpec {
	port := make([]corev1.ContainerPort, 0)
	cablePort := corev1.ContainerPort{}
	cablePort.Name = observerconst.CablePortName
	cablePort.ContainerPort = int32(observerconst.CablePort)
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
	volumeMounts := make([]corev1.VolumeMount, 0)
	volumeMounts = append(volumeMounts, volumeMountDataFile)
	volumeMounts = append(volumeMounts, volumeMountDataLog)
	volumeMounts = append(volumeMounts, volumeMountLog)

	readinessProbeHTTP := corev1.HTTPGetAction{}
	readinessProbeHTTP.Port = intstr.FromInt(observerconst.CablePort)
	readinessProbeHTTP.Path = observerconst.CableReadinessUrl
	readinessProbe := corev1.Probe{}
	readinessProbe.Handler.HTTPGet = &readinessProbeHTTP
	readinessProbe.PeriodSeconds = observerconst.CableReadinessPeriod

	container := corev1.Container{
		Name:            observerconst.ImgOb,
		Image:           GenerateImage(obClusterSpec.Version),
		ImagePullPolicy: observerconst.ImgPullPolicy,
		Ports:           port,
		Resources:       resources,
		VolumeMounts:    volumeMounts,
		ReadinessProbe:  &readinessProbe,
	}
	containers := make([]corev1.Container, 0)
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
	volumes := make([]corev1.Volume, 0)
	volumes = append(volumes, volumeDataFile)
	volumes = append(volumes, volumeDataLog)
	volumes = append(volumes, volumeLog)

	podSpec := corev1.PodSpec{
		Containers: containers,
		Volumes:    volumes,
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
		zoneList = append(zoneList, zone)
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
