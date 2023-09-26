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

	obagentconst "github.com/oceanbase/ob-operator/pkg/const/obagent"
	oceanbaseconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"
	podconst "github.com/oceanbase/ob-operator/pkg/const/pod"
	secretconst "github.com/oceanbase/ob-operator/pkg/const/secret"
	clusterstatus "github.com/oceanbase/ob-operator/pkg/const/status/obcluster"
	observerstatus "github.com/oceanbase/ob-operator/pkg/oceanbase/const/status/server"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
)

func (m *OBServerManager) WaitOBServerReady() error {
	for i := 0; i < podconst.ReadyTimeoutSeconds; i++ {
		observer, err := m.getOBServer()
		if err != nil {
			return errors.Wrap(err, "Get observer from K8s")
		}
		if observer.Status.Ready {
			m.Logger.Info("Pod is ready")
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("Timeout to wait pod ready")
}

func (m *OBServerManager) getOceanbaseOperationManager() (*operation.OceanbaseOperationManager, error) {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return nil, errors.Wrap(err, "Get obcluster from K8s")
	}
	return GetOceanbaseOperationManagerFromOBCluster(m.Client, m.Logger, obcluster)
}

func (m *OBServerManager) AddServer() error {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		m.Logger.Error(err, "Get oceanbase operation manager failed")
		return errors.Wrap(err, "Get oceanbase operation manager")
	}
	serverInfo := &model.ServerInfo{
		Ip:   m.OBServer.Status.PodIp,
		Port: oceanbaseconst.RpcPort,
	}
	obs, err := oceanbaseOperationManager.GetServer(serverInfo)
	if obs != nil {
		m.Logger.Info("Observer already exists in obcluster")
		return nil
	}
	if err != nil {
		m.Logger.Error(err, "Get observer failed")
		return errors.Wrap(err, "Failed to get observer")
	}
	return oceanbaseOperationManager.AddServer(serverInfo)
}

func (m *OBServerManager) WaitOBClusterBootstrapped() error {
	for i := 0; i < oceanbaseconst.BootstrapTimeoutSeconds; i++ {
		obcluster, err := m.getOBCluster()
		if err != nil {
			return errors.Wrap(err, "Get obcluster from K8s")
		}
		if obcluster.Status.Status == clusterstatus.Bootstrapped {
			m.Logger.Info("Obcluster bootstrapped")
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("Timeout to wait obcluster bootstrapped")
}

func (m *OBServerManager) generateStaticIpAnnotation() map[string]string {
	annotations := make(map[string]string)
	switch m.OBServer.Status.CNI {
	case oceanbaseconst.CNICalico:
		if m.OBServer.Status.PodIp != "" {
			annotations[oceanbaseconst.AnnotationCalicoIpAddrs] = fmt.Sprintf("[\"%s\"]", m.OBServer.Status.PodIp)
		}
	default:
		m.Logger.Info("static ip not supported, set empty annotation")
	}
	return annotations
}

func (m *OBServerManager) CreateOBPod() error {
	m.Logger.Info("create observer pod")
	obcluster, err := m.getOBCluster()
	if err != nil {
		return errors.Wrap(err, "Get obcluster from K8s")
	}
	ownerReferenceList := make([]metav1.OwnerReference, 0)
	ownerReference := metav1.OwnerReference{
		APIVersion: m.OBServer.APIVersion,
		Kind:       m.OBServer.Kind,
		Name:       m.OBServer.Name,
		UID:        m.OBServer.GetUID(),
	}
	annotations := m.generateStaticIpAnnotation()
	ownerReferenceList = append(ownerReferenceList, ownerReference)
	observerPodSpec := m.createOBPodSpec(obcluster)
	// create pod
	observerPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            m.OBServer.Name,
			Namespace:       m.OBServer.Namespace,
			OwnerReferences: ownerReferenceList,
			Labels:          m.OBServer.Labels,
			Annotations:     annotations,
		},
		Spec: observerPodSpec,
	}
	err = m.Client.Create(m.Ctx, observerPod)
	if err != nil {
		m.Logger.Error(err, "failed to create pod")
		return errors.Wrap(err, "failed to create pod")
	}
	return nil
}

func (m *OBServerManager) generatePVCSpec(name string, storageSpec *v1alpha1.StorageSpec) corev1.PersistentVolumeClaimSpec {
	_ = name
	pvcSpec := &corev1.PersistentVolumeClaimSpec{}
	requestsResources := corev1.ResourceList{}
	requestsResources["storage"] = storageSpec.Size
	storageClassName := storageSpec.StorageClass
	pvcSpec.StorageClassName = &(storageClassName)
	accessModes := []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
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
		Name:            fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.DataVolumeSuffix),
		Namespace:       m.OBServer.Namespace,
		OwnerReferences: ownerReferenceList,
		Labels:          m.OBServer.Labels,
	}
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: objectMeta,
		Spec:       m.generatePVCSpec(fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.DataVolumeSuffix), m.OBServer.Spec.OBServerTemplate.Storage.DataStorage),
	}
	err := m.Client.Create(m.Ctx, pvc)
	if err != nil {
		return errors.Wrap(err, "Create pvc of data file")
	}

	objectMeta = metav1.ObjectMeta{
		Name:            fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.ClogVolumeSuffix),
		Namespace:       m.OBServer.Namespace,
		OwnerReferences: ownerReferenceList,
		Labels:          m.OBServer.Labels,
	}
	pvc = &corev1.PersistentVolumeClaim{
		ObjectMeta: objectMeta,
		Spec:       m.generatePVCSpec(fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.ClogVolumeSuffix), m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage),
	}
	err = m.Client.Create(m.Ctx, pvc)
	if err != nil {
		return errors.Wrap(err, "Create pvc of data log")
	}

	objectMeta = metav1.ObjectMeta{
		Name:            fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.LogVolumeSuffix),
		Namespace:       m.OBServer.Namespace,
		OwnerReferences: ownerReferenceList,
		Labels:          m.OBServer.Labels,
	}
	pvc = &corev1.PersistentVolumeClaim{
		ObjectMeta: objectMeta,
		Spec:       m.generatePVCSpec(fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.LogVolumeSuffix), m.OBServer.Spec.OBServerTemplate.Storage.LogStorage),
	}
	err = m.Client.Create(m.Ctx, pvc)
	if err != nil {
		return errors.Wrap(err, "Create pvc of log")
	}

	return nil
}

func (m *OBServerManager) createOBPodSpec(obcluster *v1alpha1.OBCluster) corev1.PodSpec {
	containers := make([]corev1.Container, 0)
	observerContainer := m.createOBServerContainer()
	containers = append(containers, observerContainer)

	// TODO, add monitor container
	volumeDataFile := corev1.Volume{}
	volumeDataFile.Name = fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.DataVolumeSuffix)
	volumeDataFileSource := &corev1.PersistentVolumeClaimVolumeSource{
		ClaimName: fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.DataVolumeSuffix),
	}
	volumeDataFile.VolumeSource.PersistentVolumeClaim = volumeDataFileSource

	volumeDataLog := corev1.Volume{}
	volumeDataLog.Name = fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.ClogVolumeSuffix)
	volumeDataLogSource := &corev1.PersistentVolumeClaimVolumeSource{
		ClaimName: fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.ClogVolumeSuffix),
	}
	volumeDataLog.VolumeSource.PersistentVolumeClaim = volumeDataLogSource

	volumeLog := corev1.Volume{}
	volumeLog.Name = fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.LogVolumeSuffix)
	volumeLogSource := &corev1.PersistentVolumeClaimVolumeSource{
		ClaimName: fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.LogVolumeSuffix),
	}
	volumeLog.VolumeSource.PersistentVolumeClaim = volumeLogSource

	volumes := make([]corev1.Volume, 0)
	volumes = append(volumes, volumeDataFile)
	volumes = append(volumes, volumeDataLog)
	volumes = append(volumes, volumeLog)

	if m.OBServer.Spec.BackupVolume != nil {
		volumes = append(volumes, *m.OBServer.Spec.BackupVolume.Volume)
	}

	if m.OBServer.Spec.MonitorTemplate != nil {
		monitorContainer := m.createMonitorContainer(obcluster)
		containers = append(containers, monitorContainer)
	}

	podSpec := corev1.PodSpec{
		Volumes:      volumes,
		Containers:   containers,
		NodeSelector: m.OBServer.Spec.NodeSelector,
	}
	return podSpec
}

func (m *OBServerManager) createMonitorContainer(obcluster *v1alpha1.OBCluster) corev1.Container {
	// port info
	ports := make([]corev1.ContainerPort, 0)
	httpPort := corev1.ContainerPort{}
	httpPort.Name = obagentconst.HttpPortName
	httpPort.ContainerPort = obagentconst.HttpPort
	httpPort.Protocol = corev1.ProtocolTCP
	pprofPort := corev1.ContainerPort{}
	pprofPort.Name = obagentconst.PprofPortName
	pprofPort.ContainerPort = obagentconst.PprofPort
	pprofPort.Protocol = corev1.ProtocolTCP
	ports = append(ports, httpPort)
	ports = append(ports, pprofPort)

	// resource info
	monagentResource := corev1.ResourceList{}
	monagentResource["cpu"] = m.OBServer.Spec.MonitorTemplate.Resource.Cpu
	monagentResource["memory"] = m.OBServer.Spec.MonitorTemplate.Resource.Memory
	resources := corev1.ResourceRequirements{
		Limits: monagentResource,
	}

	readinessProbeHTTP := corev1.HTTPGetAction{}
	readinessProbeHTTP.Port = intstr.FromInt(obagentconst.HttpPort)
	readinessProbeHTTP.Path = obagentconst.StatUrl
	readinessProbe := corev1.Probe{}
	readinessProbe.ProbeHandler.HTTPGet = &readinessProbeHTTP
	readinessProbe.PeriodSeconds = obagentconst.ProbeCheckPeriodSeconds
	readinessProbe.InitialDelaySeconds = obagentconst.ProbeCheckDelaySeconds

	env := make([]corev1.EnvVar, 0)
	envOBModuleStatus := corev1.EnvVar{
		Name:  obagentconst.EnvOBMonitorStatus,
		Value: obagentconst.ActiveStatus,
	}
	envClusterName := corev1.EnvVar{
		Name:  obagentconst.EnvClusterName,
		Value: m.OBServer.Spec.ClusterName,
	}
	envClusterId := corev1.EnvVar{
		Name:  obagentconst.EnvClusterId,
		Value: fmt.Sprintf("%d", m.OBServer.Spec.ClusterId),
	}
	envZoneName := corev1.EnvVar{
		Name:  obagentconst.EnvZoneName,
		Value: m.OBServer.Spec.Zone,
	}
	envMonitorUser := corev1.EnvVar{
		Name:  obagentconst.EnvMonitorUser,
		Value: obagentconst.MonitorUser,
	}
	envMonitorPassword := corev1.EnvVar{
		Name: obagentconst.EnvMonitorPASSWORD,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: obcluster.Spec.UserSecrets.Monitor,
				},
				Key: secretconst.PasswordKeyName,
			},
		},
	}
	env = append(env, envOBModuleStatus)
	env = append(env, envClusterName)
	env = append(env, envClusterId)
	env = append(env, envZoneName)
	env = append(env, envMonitorUser)
	env = append(env, envMonitorPassword)

	container := corev1.Container{
		Name:            obagentconst.ContainerName,
		Image:           m.OBServer.Spec.MonitorTemplate.Image,
		ImagePullPolicy: "IfNotPresent",
		Ports:           ports,
		Resources:       resources,
		ReadinessProbe:  &readinessProbe,
		WorkingDir:      obagentconst.InstallPath,
		Env:             env,
	}
	return container
}

// TODO move hardcoded values to another file
func (m *OBServerManager) createOBServerContainer() corev1.Container {
	// port info
	ports := make([]corev1.ContainerPort, 0)
	mysqlPort := corev1.ContainerPort{}
	mysqlPort.Name = oceanbaseconst.SqlPortName
	mysqlPort.ContainerPort = oceanbaseconst.SqlPort
	mysqlPort.Protocol = corev1.ProtocolTCP
	rpcPort := corev1.ContainerPort{}
	rpcPort.Name = oceanbaseconst.RpcPortName
	rpcPort.ContainerPort = oceanbaseconst.RpcPort
	rpcPort.Protocol = corev1.ProtocolTCP
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
	volumeMountDataFile.Name = fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.DataVolumeSuffix)
	volumeMountDataFile.MountPath = oceanbaseconst.DataPath
	volumeMountDataLog := corev1.VolumeMount{}
	volumeMountDataLog.Name = fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.ClogVolumeSuffix)
	volumeMountDataLog.MountPath = oceanbaseconst.ClogPath
	volumeMountLog := corev1.VolumeMount{}
	volumeMountLog.Name = fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.LogVolumeSuffix)
	volumeMountLog.MountPath = oceanbaseconst.LogPath

	volumeMounts := make([]corev1.VolumeMount, 0)
	volumeMounts = append(volumeMounts, volumeMountDataFile)
	volumeMounts = append(volumeMounts, volumeMountDataLog)
	volumeMounts = append(volumeMounts, volumeMountLog)

	if m.OBServer.Spec.BackupVolume != nil {
		volumeMountBackup := corev1.VolumeMount{}
		volumeMountBackup.Name = fmt.Sprintf(m.OBServer.Spec.BackupVolume.Volume.Name)
		volumeMountBackup.MountPath = oceanbaseconst.BackupPath
		volumeMounts = append(volumeMounts, volumeMountBackup)
	}

	readinessProbeTCP := corev1.TCPSocketAction{}
	readinessProbeTCP.Port = intstr.FromInt(oceanbaseconst.SqlPort)
	readinessProbe := corev1.Probe{}
	readinessProbe.ProbeHandler.TCPSocket = &readinessProbeTCP
	readinessProbe.PeriodSeconds = oceanbaseconst.ProbeCheckPeriodSeconds
	readinessProbe.InitialDelaySeconds = oceanbaseconst.ProbeCheckDelaySeconds

	startOBServerCmd := "/home/admin/oceanbase/bin/oceanbase-helper start"

	cmds := []string{
		"bash",
		"-c",
		startOBServerCmd,
	}

	// TODO make a new image take environment variables as commandline option
	env := make([]corev1.EnvVar, 0)
	envLib := corev1.EnvVar{
		Name:  "LD_LIBRARY_PATH",
		Value: "/home/admin/oceanbase/lib",
	}
	cpuCount := m.OBServer.Spec.OBServerTemplate.Resource.Cpu.Value()
	if cpuCount < 16 {
		cpuCount = 16
	}
	envCpu := corev1.EnvVar{
		Name:  "CPU_COUNT",
		Value: fmt.Sprintf("%d", cpuCount),
	}
	datafileSize, ok := m.OBServer.Spec.OBServerTemplate.Storage.DataStorage.Size.AsInt64()
	if !ok {
		m.Logger.Error(errors.New("Parse datafile size failed"), "failed to parse datafiel size")
	}
	envDataFile := corev1.EnvVar{
		Name:  "DATAFILE_SIZE",
		Value: fmt.Sprintf("%dG", datafileSize*oceanbaseconst.DefaultDiskUsePercent/oceanbaseconst.GigaConverter/100),
	}
	clogDiskSize, ok := m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size.AsInt64()
	if !ok {
		m.Logger.Error(errors.New("Parse log disk size failed"), "failed to parse log disk size")
	}
	envLogDisk := corev1.EnvVar{
		Name:  "LOG_DISK_SIZE",
		Value: fmt.Sprintf("%dG", clogDiskSize*oceanbaseconst.DefaultDiskUsePercent/oceanbaseconst.GigaConverter/100),
	}
	envClusterName := corev1.EnvVar{
		Name:  "CLUSTER_NAME",
		Value: m.OBServer.Spec.ClusterName,
	}
	envClusterId := corev1.EnvVar{
		Name:  "CLUSTER_ID",
		Value: fmt.Sprintf("%d", m.OBServer.Spec.ClusterId),
	}
	envZoneName := corev1.EnvVar{
		Name:  "ZONE_NAME",
		Value: m.OBServer.Spec.Zone,
	}
	env = append(env, envLib)
	env = append(env, envCpu)
	env = append(env, envDataFile)
	env = append(env, envLogDisk)
	env = append(env, envClusterName)
	env = append(env, envClusterId)
	env = append(env, envZoneName)

	container := corev1.Container{
		Name:            oceanbaseconst.ContainerName,
		Image:           m.OBServer.Spec.OBServerTemplate.Image,
		ImagePullPolicy: "IfNotPresent",
		Ports:           ports,
		Resources:       resources,
		VolumeMounts:    volumeMounts,
		ReadinessProbe:  &readinessProbe,
		WorkingDir:      oceanbaseconst.InstallPath,
		Env:             env,
		Command:         cmds,
	}
	return container
}

func (m *OBServerManager) DeleteOBServerInCluster() error {
	m.Logger.Info("delete observer in cluster")
	operationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrapf(err, "Get oceanbase operation manager failed")
	}
	observerInfo := &model.ServerInfo{
		Ip:   m.OBServer.Status.PodIp,
		Port: oceanbaseconst.RpcPort,
	}
	observer, err := operationManager.GetServer(observerInfo)
	if err != nil {
		return err
	}
	if observer != nil && observer.Status != "deleting" {
		if observer.Status == "deleting" {
			m.Logger.Info("observer is deleting", "observer", observerInfo.Ip)
		} else {
			m.Logger.Info("need to delete observer")
			err = operationManager.DeleteServer(observerInfo)
			if err != nil {
				return errors.Wrapf(err, "Failed to delete observer %s", m.OBServer.Status.PodIp)
			}
		}
	} else {
		m.Logger.Info("observer already deleted", "observer", observerInfo.Ip)
	}
	return nil
}

func (m *OBServerManager) AnnotateOBServerPod() error {
	observerPod, err := m.getPod()
	if err != nil {
		return errors.Wrapf(err, "Failed to get pod of observer %s", m.OBServer.Name)
	}
	if m.OBServer.Status.CNI == oceanbaseconst.CNICalico {
		m.Logger.Info("Update pod annotation, cni is calico")
		observerPod.Annotations[oceanbaseconst.AnnotationCalicoIpAddrs] = fmt.Sprintf("[\"%s\"]", m.OBServer.Status.PodIp)
	}
	err = m.Client.Update(m.Ctx, observerPod)
	if err != nil {
		return errors.Wrapf(err, "Failed to update pod annotation of observer %s", m.OBServer.Name)
	}
	return nil
}

func (m *OBServerManager) UpgradeOBServerImage() error {
	observerPod, err := m.getPod()
	if err != nil {
		return errors.Wrapf(err, "Failed to get pod of observer %s", m.OBServer.Name)
	}
	for idx, container := range observerPod.Spec.Containers {
		if container.Name == oceanbaseconst.ContainerName {
			observerPod.Spec.Containers[idx].Image = m.OBServer.Spec.OBServerTemplate.Image
			break
		}
	}
	err = m.Client.Update(m.Ctx, observerPod)
	if err != nil {
		return errors.Wrapf(err, "Failed to update pod of observer %s", m.OBServer.Name)
	}
	return nil
}

func (m *OBServerManager) WaitOBServerPodReady() error {
	observerPodRestarted := false
	for i := 0; i < oceanbaseconst.DefaultStateWaitTimeout; i++ {
		observerPod, err := m.getPod()
		if err != nil {
			return errors.Wrapf(err, "Failed to get pod of observer %s", m.OBServer.Name)
		}
		for _, containerStatus := range observerPod.Status.ContainerStatuses {
			if containerStatus.Name != oceanbaseconst.ContainerName {
				continue
			}
			if containerStatus.Ready && containerStatus.Image == m.OBServer.Spec.OBServerTemplate.Image {
				observerPodRestarted = true
			}
		}
		if observerPodRestarted {
			m.Logger.Info("observer pod restarted")
			break
		}
		time.Sleep(time.Second)
	}
	if !observerPodRestarted {
		return errors.Errorf("observer %s pod still not restart when timeout", m.OBServer.Name)
	}
	return nil
}

func (m *OBServerManager) WaitOBServerActiveInCluster() error {
	m.Logger.Info("wait observer active in cluster")
	observerInfo := &model.ServerInfo{
		Ip:   m.OBServer.Status.PodIp,
		Port: oceanbaseconst.RpcPort,
	}
	active := false
	for i := 0; i < oceanbaseconst.DefaultStateWaitTimeout; i++ {
		operationManager, err := m.getOceanbaseOperationManager()
		if err != nil {
			return errors.Wrapf(err, "Get oceanbase operation manager failed")
		}
		observer, _ := operationManager.GetServer(observerInfo)
		if observer != nil {
			if observer.StartServiceTime > 0 && observer.Status == observerstatus.Active {
				m.Logger.Info("Observer active")
				active = true
				break
			}
		} else {
			m.Logger.Info("OBServer is nil, check next time")
		}
		time.Sleep(time.Second)
	}
	if !active {
		m.Logger.Info("Wait observer active timeout")
		return errors.Errorf("Wait observer %s active timeout", observerInfo.Ip)
	}
	m.Logger.Info("observer active", "observer", observerInfo)
	return nil
}

func (m *OBServerManager) WaitOBServerDeletedInCluster() error {
	m.Logger.Info("wait observer deleted in cluster")
	observerInfo := &model.ServerInfo{
		Ip:   m.OBServer.Status.PodIp,
		Port: oceanbaseconst.RpcPort,
	}
	deleted := false
	for i := 0; i < oceanbaseconst.ServerDeleteTimeoutSeconds; i++ {
		operationManager, err := m.getOceanbaseOperationManager()
		if err != nil {
			return errors.Wrapf(err, "Get oceanbase operation manager failed")
		}
		observer, err := operationManager.GetServer(observerInfo)
		if observer == nil && err == nil {
			m.Logger.Info("Observer deleted")
			deleted = true
			break
		} else if err != nil {
			m.Logger.Error(err, "Query observer info failed")
		}
		time.Sleep(time.Second)
	}
	if !deleted {
		m.Logger.Info("Wait observer deleted timeout")
		return errors.Errorf("Wait observer %s deleted timeout", observerInfo.Ip)
	}
	m.Logger.Info("observer deleted", "observer", observerInfo)
	return nil
}
