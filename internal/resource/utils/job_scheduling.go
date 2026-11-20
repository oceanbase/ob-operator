package utils

import (
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// JobScheduling is internal placement for cluster preflight Jobs, not a new CRD
// field. Use one actual target zone rather than combining zones' selectors or
// tolerations, which could broaden access to unrelated node pools.
type JobScheduling struct {
	NodeSelector       map[string]string
	Affinity           *corev1.Affinity
	Tolerations        []corev1.Toleration
	ServiceAccountName string
}

func BootstrapJobScheduling(cluster *v1alpha1.OBCluster) *JobScheduling {
	scheduling := &JobScheduling{ServiceAccountName: cluster.Spec.ServiceAccount}
	if len(cluster.Spec.Topology) > 0 {
		zone := cluster.Spec.Topology[0]
		scheduling.NodeSelector = zone.NodeSelector
		scheduling.Affinity = zone.Affinity
		scheduling.Tolerations = zone.Tolerations
	}
	if template := cluster.Spec.OBServerTemplate; template != nil && template.PodFields != nil {
		if sa := template.PodFields.ServiceAccountName; sa != nil && *sa != "" {
			scheduling.ServiceAccountName = *sa
		}
	}
	return scheduling
}

func ApplyJobScheduling(pod *corev1.PodSpec, scheduling *JobScheduling) {
	if scheduling == nil {
		return
	}
	// Deep-copy mutable fields: Jobs must never alias or mutate the source CR.
	copy := (&corev1.PodSpec{NodeSelector: scheduling.NodeSelector, Affinity: scheduling.Affinity, Tolerations: scheduling.Tolerations}).DeepCopy()
	pod.NodeSelector, pod.Affinity, pod.Tolerations = copy.NodeSelector, copy.Affinity, copy.Tolerations
	// Kubernetes ServiceAccount admission supplies imagePullSecrets from this
	// same workload account; do not copy credential contents or use manager SA.
	pod.ServiceAccountName = scheduling.ServiceAccountName
}
