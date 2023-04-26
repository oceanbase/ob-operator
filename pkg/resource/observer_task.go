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
	"fmt"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	cloudv2alpha1 "github.com/oceanbase/ob-operator/api/v2alpha1"
	util "github.com/oceanbase/ob-operator/pkg/util"
)

func (m *OBServerManager) WaitOBPodReady() error {
	for i := 0; i < 60; i++ {
		observer, err := m.getOBServer()
		if err != nil {
			return errors.Wrap(err, "get observer from K8s")
		}
		if observer.Status.Ready {
			m.Logger.Info("Pod is ready")
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("Timeout to wait pod ready")
}

func (m *OBServerManager) AddServer() error {
	return nil
}

func (m *OBServerManager) generateOBServerStartArgs() map[string]interface{} {
	observerStartArgs := make(map[string]interface{})
	observerStartArgs["clusterName"] = m.OBServer.Spec.ClusterName
	observerStartArgs["clusterId"] = m.OBServer.Spec.ClusterId
	observerStartArgs["zoneName"] = m.OBServer.Spec.Zone
	cpu, _ := m.OBServer.Spec.OBServerTemplate.Resource.Cpu.AsInt64()
	memory, _ := m.OBServer.Spec.OBServerTemplate.Resource.Memory.AsInt64()
	observerStartArgs["cpuLimit"] = cpu
	observerStartArgs["memoryLimit"] = memory / 1073741824
	observerStartArgs["version"] = "4"

	//TODO no need to pass this parameter
	observerStartArgs["rsList"] = fmt.Sprintf("%s:2881", m.OBServer.Status.PodIp)
	return observerStartArgs
}

func (m *OBServerManager) StartOBServer() error {
	url := fmt.Sprintf("http://%s:19001/api/ob/start", m.OBServer.Status.PodIp)
	observerStartArgs := m.generateOBServerStartArgs()
	code, resp := util.HTTPPOST(url, util.CovertToJSON(observerStartArgs))
	m.Logger.Info("get resp", "resp", resp)
	if code != 200 {
		return errors.New("start observer failed")
	}
	return nil
}

func (m *OBServerManager) CreateOBPod() error {
	m.Logger.Info("create observer pod")
	ownerReferenceList := make([]metav1.OwnerReference, 0)
	ownerReference := metav1.OwnerReference{
		APIVersion: m.OBServer.APIVersion,
		Kind:       m.OBServer.Kind,
		Name:       m.OBServer.Name,
		UID:        m.OBServer.GetUID(),
	}
	ownerReferenceList = append(ownerReferenceList, ownerReference)
	observerPodSpec := m.createOBServerPodSpec()
	// create pod
	observerPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            m.OBServer.Name,
			Namespace:       m.OBServer.Namespace,
			OwnerReferences: ownerReferenceList,
			Labels:          m.OBServer.Labels,
		},
		Spec: observerPodSpec,
	}
	err := m.Client.Create(m.Ctx, observerPod)
	if err != nil {
		m.Logger.Error(err, "failed to create pod")
		return errors.Wrap(err, "failed to create pod")
	}
	return nil
}

func (m *OBServerManager) generatePVCSpec(name string, storageSpec *cloudv2alpha1.StorageSpec) corev1.PersistentVolumeClaimSpec {
	pvcSpec := &corev1.PersistentVolumeClaimSpec{}
	requestsResources := corev1.ResourceList{}
	requestsResources["storage"] = storageSpec.Size
	storageClassName := storageSpec.StorageClass
	pvcSpec.StorageClassName = &(storageClassName)
	accessModes := make([]corev1.PersistentVolumeAccessMode, 0)
	accessModes = append(accessModes, corev1.ReadWriteOnce)
	pvcSpec.AccessModes = accessModes
	pvcSpec.Resources.Requests = requestsResources
	return *pvcSpec
}

func (m *OBServerManager) CreateOBPVC() error {
	ownerReferenceList := make([]metav1.OwnerReference, 0)
	ownerReference := metav1.OwnerReference{
		APIVersion: m.OBServer.APIVersion,
		Kind:       m.OBServer.Kind,
		Name:       m.OBServer.Name,
		UID:        m.OBServer.GetUID(),
	}
	ownerReferenceList = append(ownerReferenceList, ownerReference)

	objectMeta := metav1.ObjectMeta{
		Name:            fmt.Sprintf("%s-data-file", m.OBServer.Name),
		Namespace:       m.OBServer.Namespace,
		OwnerReferences: ownerReferenceList,
		Labels:          m.OBServer.Labels,
	}
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: objectMeta,
		Spec:       m.generatePVCSpec(fmt.Sprintf("%s-data-file", m.OBServer.Name), m.OBServer.Spec.OBServerTemplate.Storage.DataStorage),
	}
	err := m.Client.Create(m.Ctx, pvc)
	if err != nil {
		return errors.Wrap(err, "Create pvc of data file")
	}

	objectMeta = metav1.ObjectMeta{
		Name:            fmt.Sprintf("%s-data-log", m.OBServer.Name),
		Namespace:       m.OBServer.Namespace,
		OwnerReferences: ownerReferenceList,
		Labels:          m.OBServer.Labels,
	}
	pvc = &corev1.PersistentVolumeClaim{
		ObjectMeta: objectMeta,
		Spec:       m.generatePVCSpec(fmt.Sprintf("%s-log-file", m.OBServer.Name), m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage),
	}
	err = m.Client.Create(m.Ctx, pvc)
	if err != nil {
		return errors.Wrap(err, "Create pvc of data log")
	}

	objectMeta = metav1.ObjectMeta{
		Name:            fmt.Sprintf("%s-log", m.OBServer.Name),
		Namespace:       m.OBServer.Namespace,
		OwnerReferences: ownerReferenceList,
		Labels:          m.OBServer.Labels,
	}
	pvc = &corev1.PersistentVolumeClaim{
		ObjectMeta: objectMeta,
		Spec:       m.generatePVCSpec(fmt.Sprintf("%s-log", m.OBServer.Name), m.OBServer.Spec.OBServerTemplate.Storage.LogStorage),
	}
	err = m.Client.Create(m.Ctx, pvc)
	if err != nil {
		return errors.Wrap(err, "Create pvc of log")
	}
	return nil
}

func (m *OBServerManager) createOBServerPodSpec() corev1.PodSpec {
	containers := make([]corev1.Container, 0)
	observerContainer := m.createOBServerContainer()
	containers = append(containers, observerContainer)

	// TODO, add monitor container
	volumeDataFile := corev1.Volume{}
	volumeDataFile.Name = fmt.Sprintf("%s-data-file", m.OBServer.Name)
	volumeDataFileSource := &corev1.PersistentVolumeClaimVolumeSource{
		ClaimName: fmt.Sprintf("%s-data-file", m.OBServer.Name),
	}
	volumeDataFile.VolumeSource.PersistentVolumeClaim = volumeDataFileSource

	volumeDataLog := corev1.Volume{}
	volumeDataLog.Name = fmt.Sprintf("%s-data-log", m.OBServer.Name)
	volumeDataLogSource := &corev1.PersistentVolumeClaimVolumeSource{
		ClaimName: fmt.Sprintf("%s-data-log", m.OBServer.Name),
	}
	volumeDataLog.VolumeSource.PersistentVolumeClaim = volumeDataLogSource

	volumeLog := corev1.Volume{}
	volumeLog.Name = fmt.Sprintf("%s-log", m.OBServer.Name)
	volumeLogSource := &corev1.PersistentVolumeClaimVolumeSource{
		ClaimName: fmt.Sprintf("%s-log", m.OBServer.Name),
	}
	volumeLog.VolumeSource.PersistentVolumeClaim = volumeLogSource

	volumes := make([]corev1.Volume, 0)
	volumes = append(volumes, volumeDataFile)
	volumes = append(volumes, volumeDataLog)
	volumes = append(volumes, volumeLog)

	podSpec := corev1.PodSpec{
		Volumes:    volumes,
		Containers: containers,
	}
	return podSpec
}

// TODO move hardcoded values to another file
func (m *OBServerManager) createOBServerContainer() corev1.Container {
	// port info
	ports := make([]corev1.ContainerPort, 0)
	cablePort := corev1.ContainerPort{}
	cablePort.Name = "cable"
	cablePort.ContainerPort = 19001
	cablePort.Protocol = corev1.ProtocolTCP
	mysqlPort := corev1.ContainerPort{}
	mysqlPort.Name = "mysql"
	mysqlPort.ContainerPort = 2881
	mysqlPort.Protocol = corev1.ProtocolTCP
	rpcPort := corev1.ContainerPort{}
	rpcPort.Name = "rpc"
	rpcPort.ContainerPort = 2882
	rpcPort.Protocol = corev1.ProtocolTCP
	ports = append(ports, cablePort)
	ports = append(ports, mysqlPort)
	ports = append(ports, rpcPort)

	// resource info
	observerResource := corev1.ResourceList{}
	observerResource["cpu"] = m.OBServer.Spec.OBServerTemplate.Resource.Cpu
	observerResource["memory"] = m.OBServer.Spec.OBServerTemplate.Resource.Memory
	resources := corev1.ResourceRequirements{
		Requests: observerResource,
		Limits:   observerResource,
	}

	// volume mounts
	volumeMountDataFile := corev1.VolumeMount{}
	volumeMountDataFile.Name = fmt.Sprintf("%s-data-file", m.OBServer.Name)
	volumeMountDataFile.MountPath = "/home/admin/data-file"
	volumeMountDataLog := corev1.VolumeMount{}
	volumeMountDataLog.Name = fmt.Sprintf("%s-data-log", m.OBServer.Name)
	volumeMountDataLog.MountPath = "/home/admin/data-log"
	volumeMountLog := corev1.VolumeMount{}
	volumeMountLog.Name = fmt.Sprintf("%s-log", m.OBServer.Name)
	volumeMountLog.MountPath = "/home/admin/log"

	volumeMounts := make([]corev1.VolumeMount, 0)
	volumeMounts = append(volumeMounts, volumeMountDataFile)
	volumeMounts = append(volumeMounts, volumeMountDataLog)
	volumeMounts = append(volumeMounts, volumeMountLog)

	readinessProbeTCP := corev1.TCPSocketAction{}
	readinessProbeTCP.Port = intstr.FromInt(2881)
	readinessProbe := corev1.Probe{}
	readinessProbe.ProbeHandler.TCPSocket = &readinessProbeTCP
	readinessProbe.PeriodSeconds = 2
	readinessProbe.InitialDelaySeconds = 5

	makeLogDirCmd := "mkdir -p /home/admin/log/log && ln -sf /home/admin/log/log /home/admin/oceanbase/log"
	makeStoreDirCmd := "mkdir -p /home/admin/oceanbase/store"
	makeCLogDirCmd := "mkdir -p /home/admin/data-log/clog && ln -sf /home/admin/data-log/clog /home/admin/oceanbase/store/clog"
	makeILogDirCmd := "mkdir -p /home/admin/data-log/ilog && ln -sf /home/admin/data-log/ilog /home/admin/oceanbase/store/ilog"
	makeSLogDirCmd := "mkdir -p /home/admin/data-file/slog && ln -sf /home/admin/data-file/slog /home/admin/oceanbase/store/slog"
	makeEtcDirCmd := "mkdir -p /home/admin/data-file/etc && ln -sf /home/admin/data-file/etc /home/admin/oceanbase/store/etc"
	makeSortDirCmd := "mkdir -p /home/admin/data-file/sort_dir && ln -sf /home/admin/data-file/sort_dir /home/admin/oceanbase/store/sort_dir"
	makeSstableDirCmd := "mkdir -p /home/admin/data-file/sstable && ln -sf /home/admin/data-file/sstable /home/admin/oceanbase/store/sstable"

	// TODO this config is only for small quota, should calculate based on resource
	optStr := "cpu_count=16,memory_limit=9G,system_memory=1G,__min_full_resource_pool_memory=1073741824,datafile_size=40G,log_disk_size=40G,net_thread_count=2,stack_size=512K,cache_wash_threshold=1G,schema_history_expire_time=1d,enable_separate_sys_clog=false,enable_merge_by_turn=false,enable_syslog_recycle=true,enable_syslog_wf=false,max_syslog_file_count=4"

	startObserverCmd := fmt.Sprintf("chown -R root:root /home/admin/oceanbase && /home/admin/oceanbase/bin/observer --nodaemon --appname %s --cluster_id %d --zone %s --devname eth0 -p 2881 -P 2882 -d /home/admin/oceanbase/store/ -l info -o config_additional_dir=/home/admin/oceanbase/store/etc,%s", m.OBServer.Spec.ClusterName, m.OBServer.Spec.ClusterId, m.OBServer.Spec.Zone, optStr)

	cmds := []string{
		"bash",
		"-c",
		fmt.Sprintf(" %s && %s && %s && %s && %s && %s && %s && %s && %s ", makeLogDirCmd, makeStoreDirCmd, makeCLogDirCmd, makeILogDirCmd, makeSLogDirCmd, makeEtcDirCmd, makeSortDirCmd, makeSstableDirCmd, startObserverCmd),
	}

	env := make([]corev1.EnvVar, 0)
	envLib := corev1.EnvVar{
		Name:  "LD_LIBRARY_PATH",
		Value: "/home/admin/oceanbase/lib",
	}
	env = append(env, envLib)

	container := corev1.Container{
		Name:            "observer",
		Image:           m.OBServer.Spec.OBServerTemplate.Image,
		ImagePullPolicy: "IfNotPresent",
		Ports:           ports,
		Resources:       resources,
		VolumeMounts:    volumeMounts,
		ReadinessProbe:  &readinessProbe,
		WorkingDir:      "/home/admin/oceanbase",
		Env:             env,
		Command:         cmds,
	}
	return container
}
