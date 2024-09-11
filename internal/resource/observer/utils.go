/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package observer

import (
	stderrs "errors"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	obagentconst "github.com/oceanbase/ob-operator/internal/const/obagent"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	secretconst "github.com/oceanbase/ob-operator/internal/const/secret"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

// get observer from K8s api server
func (m *OBServerManager) getOBServer() (*v1alpha1.OBServer, error) {
	// this label always exists
	observer := &v1alpha1.OBServer{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(m.OBServer.Name), observer)
	if err != nil {
		return nil, errors.Wrap(err, "get observer")
	}
	return observer, nil
}

func (m *OBServerManager) getOBZone() (*v1alpha1.OBZone, error) {
	// this label always exists
	zoneName, _ := m.OBServer.Labels[oceanbaseconst.LabelRefOBZone]
	obzone := &v1alpha1.OBZone{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(zoneName), obzone)
	if err != nil {
		return nil, errors.Wrap(err, "get obzone")
	}
	return obzone, nil
}

func (m *OBServerManager) generateNamespacedName(name string) types.NamespacedName {
	var namespacedName types.NamespacedName
	namespacedName.Namespace = m.OBServer.Namespace
	namespacedName.Name = name
	return namespacedName
}

func (m *OBServerManager) getPod() (*corev1.Pod, error) {
	// this label always exists
	pod := &corev1.Pod{}
	err := m.K8sResClient.Get(m.Ctx, m.generateNamespacedName(m.OBServer.Name), pod)
	if err != nil {
		return nil, errors.Wrap(err, "get pod")
	}
	return pod, nil
}

func (m *OBServerManager) getSvc() (*corev1.Service, error) {
	svc := &corev1.Service{}
	err := m.K8sResClient.Get(m.Ctx, m.generateNamespacedName(m.OBServer.Name), svc)
	if err != nil {
		return nil, errors.Wrap(err, "get svc")
	}
	return svc, nil
}

func (m *OBServerManager) getOBCluster() (*v1alpha1.OBCluster, error) {
	// this label always exists
	clusterName, _ := m.OBServer.Labels[oceanbaseconst.LabelRefOBCluster]
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(clusterName), obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	return obcluster, nil
}

func (m *OBServerManager) getCurrentOBServerFromOB() (*model.OBServer, error) {
	if m.OBServer.Status.PodIp == "" {
		err := errors.New("pod ip is empty")
		m.Logger.Error(err, "unable to get observer info")
		return nil, err
	}
	observerInfo := &model.ServerInfo{
		Ip:   m.OBServer.Status.GetConnectAddr(),
		Port: oceanbaseconst.RpcPort,
	}
	mode, modeExist := resourceutils.GetAnnotationField(m.OBServer, oceanbaseconst.AnnotationsMode)
	if modeExist && mode == oceanbaseconst.ModeStandalone {
		observerInfo.Ip = "127.0.0.1"
	}
	operationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return nil, errors.Wrapf(err, "Get oceanbase operation manager failed")
	}
	return operationManager.GetServer(m.Ctx, observerInfo)
}

func (m *OBServerManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		observer, err := m.getOBServer()
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		observer.Status = *m.OBServer.Status.DeepCopy()
		return m.Client.Status().Update(m.Ctx, observer)
	})
}

func (m *OBServerManager) setRecoveryStatus() {
	if m.OBServer.SupportStaticIP() {
		m.Logger.Info("Current server can keep static ip address or the cluster runs as standalone, recover by recreating pod")
		m.OBServer.Status.Status = serverstatus.Recover
	} else {
		m.Logger.Info("OBServer is not recoverable, delete current observer and wait recreate")
		m.OBServer.Status.Status = serverstatus.Unrecoverable
	}
}

func (m *OBServerManager) getPVCs() (*corev1.PersistentVolumeClaimList, error) {
	pvcs := &corev1.PersistentVolumeClaimList{}
	err := m.K8sResClient.List(m.Ctx, pvcs, client.InNamespace(m.OBServer.Namespace), client.MatchingLabels{oceanbaseconst.LabelRefUID: m.OBServer.Labels[oceanbaseconst.LabelRefUID]})
	if err != nil {
		return nil, errors.Wrap(err, "list pvc")
	}
	return pvcs, nil
}

func (m *OBServerManager) checkIfStorageExpand(pvcs *corev1.PersistentVolumeClaimList) bool {
	for _, pvc := range pvcs.Items {
		switch {
		case strings.HasSuffix(pvc.Name, oceanbaseconst.DataVolumeSuffix):
			if pvc.Spec.Resources.Requests.Storage().Cmp(m.OBServer.Spec.OBServerTemplate.Storage.DataStorage.Size) < 0 {
				return true
			}
		case strings.HasSuffix(pvc.Name, oceanbaseconst.ClogVolumeSuffix):
			if pvc.Spec.Resources.Requests.Storage().Cmp(m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size) < 0 {
				return true
			}
		case strings.HasSuffix(pvc.Name, oceanbaseconst.LogVolumeSuffix):
			if pvc.Spec.Resources.Requests.Storage().Cmp(m.OBServer.Spec.OBServerTemplate.Storage.LogStorage.Size) < 0 {
				return true
			}
		case pvc.Name == m.OBServer.Name:
			sum := resource.Quantity{}
			sum.Add(m.OBServer.Spec.OBServerTemplate.Storage.DataStorage.Size)
			sum.Add(m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size)
			sum.Add(m.OBServer.Spec.OBServerTemplate.Storage.LogStorage.Size)
			if pvc.Spec.Resources.Requests.Storage().Cmp(sum) < 0 {
				return true
			}
		}
	}
	return false
}

func (m *OBServerManager) checkIfResourceChanged(pod *corev1.Pod) bool {
	if len(pod.Spec.Containers) > 0 {
		tmplRes := m.OBServer.Spec.OBServerTemplate.Resource
		for i, container := range pod.Spec.Containers {
			if container.Name == oceanbaseconst.ContainerName {
				containerRes := pod.Spec.Containers[i].Resources.Limits
				if containerRes.Cpu().Cmp(tmplRes.Cpu) != 0 || containerRes.Memory().Cmp(tmplRes.Memory) != 0 {
					return true
				}
			}
		}
	}
	return false
}

func (m *OBServerManager) checkIfBackupVolumeMutated(pod *corev1.Pod) bool {
	addingVolume := m.OBServer.Spec.BackupVolume != nil
	volumeExist := false

	for _, container := range pod.Spec.Containers {
		if container.Name == oceanbaseconst.ContainerName {
			for _, volumeMount := range container.VolumeMounts {
				if volumeMount.MountPath == oceanbaseconst.BackupPath {
					volumeExist = true
					break
				}
			}
		}
	}

	return addingVolume != volumeExist
}

func (m *OBServerManager) checkIfMonitorMutated(pod *corev1.Pod) bool {
	addingMonitor := m.OBServer.Spec.MonitorTemplate != nil
	monitorExist := false
	for _, container := range pod.Spec.Containers {
		if container.Name == obagentconst.ContainerName {
			monitorExist = true
		}
	}
	return addingMonitor != monitorExist
}

func (m *OBServerManager) generatePVCSpec(storageSpec *apitypes.StorageSpec) corev1.PersistentVolumeClaimSpec {
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

func (m *OBServerManager) createOBPodSpec(obcluster *v1alpha1.OBCluster) corev1.PodSpec {
	containers := make([]corev1.Container, 0)
	observerContainer := m.createOBServerContainer(obcluster)
	containers = append(containers, observerContainer)

	volumes := make([]corev1.Volume, 0)

	singlePvcAnnoVal, singlePvcExist := resourceutils.GetAnnotationField(m.OBServer, oceanbaseconst.AnnotationsSinglePVC)
	if singlePvcExist && singlePvcAnnoVal == "true" {
		singleVolume := corev1.Volume{}
		singleVolumeSource := &corev1.PersistentVolumeClaimVolumeSource{}
		singleVolume.Name = m.OBServer.Name
		singleVolumeSource.ClaimName = m.OBServer.Name
		singleVolume.VolumeSource.PersistentVolumeClaim = singleVolumeSource

		volumes = append(volumes, singleVolume)
	} else {
		volumeDataFile := corev1.Volume{}
		volumeDataFileSource := &corev1.PersistentVolumeClaimVolumeSource{}
		volumeDataFile.Name = fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.DataVolumeSuffix)
		volumeDataFileSource.ClaimName = fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.DataVolumeSuffix)
		volumeDataFile.VolumeSource.PersistentVolumeClaim = volumeDataFileSource

		volumeDataLog := corev1.Volume{}
		volumeDataLog.Name = fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.ClogVolumeSuffix)
		volumeDataLogSource := &corev1.PersistentVolumeClaimVolumeSource{}
		volumeDataLogSource.ClaimName = fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.ClogVolumeSuffix)
		volumeDataLog.VolumeSource.PersistentVolumeClaim = volumeDataLogSource

		volumeLog := corev1.Volume{}
		volumeLog.Name = fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.LogVolumeSuffix)
		volumeLogSource := &corev1.PersistentVolumeClaimVolumeSource{}
		volumeLogSource.ClaimName = fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.LogVolumeSuffix)
		volumeLog.VolumeSource.PersistentVolumeClaim = volumeLogSource

		volumes = append(volumes, volumeDataFile, volumeDataLog, volumeLog)
	}

	if m.OBServer.Spec.BackupVolume != nil {
		volumes = append(volumes, *m.OBServer.Spec.BackupVolume.Volume)
	}

	if m.OBServer.Spec.MonitorTemplate != nil {
		monitorContainer := m.createMonitorContainer(obcluster)
		containers = append(containers, monitorContainer)
	}

	podSpec := corev1.PodSpec{
		Volumes:            volumes,
		Containers:         containers,
		NodeSelector:       m.OBServer.Spec.NodeSelector,
		Affinity:           m.OBServer.Spec.Affinity,
		Tolerations:        m.OBServer.Spec.Tolerations,
		ServiceAccountName: m.OBServer.Spec.ServiceAccount,
		SchedulerName:      resourceutils.GetSchedulerName(m.OBServer.Spec.OBServerTemplate.PodFields),
	}
	podFields := m.OBServer.Spec.OBServerTemplate.PodFields
	if podFields != nil {
		if podFields.PriorityClassName != nil && *podFields.PriorityClassName != "" {
			podSpec.PriorityClassName = *podFields.PriorityClassName
		}
		if podFields.RuntimeClassName != nil && *podFields.RuntimeClassName != "" {
			podSpec.RuntimeClassName = podFields.RuntimeClassName
		}
		if podFields.PreemptionPolicy != nil && *podFields.PreemptionPolicy != "" {
			podSpec.PreemptionPolicy = podFields.PreemptionPolicy
		}
		if podFields.Priority != nil {
			podSpec.Priority = podFields.Priority
		}
		if podFields.SecurityContext != nil {
			podSpec.SecurityContext = podFields.SecurityContext
		}
		if podFields.DNSPolicy != nil && *podFields.DNSPolicy != "" {
			podSpec.DNSPolicy = *podFields.DNSPolicy
		}
	}
	return podSpec
}

func (m *OBServerManager) getVarsReplacer(obcluster *v1alpha1.OBCluster) *strings.Replacer {
	replacePairs := []string{
		"${observer-name}", m.OBServer.Name,
		"${obzone-name}", m.OBServer.Labels[oceanbaseconst.LabelRefOBZone],
		"${obcluster-name}", m.OBServer.Labels[oceanbaseconst.LabelRefOBZone],
	}

	if obcluster != nil {
		replacePairs = append(replacePairs,
			"${obcluster-cluster-name}", obcluster.Spec.ClusterName,
			"${obcluster-cluster-id}", string(obcluster.Spec.ClusterId),
		)
	}
	return strings.NewReplacer(replacePairs...)
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
	monagentResource["memory"] = m.OBServer.Spec.MonitorTemplate.Resource.Memory
	if !m.OBServer.Spec.MonitorTemplate.Resource.Cpu.IsZero() {
		monagentResource["cpu"] = m.OBServer.Spec.MonitorTemplate.Resource.Cpu
	}
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
func (m *OBServerManager) createOBServerContainer(obcluster *v1alpha1.OBCluster) corev1.Container {
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
	infoPort := corev1.ContainerPort{}
	infoPort.Name = "info"
	infoPort.ContainerPort = 8080
	infoPort.Protocol = corev1.ProtocolTCP

	ports = append(ports, mysqlPort)
	ports = append(ports, rpcPort)
	ports = append(ports, infoPort)

	// resource info
	observerResource := corev1.ResourceList{}
	observerResource["memory"] = m.OBServer.Spec.OBServerTemplate.Resource.Memory
	if !m.OBServer.Spec.OBServerTemplate.Resource.Cpu.IsZero() {
		observerResource["cpu"] = m.OBServer.Spec.OBServerTemplate.Resource.Cpu
	}
	resources := corev1.ResourceRequirements{
		Requests: observerResource,
		Limits:   observerResource,
	}

	// volume mounts
	volumeMountDataFile := corev1.VolumeMount{}
	volumeMountDataFile.MountPath = oceanbaseconst.DataPath
	volumeMountDataLog := corev1.VolumeMount{}
	volumeMountDataLog.MountPath = oceanbaseconst.ClogPath
	volumeMountLog := corev1.VolumeMount{}
	volumeMountLog.MountPath = oceanbaseconst.LogPath

	// set subpath
	singlePvcAnnoVal, singlePvcExist := resourceutils.GetAnnotationField(m.OBServer, oceanbaseconst.AnnotationsSinglePVC)
	if singlePvcExist && singlePvcAnnoVal == "true" {
		volumeMountDataFile.Name = m.OBServer.Name
		volumeMountDataLog.Name = m.OBServer.Name
		volumeMountLog.Name = m.OBServer.Name
		volumeMountDataFile.SubPath = oceanbaseconst.DataVolumeSuffix
		volumeMountDataLog.SubPath = oceanbaseconst.ClogVolumeSuffix
		volumeMountLog.SubPath = oceanbaseconst.LogVolumeSuffix
	} else {
		volumeMountDataFile.Name = fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.DataVolumeSuffix)
		volumeMountDataLog.Name = fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.ClogVolumeSuffix)
		volumeMountLog.Name = fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.LogVolumeSuffix)
	}

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
	readinessProbe.PeriodSeconds = int32(obcfg.GetConfig().Time.ProbeCheckPeriodSeconds)
	readinessProbe.InitialDelaySeconds = int32(obcfg.GetConfig().Time.ProbeCheckDelaySeconds)
	readinessProbe.FailureThreshold = 32

	startOBServerCmd := "/home/admin/oceanbase/bin/oceanbase-helper start"

	cmds := []string{
		"bash",
		"-c",
		startOBServerCmd,
	}

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
		m.Logger.Error(errors.New("Parse datafile size failed"), "failed to parse datafile size")
	}
	envDataFile := corev1.EnvVar{
		Name:  "DATAFILE_SIZE",
		Value: fmt.Sprintf("%dG", datafileSize*int64(obcfg.GetConfig().Resource.InitialDataDiskUsePercent)/oceanbaseconst.GigaConverter/100),
	}
	clogDiskSize, ok := m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size.AsInt64()
	if !ok {
		m.Logger.Error(errors.New("Parse log disk size failed"), "failed to parse log disk size")
	}
	envLogDisk := corev1.EnvVar{
		Name:  "LOG_DISK_SIZE",
		Value: fmt.Sprintf("%dG", clogDiskSize*int64(obcfg.GetConfig().Resource.DefaultDiskUsePercent)/oceanbaseconst.GigaConverter/100),
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

	mode, modeAnnoExist := resourceutils.GetAnnotationField(m.OBServer, oceanbaseconst.AnnotationsMode)
	if modeAnnoExist {
		switch mode {
		case oceanbaseconst.ModeStandalone:
			envMode := corev1.EnvVar{
				Name:  "STANDALONE",
				Value: oceanbaseconst.ModeStandalone,
			}
			env = append(env, envMode)
		case oceanbaseconst.ModeService:
			svc, err := m.getSvc()
			if err != nil {
				if kubeerrors.IsNotFound(err) {
					m.Logger.Info("Svc not found")
				} else {
					m.Logger.Error(err, "Failed to get svc")
				}
			} else {
				envSvcIp := corev1.EnvVar{
					Name:  "SVC_IP",
					Value: svc.Spec.ClusterIP,
				}
				env = append(env, envSvcIp)
			}
		}
	}

	startupParameters := make([]string, 0)
	for _, parameter := range obcluster.Spec.Parameters {
		reserved := false
		for _, reservedParameter := range oceanbaseconst.ReservedParameters {
			if parameter.Name == reservedParameter {
				reserved = true
				break
			}
		}
		if !reserved {
			startupParameters = append(startupParameters, fmt.Sprintf("%s='%s'", parameter.Name, parameter.Value))
		}
	}
	if len(startupParameters) != 0 {
		envExtraOpt := corev1.EnvVar{
			Name:  "EXTRA_OPTION",
			Value: strings.Join(startupParameters, ","),
		}
		env = append(env, envExtraOpt)
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

func (m *OBServerManager) generateStaticIpAnnotation() map[string]string {
	annotations := make(map[string]string)
	switch m.OBServer.Status.CNI {
	case oceanbaseconst.CNICalico:
		if m.OBServer.Status.PodIp != "" {
			annotations[oceanbaseconst.AnnotationCalicoIpAddrs] = fmt.Sprintf("[\"%s\"]", m.OBServer.Status.PodIp)
		}
	default:
		m.Logger.Info("No CNI is configured, set empty annotation")
	}
	return annotations
}

func (m *OBServerManager) getOceanbaseOperationManager() (*operation.OceanbaseOperationManager, error) {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return nil, errors.Wrap(err, "Get obcluster from K8s")
	}
	return resourceutils.GetSysOperationClient(m.Client, m.Logger, obcluster)
}

func (m *OBServerManager) inMasterK8s() bool {
	return m.OBServer.Spec.K8sCluster == ""
}

func (m *OBServerManager) cleanWorkerK8sResource() error {
	var errs error

	// delete svc
	svc, err := m.getSvc()
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			m.Logger.Info("Svc not found")
		} else {
			errs = stderrs.Join(errs, errors.Wrap(err, "Failed to get svc"))
		}
	} else {
		if err := m.K8sResClient.Delete(m.Ctx, svc); err != nil {
			errs = stderrs.Join(errs, errors.Wrap(err, "Failed to delete svc"))
		}
	}

	// delete pod
	pod, err := m.getPod()
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			m.Logger.Info("Pod not found")
		} else {
			errs = stderrs.Join(errs, errors.Wrap(err, "Failed to get pod"))
		}
	} else {
		if err := m.K8sResClient.Delete(m.Ctx, pod); err != nil {
			errs = stderrs.Join(errs, errors.Wrap(err, "Failed to delete pod"))
		}
	}

	// delete pvc
	pvc := &corev1.PersistentVolumeClaim{}
	if err := m.K8sResClient.DeleteAllOf(m.Ctx, pvc,
		client.InNamespace(m.OBServer.Namespace),
		client.MatchingLabels{oceanbaseconst.LabelRefUID: m.OBServer.Labels[oceanbaseconst.LabelRefUID]},
	); err != nil {
		errs = stderrs.Join(errs, errors.Wrap(err, "Failed to delete pvc"))
	}

	return errs
}
