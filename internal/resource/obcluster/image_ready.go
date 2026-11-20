package obcluster

import (
	"cmp"
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
)

const imageJobPrefix = "image-pull-ready-"

func (m *OBClusterManager) imageCheckJob() *batchv1.Job {
	cluster := m.OBCluster
	image := cluster.Spec.OBServerTemplate.Image
	identity := sha256.Sum256([]byte(fmt.Sprintf("%s/%d/%s", cluster.UID, cluster.Generation, image)))
	pod := corev1.PodSpec{
		Containers: []corev1.Container{{
			Name: "helper-check-image-pull-ready", Image: image, ImagePullPolicy: corev1.PullIfNotPresent,
			Command: []string{"bash", "-c", "/home/admin/oceanbase/bin/oceanbase-helper help"},
		}},
		RestartPolicy: corev1.RestartPolicyNever,
		SchedulerName: resourceutils.GetSchedulerName(cluster.Spec.OBServerTemplate.PodFields),
	}
	resourceutils.ApplyJobScheduling(&pod, resourceutils.BootstrapJobScheduling(cluster))
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s%x", imageJobPrefix, identity[:12]), Namespace: cluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: v1alpha1.GroupVersion.String(), Kind: "OBCluster", Name: cluster.Name, UID: cluster.UID,
			}},
		},
		Spec: batchv1.JobSpec{
			Template:                corev1.PodTemplateSpec{Spec: pod},
			TTLSecondsAfterFinished: ptr.To[int32](120), BackoffLimit: ptr.To[int32](0),
		},
	}
}

// Compare the command, image identity and scheduling that make this an image check.
// Job metadata, TTL, and API defaults on unrelated Pod fields do not affect the result.
func imageProbeSpec(pod corev1.PodSpec) corev1.PodSpec {
	container := pod.Containers[0]
	return corev1.PodSpec{
		Containers: []corev1.Container{{
			Name: container.Name, Image: container.Image, ImagePullPolicy: container.ImagePullPolicy,
			Command: container.Command, Args: container.Args, WorkingDir: container.WorkingDir,
			Env: container.Env, EnvFrom: container.EnvFrom, VolumeMounts: container.VolumeMounts,
		}},
		RestartPolicy: pod.RestartPolicy, ServiceAccountName: cmp.Or(pod.ServiceAccountName, "default"),
		SchedulerName: cmp.Or(pod.SchedulerName, corev1.DefaultSchedulerName),
		NodeName:      pod.NodeName, NodeSelector: pod.NodeSelector, Affinity: pod.Affinity, Tolerations: pod.Tolerations,
	}
}

func matchesImageCheckJob(want, got *batchv1.Job) bool {
	if got.Namespace != want.Namespace || got.DeletionTimestamp != nil ||
		!equality.Semantic.DeepEqual(want.OwnerReferences, got.OwnerReferences) {
		return false
	}
	if ptr.Deref(got.Spec.Completions, int32(1)) != 1 || ptr.Deref(got.Spec.Parallelism, int32(1)) != 1 || ptr.Deref(got.Spec.Suspend, false) {
		return false
	}
	pod := got.Spec.Template.Spec
	return len(pod.Containers) == 1 && len(pod.InitContainers) == 0 && equality.Semantic.DeepEqual(imageProbeSpec(want.Spec.Template.Spec), imageProbeSpec(pod))
}

func (m *OBClusterManager) findOrCreateImageCheck(ctx context.Context, reader client.Reader, want *batchv1.Job) (*batchv1.Job, error) {
	job := &batchv1.Job{}
	key := client.ObjectKeyFromObject(want)
	err := reader.Get(ctx, key, job)
	if apierrors.IsNotFound(err) {
		// Previous versions used a random eight-character suffix for this read-only Job.
		var list batchv1.JobList
		if err := reader.List(ctx, &list, client.InNamespace(want.Namespace)); err != nil {
			return nil, err
		}
		var previous *batchv1.Job
		for i := range list.Items {
			candidate := &list.Items[i]
			if len(candidate.Name) == len(imageJobPrefix)+8 && strings.HasPrefix(candidate.Name, imageJobPrefix) &&
				matchesImageCheckJob(want, candidate) && (previous == nil || previous.CreationTimestamp.Before(&candidate.CreationTimestamp)) {
				previous = candidate
			}
		}
		if previous != nil {
			return previous, nil
		}
		job = want.DeepCopy()
		err = m.Client.Create(ctx, job)
		if apierrors.IsAlreadyExists(err) {
			err = reader.Get(ctx, key, job)
		}
	}
	if err != nil {
		return nil, err
	}
	if !matchesImageCheckJob(want, job) {
		return nil, fmt.Errorf("image check Job %s does not match the cluster owner, image, command or placement", key)
	}
	return job, nil
}

// Older Jobs permit 32 retries. Surface their failed helper Pod without waiting for all retries.
func imageCheckResult(ctx context.Context, reader client.Reader, job *batchv1.Job) (bool, error) {
	for _, condition := range job.Status.Conditions {
		if condition.Status != corev1.ConditionTrue {
			continue
		}
		switch condition.Type {
		case batchv1.JobFailed, batchv1.JobFailureTarget:
			return false, fmt.Errorf("image check Job %s failed: %s", job.Name, condition.Message)
		case batchv1.JobComplete:
			if job.Status.Succeeded == 0 {
				return false, fmt.Errorf("image check Job %s completed without a successful Pod", job.Name)
			}
			return true, nil
		}
	}
	if job.Status.Failed == 0 || ptr.Deref(job.Spec.BackoffLimit, int32(0)) == 0 {
		return false, nil
	}
	var pods corev1.PodList
	if err := reader.List(ctx, &pods, client.InNamespace(job.Namespace), client.MatchingLabels{"job-name": job.Name}); err != nil {
		return false, err
	}
	for _, pod := range pods.Items {
		owner := metav1.GetControllerOf(&pod)
		if owner != nil && owner.Kind == "Job" && owner.Name == job.Name && owner.UID == job.UID && pod.Status.Phase == corev1.PodFailed {
			return false, fmt.Errorf("image check Job %s failed in Pod %s: %s", job.Name, pod.Name, pod.Status.Message)
		}
	}
	return false, nil
}

func (m *OBClusterManager) checkImageReady() error {
	if m.OBCluster.UID == "" {
		return fmt.Errorf("cannot run image check without a cluster UID")
	}
	reader := m.APIReader
	if reader == nil {
		reader = m.Client
	}
	want := m.imageCheckJob()
	var current *batchv1.Job
	interval := time.Duration(max(1, obcfg.GetConfig().Time.CheckJobInterval)) * time.Second
	return wait.PollUntilContextTimeout(m.Ctx, interval, 8000*interval, true, func(ctx context.Context) (bool, error) {
		var err error
		if current == nil {
			current, err = m.findOrCreateImageCheck(ctx, reader, want)
		} else {
			key, uid := client.ObjectKeyFromObject(current), current.UID
			current = &batchv1.Job{}
			err = reader.Get(ctx, key, current)
			if apierrors.IsNotFound(err) {
				// TTL collection can win while the manager is offline. Repeat only the read-only probe.
				current = nil
				return false, nil
			}
			if err == nil && (current.UID != uid || !matchesImageCheckJob(want, current)) {
				return false, fmt.Errorf("image check Job %s changed identity", key)
			}
		}
		if err != nil {
			return false, err
		}
		return imageCheckResult(ctx, reader, current)
	})
}
