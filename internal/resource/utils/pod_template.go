package utils

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apipod "k8s.io/kubernetes/pkg/api/v1/pod"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/clientcache"
	obagentconst "github.com/oceanbase/ob-operator/internal/const/obagent"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
)

func PodHasBackupVolume(pod *corev1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == oceanbaseconst.ContainerName {
			for _, mount := range container.VolumeMounts {
				if mount.MountPath == oceanbaseconst.BackupPath {
					return true
				}
			}
		}
	}
	return false
}

func MonitorMatchesPod(server *v1alpha1.OBServer, pod *corev1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == obagentconst.ContainerName {
			revision := server.Annotations[oceanbaseconst.AnnotationsMonitorRevision]
			return server.Spec.MonitorTemplate != nil && (revision == "" || revision == pod.Annotations[oceanbaseconst.AnnotationsMonitorRevision])
		}
	}
	return server.Spec.MonitorTemplate == nil
}

// ZonePodTemplatesReady observes durable child state. Observers can finish their
// rollouts at different times and may place their Pods in a remote cluster.
func ZonePodTemplatesReady(ctx context.Context, local client.Client, zone *v1alpha1.OBZone) (bool, error) {
	servers := &v1alpha1.OBServerList{}
	if err := local.List(ctx, servers, client.InNamespace(zone.Namespace), client.MatchingLabels{oceanbaseconst.LabelRefOBZone: zone.Name}); err != nil {
		return false, err
	}
	if len(servers.Items) != zone.Spec.Topology.Replica {
		return false, nil
	}
	desired := zone.Spec.Topology.EffectiveMonitor(zone.Spec.MonitorTemplate)
	backup := zone.Spec.BackupVolume != nil
	for i := range servers.Items {
		s := &servers.Items[i]
		if !IsLiveChild(s, zone, "OBZone") || s.UID == "" ||
			s.Status.Status != serverstatus.Running || !s.Status.Ready ||
			!s.Spec.MonitorTemplate.Equal(desired) || (s.Spec.BackupVolume != nil) != backup {
			return false, nil
		}
		podClient := local
		if s.Spec.K8sCluster != "" {
			remote, err := clientcache.GetCtrlRuntimeClientFromK8sName(ctx, s.Spec.K8sCluster)
			if err != nil {
				return false, err
			}
			podClient = remote
		}
		pod := &corev1.Pod{}
		if err := podClient.Get(ctx, client.ObjectKeyFromObject(s), pod); err != nil {
			return false, client.IgnoreNotFound(err)
		}
		if !IsLiveChild(pod, s, "OBServer") || pod.Status.Phase != corev1.PodRunning ||
			!apipod.IsPodReady(pod) || !MonitorMatchesPod(s, pod) || PodHasBackupVolume(pod) != backup {
			return false, nil
		}
	}
	return true, nil
}
