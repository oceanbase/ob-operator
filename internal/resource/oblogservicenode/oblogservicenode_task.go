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

	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/task/builder"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

var taskMap = builder.NewTaskHub[*OBLogServiceNodeManager]()

func CreatePod(m *OBLogServiceNodeManager) tasktypes.TaskError {
	m.Logger.Info("Creating log service node pod")
	blockOwnerDeletion := true
	ownerRef := metav1.OwnerReference{
		APIVersion:         m.Resource.APIVersion,
		Kind:               m.Resource.Kind,
		Name:               m.Resource.Name,
		UID:                m.Resource.GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
	}

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

	storePvcName := fmt.Sprintf("%s-%s", podName, oceanbaseconst.LogServiceStoreVolumeSuffix)
	logPvcName := fmt.Sprintf("%s-%s", podName, oceanbaseconst.LogServiceLogVolumeSuffix)

	if m.Resource.Spec.Storage == nil {
		return errors.New("storage is required but was nil")
	}
	if m.Resource.Spec.Storage.StoreStorage == nil {
		return errors.New("storage.storeStorage is required but was nil")
	}
	if m.Resource.Spec.Storage.LogStorage == nil {
		return errors.New("storage.logStorage is required but was nil")
	}
	if m.Resource.Spec.Resource == nil {
		return errors.New("resource is required but was nil")
	}
	if m.Resource.Spec.Resource.Memory.IsZero() {
		return errors.New("resource.memory is required but was zero")
	}

	// Create store PVC
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
					"storage": m.Resource.Spec.Storage.StoreStorage.Size,
				},
			},
			StorageClassName: &m.Resource.Spec.Storage.StoreStorage.StorageClass,
		},
	}
	if err := m.Client.Create(m.Ctx, storePvc); err != nil && !kubeerrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "create store pvc")
	}

	// Create log PVC
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
					"storage": m.Resource.Spec.Storage.LogStorage.Size,
				},
			},
			StorageClassName: &m.Resource.Spec.Storage.LogStorage.StorageClass,
		},
	}
	if err := m.Client.Create(m.Ctx, logPvc); err != nil && !kubeerrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "create log pvc")
	}

	// Create per-pod Service for stable network identity
	podLabels := map[string]string{
		"app": "oblogservice",
		oceanbaseconst.LabelRefOBLogServiceCluster: m.Resource.Spec.ClusterName,
		oceanbaseconst.LabelRefOBLogServiceZone:    fmt.Sprintf("%s-%s", m.Resource.Spec.ClusterName, m.Resource.Spec.Zone),
		"oblogservice-node":                        podName,
	}
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

	// Retrieve Service ClusterIP
	existingSvc := &corev1.Service{}
	if err := m.Client.Get(m.Ctx, types.NamespacedName{Namespace: m.Resource.Namespace, Name: svcName}, existingSvc); err == nil {
		m.Resource.Status.ServiceIP = existingSvc.Spec.ClusterIP
	}

	storeMountPath := oceanbaseconst.LogServiceStoreMountPath
	logMountPath := oceanbaseconst.LogServiceLogMountPath

	// Calculate log_disk_size from storeStorage PVC size, aligned with OBCluster's approach
	logDiskSizeParam := ""
	if m.Resource.Spec.Storage.StoreStorage != nil {
		storeSizeBytes, ok := m.Resource.Spec.Storage.StoreStorage.Size.AsInt64()
		if ok && storeSizeBytes > 0 {
			logDiskSizeG := storeSizeBytes * int64(obcfg.GetConfig().Resource.DefaultDiskUsePercent) / int64(oceanbaseconst.GigaConverter) / 100
			if logDiskSizeG > 0 {
				logDiskSizeParam = fmt.Sprintf("log_disk_size=%dG", logDiskSizeG)
			}
		}
	}

	startupParameters := []string{
		fmt.Sprintf("cluster_id=%d", m.Resource.Spec.ClusterId),
		"local_ip=${POD_IP}",
		fmt.Sprintf("port=%d", rpcPort),
		fmt.Sprintf("http_ip_addr=${POD_IP}:%d", httpPort),
		fmt.Sprintf("local_storage_dir=%s", storeMountPath),
	}
	if logDiskSizeParam != "" {
		startupParameters = append(startupParameters, logDiskSizeParam)
	}
	for _, p := range m.Resource.Spec.Parameters {
		reserved := false
		for _, rp := range oceanbaseconst.LogServiceReservedParameters {
			if p.Name == rp {
				reserved = true
				break
			}
		}
		if !reserved {
			startupParameters = append(startupParameters, fmt.Sprintf("%s='%s'", p.Name, p.Value))
		}
	}

	startCmd := fmt.Sprintf(
		`mkdir -p %s %s && while [ -z "${POD_IP}" ]; do sleep 1; done && /home/admin/oblogservice/bin/oblogservice -g "%s" & sleep infinity`,
		storeMountPath, logMountPath,
		strings.Join(startupParameters, ","),
	)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            podName,
			Namespace:       m.Resource.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerRef},
			Labels:          podLabels,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: m.Resource.Spec.ServiceAccount,
			Containers: []corev1.Container{{
				Name:            "oblogservice",
				Image:           m.Resource.Spec.Image,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"bash", "-c", startCmd},
				Env: []corev1.EnvVar{{
					Name: "POD_IP",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "status.podIP",
						},
					},
				}},
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
	return m.Client.Status().Update(m.Ctx, m.Resource)
}

func WaitPodReady(m *OBLogServiceNodeManager) tasktypes.TaskError {
	m.Logger.Info("Waiting for log service node pod to be ready")
	podName := m.Resource.Status.PodName
	if podName == "" {
		podName = m.Resource.Name
	}

	for i := 0; i < 600; i++ {
		pod := &corev1.Pod{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.Resource.Namespace,
			Name:      podName,
		}, pod)
		if err != nil {
			return errors.Wrap(err, "get pod")
		}
		if pod.Status.Phase == corev1.PodRunning && pod.Status.PodIP != "" {
			m.Resource.Status.PodIP = pod.Status.PodIP
			m.Resource.Status.Ready = true
			m.Resource.Status.PodPhase = pod.Status.Phase
			// Update ServiceIP if not set yet
			if m.Resource.Status.ServiceIP == "" {
				svcName := fmt.Sprintf("%s-svc", podName)
				svc := &corev1.Service{}
				if svcErr := m.Client.Get(m.Ctx, types.NamespacedName{
					Namespace: m.Resource.Namespace,
					Name:      svcName,
				}, svc); svcErr == nil {
					m.Resource.Status.ServiceIP = svc.Spec.ClusterIP
				}
			}
			m.Logger.Info("Log service node pod is ready", "pod", podName, "ip", pod.Status.PodIP)
			return m.Client.Status().Update(m.Ctx, m.Resource)
		}
		time.Sleep(time.Second)
	}
	return errors.New("timeout waiting for log service node pod to be ready")
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
	m.Resource.Status.Ready = false
	m.Resource.Status.PodIP = ""
	return nil
}
