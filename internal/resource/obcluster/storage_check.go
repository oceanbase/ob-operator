package obcluster

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

const storageClaimUID = "oceanbase.oceanbase.com/storage-check-claim-uid"

func (m *OBClusterManager) storageCheckObjects() (*corev1.PersistentVolumeClaim, *batchv1.Job, error) {
	c := m.OBCluster
	if c.UID == "" || c.Spec.OBServerTemplate == nil || c.Spec.OBServerTemplate.Storage == nil {
		return nil, nil, fmt.Errorf("storage check needs an OBCluster UID and storage configuration")
	}
	storage := c.Spec.OBServerTemplate.Storage.RedoLogStorage
	if c.Spec.DeploymentMode == oceanbaseconst.DeploymentModeSharedStorage {
		storage = c.Spec.OBServerTemplate.Storage.DataStorage
	}
	if storage == nil {
		return nil, nil, fmt.Errorf("storage check has no filesystem to check in deployment mode %q", c.Spec.DeploymentMode)
	}
	// Older images exit 1 when the helper has no env-check command. Filesystem errors have other exit codes.
	command := "/home/admin/oceanbase/bin/oceanbase-helper env-check storage " + oceanbaseconst.ClogPath + "; rc=$?; if [ \"$rc\" -eq 1 ]; then exit 0; fi; exit \"$rc\""
	container := corev1.Container{Name: "job-runner", Image: c.Spec.OBServerTemplate.Image, Command: []string{"bash", "-c", command},
		VolumeMounts: []corev1.VolumeMount{{Name: "storage", MountPath: oceanbaseconst.ClogPath}},
	}
	// Preserve the Kubernetes default used by previous filesystem Jobs, including registries with ports.
	image, _, hasDigest := strings.Cut(container.Image, "@")
	_, tag, hasTag := strings.Cut(path.Base(image), ":")
	container.ImagePullPolicy = corev1.PullIfNotPresent
	if tag == "latest" || (!hasDigest && !hasTag) {
		container.ImagePullPolicy = corev1.PullAlways
	}
	job := m.bootstrapProbeJob("storage-check-", 300, container)
	claim := &corev1.PersistentVolumeClaim{
		ObjectMeta: *job.ObjectMeta.DeepCopy(),
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources:        corev1.VolumeResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceStorage: storage.Size}},
			StorageClassName: ptr.To(storage.StorageClass),
		},
	}
	job.Spec.Template.Spec.Volumes = []corev1.Volume{{Name: "storage", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: claim.Name}}}}
	return claim, job, nil
}

func matchesStorageClaim(want, got *corev1.PersistentVolumeClaim) bool {
	return got.Namespace == want.Namespace && equality.Semantic.DeepEqual(got.OwnerReferences, want.OwnerReferences) &&
		equality.Semantic.DeepEqual(got.Spec.AccessModes, want.Spec.AccessModes) &&
		equality.Semantic.DeepEqual(got.Spec.StorageClassName, want.Spec.StorageClassName) &&
		equality.Semantic.DeepEqual(got.Spec.Resources, want.Spec.Resources) &&
		ptr.Deref(got.Spec.VolumeMode, corev1.PersistentVolumeFilesystem) == corev1.PersistentVolumeFilesystem &&
		got.Spec.Selector == nil && got.Spec.DataSource == nil && got.Spec.DataSourceRef == nil && got.Spec.VolumeAttributesClassName == nil
}

func (m *OBClusterManager) checkStorageEnvironment() error {
	claim, job, err := m.storageCheckObjects()
	if err != nil {
		return err
	}
	reader := m.APIReader
	if reader == nil {
		reader = m.Client
	}
	cfg := obcfg.GetConfig().Time
	interval := time.Duration(max(cfg.CheckJobInterval, 1)) * time.Second
	err = wait.PollUntilContextTimeout(m.Ctx, interval, time.Duration(max(1, cfg.CheckJobMaxRetries))*interval, true, func(ctx context.Context) (bool, error) {
		done, err := m.storageCheckStep(ctx, reader, claim, job)
		if apierrors.IsConflict(err) || apierrors.IsTimeout(err) || apierrors.IsServerTimeout(err) || apierrors.IsTooManyRequests(err) {
			return false, nil
		}
		if err != nil || !done {
			return false, err
		}
		return true, m.cleanupLegacyStorageClaims(ctx, reader, claim)
	})
	if errors.Is(err, context.DeadlineExceeded) && m.Ctx.Err() == nil {
		return errors.Join(err, m.cleanupTimedOutStorageCheck(m.Ctx, reader, claim, job))
	}
	return err
}

func (m *OBClusterManager) storageCheckStep(ctx context.Context, reader client.Reader, wantClaim *corev1.PersistentVolumeClaim, wantJob *batchv1.Job) (bool, error) {
	key := client.ObjectKeyFromObject(wantClaim)
	job := &batchv1.Job{}
	err := reader.Get(ctx, key, job)
	if err != nil && !apierrors.IsNotFound(err) {
		return false, err
	}
	claim := &corev1.PersistentVolumeClaim{}
	claimErr := reader.Get(ctx, key, claim)
	if claimErr != nil && !apierrors.IsNotFound(claimErr) {
		return false, claimErr
	}
	if claimErr == nil && (!matchesStorageClaim(wantClaim, claim) || claim.UID == "") {
		return false, fmt.Errorf("storage check PVC %s has an unexpected identity or specification", key)
	}
	if apierrors.IsNotFound(err) {
		if apierrors.IsNotFound(claimErr) {
			claim = wantClaim.DeepCopy()
			if err = m.Client.Create(ctx, claim); apierrors.IsAlreadyExists(err) {
				return false, nil
			} else if err != nil {
				return false, err
			}
		}
		if claim.DeletionTimestamp != nil {
			return false, nil
		}
		if claim.UID == "" {
			return false, fmt.Errorf("storage check PVC %s returned without a UID", key)
		}
		job = wantJob.DeepCopy()
		job.Annotations = map[string]string{storageClaimUID: string(claim.UID)}
		if err = m.Client.Create(ctx, job); apierrors.IsAlreadyExists(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}
		return false, nil
	}
	// A previous failed attempt may still be terminating. Validate even a deleting
	// Job before waiting for its name to become available for the next attempt.
	match := job
	if job.DeletionTimestamp != nil {
		match = job.DeepCopy()
		match.DeletionTimestamp = nil
	}
	if !matchesBootstrapProbe(wantJob, match) || job.UID == "" || job.Annotations[storageClaimUID] == "" ||
		(claimErr == nil && types.UID(job.Annotations[storageClaimUID]) != claim.UID) {
		return false, fmt.Errorf("storage check Job %s does not match its owner, probe or PVC UID", key)
	}
	if job.DeletionTimestamp != nil {
		return false, nil
	}
	done, resultErr := bootstrapProbeResult(ctx, reader, job)
	if !done && resultErr == nil {
		// Another reconciler can finish the Job and remove its claim between the reads.
		// Re-read on the next poll. Never create a substitute claim for a running Job.
		return false, nil
	}
	if resultErr != nil {
		// OBResourceRescue can reset the same generation. Keep this attempt's UID
		// and cause in its Warning and returned error, then release the failed Job name.
		resultErr = fmt.Errorf("filesystem probe Job %s (UID %s): %w", key, job.UID, resultErr)
		if m.Recorder != nil {
			m.Recorder.Event(m.OBCluster, corev1.EventTypeWarning, "FilesystemProbeFailed", resultErr.Error())
		}
		if err = m.Client.Delete(ctx, job, client.Preconditions{UID: &job.UID}, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
			return false, errors.Join(resultErr, err)
		}
	}
	// A successful result stays on the Job until TTL. If the manager stops between deletion and checkpoint,
	// another manager may consume the same result even after the original claim is gone.
	if claimErr == nil && claim.DeletionTimestamp == nil {
		if err = m.Client.Delete(ctx, claim, client.Preconditions{UID: &claim.UID}); client.IgnoreNotFound(err) != nil {
			return false, err
		}
	}
	return done, resultErr
}

// Old Jobs had random names and no owner reference. Run a fresh owned check. Only retire
// the previous operator's specially named, same-UID scratch claims. Ordinary claims and
// claims with other owners or different templates survive.
func (m *OBClusterManager) cleanupLegacyStorageClaims(ctx context.Context, reader client.Reader, want *corev1.PersistentVolumeClaim) error {
	var list corev1.PersistentVolumeClaimList
	if err := reader.List(ctx, &list, client.InNamespace(want.Namespace)); err != nil {
		return err
	}
	prefix := m.OBCluster.Name + "check-clog-claim-"
	for i := range list.Items {
		pvc := &list.Items[i]
		suffix, legacy := strings.CutPrefix(pvc.Name, prefix)
		if !legacy || len(suffix) != 6 || strings.Trim(suffix, "bcdfghjklmnpqrstvwxz2456789") != "" || pvc.UID == "" || pvc.DeletionTimestamp != nil || !matchesStorageClaim(want, pvc) {
			continue
		}
		// The old claim has no link to the old, ownerless Job. Leave active volumes
		// protected by their OBCluster owner if another workload still references them.
		inUse, err := storageClaimInUse(ctx, reader, pvc)
		if err != nil {
			return err
		}
		if inUse {
			continue
		}
		if err := m.Client.Delete(ctx, pvc, client.Preconditions{UID: &pvc.UID, ResourceVersion: &pvc.ResourceVersion}); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}

func (m *OBClusterManager) cleanupTimedOutStorageCheck(ctx context.Context, reader client.Reader, wantClaim *corev1.PersistentVolumeClaim, wantJob *batchv1.Job) error {
	key := client.ObjectKeyFromObject(wantClaim)
	claim := &corev1.PersistentVolumeClaim{}
	if err := reader.Get(ctx, key, claim); err != nil {
		return client.IgnoreNotFound(err)
	}
	if !matchesStorageClaim(wantClaim, claim) || claim.UID == "" {
		return fmt.Errorf("cannot clean up unrelated PVC %s", key)
	}
	job := &batchv1.Job{}
	if err := reader.Get(ctx, key, job); err == nil {
		if !matchesBootstrapProbe(wantJob, job) || job.UID == "" || job.Annotations[storageClaimUID] != string(claim.UID) {
			return fmt.Errorf("cannot clean up unrelated Job %s", key)
		}
		if err = m.Client.Delete(ctx, job, client.Preconditions{UID: &job.UID}, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
			return err
		}
	} else if !apierrors.IsNotFound(err) {
		return err
	}
	return client.IgnoreNotFound(m.Client.Delete(ctx, claim, client.Preconditions{UID: &claim.UID}))
}

func storageClaimInUse(ctx context.Context, reader client.Reader, pvc *corev1.PersistentVolumeClaim) (bool, error) {
	usesClaim := func(pod corev1.PodSpec) bool {
		for _, volume := range pod.Volumes {
			if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == pvc.Name {
				return true
			}
		}
		return false
	}
	var pods corev1.PodList
	if err := reader.List(ctx, &pods, client.InNamespace(pvc.Namespace)); err != nil {
		return false, err
	}
	for i := range pods.Items {
		pod := &pods.Items[i]
		if pod.Status.Phase != corev1.PodSucceeded && pod.Status.Phase != corev1.PodFailed && usesClaim(pod.Spec) {
			return true, nil
		}
	}
	var jobs batchv1.JobList
	if err := reader.List(ctx, &jobs, client.InNamespace(pvc.Namespace)); err != nil {
		return false, err
	}
	for i := range jobs.Items {
		job := &jobs.Items[i]
		if !usesClaim(job.Spec.Template.Spec) {
			continue
		}
		terminal := false
		for _, condition := range job.Status.Conditions {
			if condition.Status == corev1.ConditionTrue && (condition.Type == batchv1.JobComplete || condition.Type == batchv1.JobFailed) {
				terminal = true
				break
			}
		}
		if !terminal {
			return true, nil
		}
	}
	return false, nil
}
