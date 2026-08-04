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

package oblogservicenode

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	apipod "k8s.io/kubernetes/pkg/api/v1/pod"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	lsstatus "github.com/oceanbase/ob-operator/internal/const/status/oblogservicecluster"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/task/builder"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

var taskMap = builder.NewTaskHub[*OBLogServiceNodeManager]()

func CreateSvc(m *OBLogServiceNodeManager) tasktypes.TaskError {
	mode, modeAnnoExist := resourceutils.GetAnnotationField(m.Resource, oceanbaseconst.AnnotationsMode)
	if !modeAnnoExist || mode != oceanbaseconst.ModeService {
		m.Logger.Info("Skipping service creation (mode annotation not set to service)")
		return nil
	}
	m.Logger.Info("Creating log service node service")
	podName := m.Resource.Name
	svcName := fmt.Sprintf("%s-svc", podName)

	rpcPort := m.Resource.Spec.RpcPort
	if rpcPort == 0 {
		rpcPort = int32(oceanbaseconst.LogServiceRpcPort)
	}
	httpPort := m.Resource.Spec.HttpPort
	if httpPort == 0 {
		httpPort = int32(oceanbaseconst.LogServiceHttpPort)
	}

	blockOwnerDeletion := true
	ownerRef := metav1.OwnerReference{
		APIVersion:         m.Resource.APIVersion,
		Kind:               m.Resource.Kind,
		Name:               m.Resource.Name,
		UID:                m.Resource.GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
	}

	podLabels := m.buildPodLabels()
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            svcName,
			Namespace:       m.Resource.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerRef},
		},
		Spec: corev1.ServiceSpec{
			Selector: podLabels,
			Ports: []corev1.ServicePort{
				{Name: "rpc", Port: rpcPort, TargetPort: intstr.FromInt32(rpcPort)},
				{Name: "http", Port: httpPort, TargetPort: intstr.FromInt32(httpPort)},
			},
		},
	}
	if err := m.Client.Create(m.Ctx, svc); err != nil && !kubeerrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "create service")
	}

	existingSvc := &corev1.Service{}
	if err := m.Client.Get(m.Ctx, types.NamespacedName{Namespace: m.Resource.Namespace, Name: svcName}, existingSvc); err == nil {
		m.Resource.Status.ServiceIP = existingSvc.Spec.ClusterIP
	}
	return nil
}

func CreatePVC(m *OBLogServiceNodeManager) tasktypes.TaskError {
	m.Logger.Info("Creating log service node PVCs")

	if m.Resource.Spec.Storage == nil {
		return errors.New("storage is required but was nil")
	}
	if m.Resource.Spec.Storage.StoreStorage == nil {
		return errors.New("storage.storeStorage is required but was nil")
	}
	if m.Resource.Spec.Storage.LogStorage == nil {
		return errors.New("storage.logStorage is required but was nil")
	}

	// Snapshot the storage fields into locals after the nil checks so that any
	// hypothetical concurrent mutation of m.Resource.Spec.Storage between the
	// checks and the PVC construction cannot turn a non-nil check into a nil
	// deref (observed as an intermittent panic at the LogStorage.Size use).
	storeStorage := m.Resource.Spec.Storage.StoreStorage
	logStorage := m.Resource.Spec.Storage.LogStorage

	podName := m.Resource.Name
	storePvcName := fmt.Sprintf("%s-%s", podName, oceanbaseconst.LogServiceStoreVolumeSuffix)
	logPvcName := fmt.Sprintf("%s-%s", podName, oceanbaseconst.LogServiceLogVolumeSuffix)

	blockOwnerDeletion := true
	ownerRef := metav1.OwnerReference{
		APIVersion:         m.Resource.APIVersion,
		Kind:               m.Resource.Kind,
		Name:               m.Resource.Name,
		UID:                m.Resource.GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
	}

	storePvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:            storePvcName,
			Namespace:       m.Resource.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerRef},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": storeStorage.Size,
				},
			},
			StorageClassName: &storeStorage.StorageClass,
		},
	}
	if err := createOrValidateLogServicePVC(m, storePvc, "store"); err != nil {
		return err
	}

	logPvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:            logPvcName,
			Namespace:       m.Resource.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerRef},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": logStorage.Size,
				},
			},
			StorageClassName: &logStorage.StorageClass,
		},
	}
	return createOrValidateLogServicePVC(m, logPvc, "log")
}

func createOrValidateLogServicePVC(m *OBLogServiceNodeManager, desired *corev1.PersistentVolumeClaim, role string) error {
	if err := m.Client.Create(m.Ctx, desired); err != nil {
		if !kubeerrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "create %s pvc", role)
		}

		existing := &corev1.PersistentVolumeClaim{}
		key := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
		if m.APIReader == nil {
			return errors.Errorf("validate existing %s pvc %s: uncached Kubernetes API reader is not configured", role, key.String())
		}
		if err := m.APIReader.Get(m.Ctx, key, existing); err != nil {
			return errors.Wrapf(err, "get existing %s pvc %s", role, key.String())
		}
		if err := validateLogServicePVC(existing, desired, role); err != nil {
			return err
		}
	}
	return nil
}

func validateLogServicePVC(existing, desired *corev1.PersistentVolumeClaim, role string) error {
	if existing.DeletionTimestamp != nil {
		return errors.Errorf("existing %s pvc %s/%s is being deleted and cannot be reused", role, existing.Namespace, existing.Name)
	}
	if len(desired.OwnerReferences) == 0 {
		return errors.Errorf("cannot validate existing %s pvc %s/%s without the current OBLogServiceNode owner reference", role, existing.Namespace, existing.Name)
	}
	desiredOwner := desired.OwnerReferences[0]
	if desiredOwner.APIVersion == "" || desiredOwner.Kind == "" || desiredOwner.Name == "" || desiredOwner.UID == "" {
		return errors.Errorf("cannot validate existing %s pvc %s/%s with an incomplete OBLogServiceNode owner reference", role, existing.Namespace, existing.Name)
	}
	ownedByCurrentNode := false
	for _, owner := range existing.OwnerReferences {
		if owner.APIVersion == desiredOwner.APIVersion && owner.Kind == desiredOwner.Kind &&
			owner.Name == desiredOwner.Name && owner.UID == desiredOwner.UID {
			ownedByCurrentNode = true
			break
		}
	}
	if !ownedByCurrentNode {
		return errors.Errorf("existing %s pvc %s/%s is not owned by the current OBLogServiceNode %s (UID %q)", role, existing.Namespace, existing.Name, desiredOwner.Name, desiredOwner.UID)
	}

	if !storageClassNamesEqual(existing.Spec.StorageClassName, desired.Spec.StorageClassName) {
		return errors.Errorf("existing %s pvc %s/%s has storage class %q, expected %q", role, existing.Namespace, existing.Name, storageClassName(existing.Spec.StorageClassName), storageClassName(desired.Spec.StorageClassName))
	}

	existingSize, hasExistingSize := existing.Spec.Resources.Requests[corev1.ResourceStorage]
	desiredSize, hasDesiredSize := desired.Spec.Resources.Requests[corev1.ResourceStorage]
	// A PVC storage request is a minimum. Keep an already-expanded claim compatible,
	// but reject a smaller claim because CreatePVC does not expand it here.
	if !hasExistingSize || !hasDesiredSize || existingSize.Cmp(desiredSize) < 0 {
		return errors.Errorf("existing %s pvc %s/%s requests storage %q, expected at least %q", role, existing.Namespace, existing.Name, existingSize.String(), desiredSize.String())
	}

	if !accessModesEqual(existing.Spec.AccessModes, desired.Spec.AccessModes) {
		return errors.Errorf("existing %s pvc %s/%s has access modes %v, expected %v", role, existing.Namespace, existing.Name, existing.Spec.AccessModes, desired.Spec.AccessModes)
	}
	if volumeMode(existing) != volumeMode(desired) {
		return errors.Errorf("existing %s pvc %s/%s has volume mode %q, expected %q", role, existing.Namespace, existing.Name, volumeMode(existing), volumeMode(desired))
	}
	if existing.Spec.Selector != nil || existing.Spec.DataSource != nil || existing.Spec.DataSourceRef != nil {
		return errors.Errorf("existing %s pvc %s/%s uses a selector or data source and cannot be safely reused", role, existing.Namespace, existing.Name)
	}

	return nil
}

func storageClassNamesEqual(existing, desired *string) bool {
	if existing == nil || desired == nil {
		return existing == nil && desired == nil
	}
	return *existing == *desired
}

func storageClassName(name *string) string {
	if name == nil {
		return "<nil>"
	}
	return *name
}

func accessModesEqual(existing, desired []corev1.PersistentVolumeAccessMode) bool {
	if len(existing) != len(desired) {
		return false
	}

	modeCounts := make(map[corev1.PersistentVolumeAccessMode]int, len(existing))
	for _, mode := range existing {
		modeCounts[mode]++
	}
	for _, mode := range desired {
		modeCounts[mode]--
		if modeCounts[mode] < 0 {
			return false
		}
	}
	return true
}

func volumeMode(pvc *corev1.PersistentVolumeClaim) corev1.PersistentVolumeMode {
	if pvc.Spec.VolumeMode == nil {
		return corev1.PersistentVolumeFilesystem
	}
	return *pvc.Spec.VolumeMode
}

func CreatePod(m *OBLogServiceNodeManager) tasktypes.TaskError {
	m.Logger.Info("Creating log service node pod")

	if m.Resource.Spec.Resource == nil {
		return errors.New("resource is required but was nil")
	}
	if m.Resource.Spec.Resource.Memory.IsZero() {
		return errors.New("resource.memory is required but was zero")
	}
	for i, parameter := range m.Resource.Spec.Parameters {
		if oceanbaseconst.ContainsManagedParameter(parameter.Name, parameter.Value, oceanbaseconst.LogServiceManagedParameters[:]) {
			return errors.Errorf("spec.parameters[%d] contains an operator-managed parameter; managed parameters are %v", i, oceanbaseconst.LogServiceManagedParameters)
		}
	}

	podName := m.Resource.Name
	storePvcName := fmt.Sprintf("%s-%s", podName, oceanbaseconst.LogServiceStoreVolumeSuffix)
	logPvcName := fmt.Sprintf("%s-%s", podName, oceanbaseconst.LogServiceLogVolumeSuffix)

	rpcPort := m.Resource.Spec.RpcPort
	if rpcPort == 0 {
		rpcPort = int32(oceanbaseconst.LogServiceRpcPort)
	}
	httpPort := m.Resource.Spec.HttpPort
	if httpPort == 0 {
		httpPort = int32(oceanbaseconst.LogServiceHttpPort)
	}

	storeMountPath := oceanbaseconst.LogServiceStoreMountPath
	logMountPath := oceanbaseconst.LogServiceLogMountPath

	// In service mode, fetch ServiceIP from existing Service if not already in status
	mode, modeAnnoExist := resourceutils.GetAnnotationField(m.Resource, oceanbaseconst.AnnotationsMode)
	isServiceMode := modeAnnoExist && mode == oceanbaseconst.ModeService
	if isServiceMode && m.Resource.Status.ServiceIP == "" {
		svcName := fmt.Sprintf("%s-svc", podName)
		svc := &corev1.Service{}
		if err := m.Client.Get(m.Ctx, types.NamespacedName{Namespace: m.Resource.Namespace, Name: svcName}, svc); err == nil {
			m.Resource.Status.ServiceIP = svc.Spec.ClusterIP
		} else {
			m.Logger.Error(err, "Failed to get service for ServiceIP lookup", "service", svcName)
		}
	}
	// Service mode uses ServiceIP as advertise address; pod-IP mode uses downward API
	var localIPEnv corev1.EnvVar
	if isServiceMode {
		advertiseIP := m.Resource.Status.GetConnectAddr()
		if advertiseIP == "" {
			return errors.New("logservice node has no ServiceIP; cannot determine advertise address")
		}
		localIPEnv = corev1.EnvVar{Name: "LOCAL_IP", Value: advertiseIP}
	} else {
		localIPEnv = corev1.EnvVar{
			Name: "LOCAL_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
			},
		}
	}

	// Calculate log_disk_size from storeStorage PVC size
	logDiskSizeEnv := ""
	if m.Resource.Spec.Storage != nil && m.Resource.Spec.Storage.StoreStorage != nil {
		storeSizeBytes, ok := m.Resource.Spec.Storage.StoreStorage.Size.AsInt64()
		if ok && storeSizeBytes > 0 {
			logDiskSizeG := storeSizeBytes * int64(obcfg.GetConfig().Resource.DefaultDiskUsePercent) / int64(oceanbaseconst.GigaConverter) / 100
			if logDiskSizeG > 0 {
				logDiskSizeEnv = fmt.Sprintf("%dG", logDiskSizeG)
			}
		}
	}

	// Build extra parameters from spec (excluding reserved ones)
	var extraParams []string
	for _, p := range m.Resource.Spec.Parameters {
		reserved := false
		for _, rp := range oceanbaseconst.LogServiceReservedParameters {
			if p.Name == rp {
				reserved = true
				break
			}
		}
		if !reserved {
			extraParams = append(extraParams, fmt.Sprintf("%s=%s", p.Name, p.Value))
		}
	}

	blockOwnerDeletion := true
	ownerRef := metav1.OwnerReference{
		APIVersion:         m.Resource.APIVersion,
		Kind:               m.Resource.Kind,
		Name:               m.Resource.Name,
		UID:                m.Resource.GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
	}

	podLabels := m.buildPodLabels()
	podAnnotations := m.generateStaticIpAnnotation()

	envVars := []corev1.EnvVar{
		{Name: "CLUSTER_ID", Value: fmt.Sprintf("%d", m.Resource.Spec.ClusterId)},
		localIPEnv,
		{Name: "RPC_PORT", Value: fmt.Sprintf("%d", rpcPort)},
		{Name: "HTTP_PORT", Value: fmt.Sprintf("%d", httpPort)},
		{Name: "STORE_MOUNT_PATH", Value: storeMountPath},
		{Name: "LOG_MOUNT_PATH", Value: logMountPath},
	}
	if logDiskSizeEnv != "" {
		envVars = append(envVars, corev1.EnvVar{Name: "LOG_DISK_SIZE", Value: logDiskSizeEnv})
	}
	if len(extraParams) > 0 {
		envVars = append(envVars, corev1.EnvVar{Name: "EXTRA_PARAMETERS", Value: strings.Join(extraParams, ",")})
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            podName,
			Namespace:       m.Resource.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerRef},
			Labels:          podLabels,
			Annotations:     podAnnotations,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: m.Resource.Spec.ServiceAccount,
			Containers: []corev1.Container{{
				Name:            "oblogservice",
				Image:           m.Resource.Spec.Image,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"/home/admin/oblogservice/bin/oceanbase-helper", "logservice", "start"},
				Env:             envVars,
				Ports: []corev1.ContainerPort{
					{Name: "rpc", ContainerPort: rpcPort, Protocol: corev1.ProtocolTCP},
					{Name: "http", ContainerPort: httpPort, Protocol: corev1.ProtocolTCP},
				},
				Resources: func() corev1.ResourceRequirements {
					nodeResource := m.Resource.Spec.Resource
					resList := corev1.ResourceList{
						corev1.ResourceMemory: nodeResource.Memory,
					}
					if !nodeResource.Cpu.IsZero() {
						resList[corev1.ResourceCPU] = nodeResource.Cpu
					}
					return corev1.ResourceRequirements{
						Requests: resList,
						Limits:   resList,
					}
				}(),
				VolumeMounts: []corev1.VolumeMount{
					{Name: oceanbaseconst.LogServiceStoreVolumeName, MountPath: storeMountPath},
					{Name: oceanbaseconst.LogServiceLogVolumeName, MountPath: logMountPath},
				},
				ReadinessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						TCPSocket: &corev1.TCPSocketAction{
							Port: intstr.FromInt32(httpPort),
						},
					},
					InitialDelaySeconds: 10,
					PeriodSeconds:       5,
				},
			}},
			Volumes: []corev1.Volume{
				{
					Name: oceanbaseconst.LogServiceStoreVolumeName,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: storePvcName},
					},
				},
				{
					Name: oceanbaseconst.LogServiceLogVolumeName,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: logPvcName},
					},
				},
			},
			NodeSelector: m.Resource.Spec.NodeSelector,
			Affinity:     m.Resource.Spec.Affinity,
			Tolerations:  m.Resource.Spec.Tolerations,
		},
	}
	if err := m.Client.Create(m.Ctx, pod); err != nil && !kubeerrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "create pod")
	}

	m.Resource.Status.PodName = podName
	return nil
}

func WaitPodReady(m *OBLogServiceNodeManager) tasktypes.TaskError {
	m.Logger.Info("Waiting for log service node pod to be ready")
	podName := m.Resource.Status.PodName
	if podName == "" {
		podName = m.Resource.Name
	}

	for range 600 {
		pod := &corev1.Pod{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.Resource.Namespace,
			Name:      podName,
		}, pod)
		if err != nil {
			return errors.Wrap(err, "get pod")
		}
		if pod.Status.Phase == corev1.PodRunning && pod.Status.PodIP != "" && apipod.IsPodReady(pod) {
			m.Resource.Status.PodIP = pod.Status.PodIP
			m.Resource.Status.CNI = resourceutils.GetCNIFromAnnotation(pod)
			m.Resource.Status.Ready = true
			m.Resource.Status.PodPhase = pod.Status.Phase
			if m.Resource.Status.ServiceIP == "" {
				mode, modeExist := resourceutils.GetAnnotationField(m.Resource, oceanbaseconst.AnnotationsMode)
				if modeExist && mode == oceanbaseconst.ModeService {
					svcName := fmt.Sprintf("%s-svc", podName)
					svc := &corev1.Service{}
					if svcErr := m.Client.Get(m.Ctx, types.NamespacedName{
						Namespace: m.Resource.Namespace,
						Name:      svcName,
					}, svc); svcErr == nil {
						m.Resource.Status.ServiceIP = svc.Spec.ClusterIP
					}
				}
			}
			m.Logger.Info("Log service node pod is ready", "pod", podName, "ip", pod.Status.PodIP)
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("timeout waiting for log service node pod to be ready")
}

func WaitClusterBootstrapped(m *OBLogServiceNodeManager) tasktypes.TaskError {
	m.Logger.Info("Waiting for log service cluster to finish bootstrap")
	clusterName := m.Resource.Spec.ClusterName

	for range 600 {
		lsCluster := &v1alpha1.OBLogServiceCluster{}
		err := m.Client.Get(m.Ctx, client.ObjectKey{
			Namespace: m.Resource.Namespace,
			Name:      clusterName,
		}, lsCluster)
		if err != nil {
			return errors.Wrap(err, "get log service cluster")
		}
		if lsCluster.Status.Status == lsstatus.Running {
			m.Logger.Info("Log service cluster bootstrap completed")
			return nil
		}
		if lsCluster.Status.Status == lsstatus.Failed {
			return errors.New("log service cluster bootstrap failed")
		}
		time.Sleep(time.Second)
	}
	return errors.New("timeout waiting for log service cluster bootstrap")
}

func DeletePod(m *OBLogServiceNodeManager) tasktypes.TaskError {
	m.Logger.Info("Deleting log service node pod")
	podName := m.Resource.Status.PodName
	if podName == "" {
		podName = m.Resource.Name
	}
	pod := &corev1.Pod{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.Resource.Namespace,
		Name:      podName,
	}, pod)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil
		}
		return errors.Wrap(err, "get pod for deletion")
	}
	if err := m.Client.Delete(m.Ctx, pod); err != nil && !kubeerrors.IsNotFound(err) {
		return errors.Wrap(err, "delete pod")
	}
	// Keep Status.PodIP: generateStaticIpAnnotation needs the old IP to pin
	// the recreated pod to the same address (Calico/KubeOvn), so the ln
	// registration recorded in logservice metadata stays valid.
	m.Resource.Status.Ready = false
	return nil
}

func WaitPodDeleted(m *OBLogServiceNodeManager) tasktypes.TaskError {
	m.Logger.Info("Waiting for log service node pod to be deleted")
	podName := m.Resource.Status.PodName
	if podName == "" {
		podName = m.Resource.Name
	}
	for range obcfg.GetConfig().Time.DefaultStateWaitTimeout {
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.Resource.Namespace,
			Name:      podName,
		}, &corev1.Pod{})
		if err != nil && kubeerrors.IsNotFound(err) {
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("timeout waiting for log service node pod to be deleted")
}

func (m *OBLogServiceNodeManager) buildPodLabels() map[string]string {
	return map[string]string{
		"app": "oblogservice",
		oceanbaseconst.LabelRefOBLogServiceCluster: m.Resource.Spec.ClusterName,
		oceanbaseconst.LabelRefOBLogServiceZone:    fmt.Sprintf("%s-%s", m.Resource.Spec.ClusterName, m.Resource.Spec.Zone),
		"oblogservice-node":                        m.Resource.Name,
	}
}

func (m *OBLogServiceNodeManager) generateStaticIpAnnotation() map[string]string {
	annotations := make(map[string]string)
	switch m.Resource.Status.CNI {
	case oceanbaseconst.CNICalico:
		if m.Resource.Status.PodIP != "" {
			annotations[oceanbaseconst.AnnotationCalicoIpAddrs] = fmt.Sprintf("[\"%s\"]", m.Resource.Status.PodIP)
		}
	case oceanbaseconst.CNIKubeOvn:
		if m.Resource.Status.PodIP != "" {
			annotations[oceanbaseconst.AnnotationKubeOvnIpAddrs] = m.Resource.Status.PodIP
		}
	default:
	}
	return annotations
}
