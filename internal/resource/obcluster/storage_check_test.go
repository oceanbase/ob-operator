package obcluster

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/oceanbase/ob-operator/internal/telemetry"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	oc "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

func storageFixture(t *testing.T, objs ...client.Object) *OBClusterManager {
	t.Helper()
	m, _ := schedulingFixture()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := batchv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	m.Client = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).WithInterceptorFuncs(interceptor.Funcs{
		Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
			switch obj.(type) {
			case *batchv1.Job:
				obj.SetUID("job-uid")
			case *corev1.PersistentVolumeClaim:
				obj.SetUID("claim-uid")
			}
			return c.Create(ctx, obj, opts...)
		},
	}).Build()
	m.APIReader = m.Client
	return m
}

func TestFilesystemProbeReplay(t *testing.T) {
	m := storageFixture(t)
	ctx := context.Background()
	claim, job, err := m.storageCheckObjects()
	if err != nil {
		t.Fatal(err)
	}
	if !m.CanReplayTask(tCheckEnvironment) || m.CanReplayTask(tBootstrap) {
		t.Fatal("unexpected replay policy")
	}
	for i := 0; i < 2; i++ {
		done, err := m.storageCheckStep(ctx, m.Client, claim, job)
		if done || err != nil {
			t.Fatalf("step %d: %v, %v", i, done, err)
		}
	}
	var jobs batchv1.JobList
	if err = m.Client.List(ctx, &jobs); err != nil || len(jobs.Items) != 1 {
		t.Fatalf("duplicate or missing job: %v %v", jobs.Items, err)
	}
	live := &jobs.Items[0]
	if live.UID != "job-uid" || live.Annotations[storageClaimUID] != "claim-uid" {
		t.Fatalf("unbound Job: %+v", live.ObjectMeta)
	}
	live.Status.Succeeded = 1
	live.Status.Conditions = []batchv1.JobCondition{{Type: batchv1.JobComplete, Status: corev1.ConditionTrue}}
	if err = m.Client.Status().Update(ctx, live); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		done, err := m.storageCheckStep(ctx, m.Client, claim, job)
		if !done || err != nil {
			t.Fatalf("completed replay %d: %v, %v", i, done, err)
		}
	}
	if err = m.Client.Get(ctx, client.ObjectKeyFromObject(claim), &corev1.PersistentVolumeClaim{}); !apierrors.IsNotFound(err) {
		t.Fatalf("scratch PVC survived completion: %v", err)
	}
}

func TestFilesystemProbeRejectsForeignObjects(t *testing.T) {
	base := storageFixture(t)
	want, job, err := base.storageCheckObjects()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*corev1.PersistentVolumeClaim, *batchv1.Job)
	}{
		{"different claim UID", func(p *corev1.PersistentVolumeClaim, j *batchv1.Job) { p.UID = "another-claim" }},
		{"foreign claim", func(p *corev1.PersistentVolumeClaim, j *batchv1.Job) { p.OwnerReferences[0].UID = "foreign" }},
		{"different request", func(p *corev1.PersistentVolumeClaim, j *batchv1.Job) {
			p.Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse("8Gi")
		}},
		{"foreign job", func(p *corev1.PersistentVolumeClaim, j *batchv1.Job) { j.OwnerReferences = nil }},
		{"different mount", func(p *corev1.PersistentVolumeClaim, j *batchv1.Job) {
			j.Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName = "unrelated"
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, j := want.DeepCopy(), job.DeepCopy()
			p.UID = "claim"
			j.UID = "job"
			j.Annotations = map[string]string{storageClaimUID: "claim"}
			tc.change(p, j)
			m := storageFixture(t, p, j)
			if done, err := m.storageCheckStep(context.Background(), m.Client, want, job); done || err == nil {
				t.Fatalf("accepted mismatch: %v %v", done, err)
			}
			if err := m.cleanupTimedOutStorageCheck(context.Background(), m.Client, want, job); err == nil {
				t.Fatal("cleaned unrelated object")
			}
			if err := m.Client.Get(context.Background(), client.ObjectKeyFromObject(p), &corev1.PersistentVolumeClaim{}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLegacyStorageClaimsAndSharedStorage(t *testing.T) {
	base := storageFixture(t)
	base.OBCluster.Spec.DeploymentMode = oc.DeploymentModeSharedStorage
	base.OBCluster.Spec.OBServerTemplate.Storage.RedoLogStorage = nil
	claim, job, err := base.storageCheckObjects()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(job.Spec.Template.Spec.Containers[0].Command[2], "env-check storage ") {
		t.Fatal("wrong helper command")
	}
	old := claim.DeepCopy()
	old.Name = "qacheck-clog-claim-ptfhnr"
	old.UID = "legacy-uid"
	inUse := old.DeepCopy()
	inUse.Name = "qacheck-clog-claim-fhklmn"
	inUse.UID = "used"
	foreign := old.DeepCopy()
	foreign.Name = "qacheck-clog-claim-hjklmn"
	foreign.OwnerReferences[0].UID = "different-cluster"
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "user", Namespace: old.Namespace}, Spec: corev1.PodSpec{Volumes: []corev1.Volume{{Name: "scratch", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: inUse.Name}}}}}}
	retryClaim := old.DeepCopy()
	retryClaim.Name = "qacheck-clog-claim-bchjkl"
	retryClaim.UID = "retry-claim"
	// A failed Pod may be between attempts in a still-active Job.
	retryJob := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "legacy-retry", Namespace: old.Namespace}, Status: batchv1.JobStatus{Failed: 1}}
	retryJob.Spec.Template.Spec.Volumes = []corev1.Volume{{Name: "scratch", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: retryClaim.Name}}}}
	m := storageFixture(t, old, inUse, foreign, pod, retryClaim, retryJob)
	if err = m.cleanupLegacyStorageClaims(context.Background(), m.Client, claim); err != nil {
		t.Fatal(err)
	}
	for _, p := range []*corev1.PersistentVolumeClaim{old, inUse, foreign, retryClaim} {
		err = m.Client.Get(context.Background(), client.ObjectKeyFromObject(p), &corev1.PersistentVolumeClaim{})
		if p == old && !apierrors.IsNotFound(err) || p != old && err != nil {
			t.Fatalf("claim %s: %v", p.Name, err)
		}
	}
}

func TestStorageProbeCompatibilityExitCodes(t *testing.T) {
	m := storageFixture(t)
	_, job, err := m.storageCheckObjects()
	if err != nil {
		t.Fatal(err)
	}
	wrapped := job.Spec.Template.Spec.Containers[0].Command[2]
	for _, tc := range []struct{ helper, result int }{{0, 0}, {1, 0}, {10, 10}, {127, 127}} {
		t.Run(fmt.Sprint(tc.helper), func(t *testing.T) {
			command := strings.Replace(wrapped, "/home/admin/oceanbase/bin/oceanbase-helper", fmt.Sprintf("fake_helper() { return %d; }; fake_helper", tc.helper), 1)
			err := exec.Command("bash", "-c", command).Run()
			code := 0
			var exit *exec.ExitError
			if errors.As(err, &exit) {
				code = exit.ExitCode()
			} else if err != nil {
				t.Fatal(err)
			}
			if code != tc.result {
				t.Fatalf("helper exit %d became %d", tc.helper, code)
			}
		})
	}
}

func TestStorageProbeConcurrentCreateAndTimeoutCleanup(t *testing.T) {
	m := storageFixture(t)
	ctx := context.Background()
	pvc, job, err := m.storageCheckObjects()
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := m.storageCheckStep(ctx, m.Client, pvc, job); err != nil {
				t.Errorf("concurrent creation: %v", err)
			}
		}()
	}
	wg.Wait()
	var claims corev1.PersistentVolumeClaimList
	var jobs batchv1.JobList
	if err = m.Client.List(ctx, &claims); err != nil || len(claims.Items) != 1 {
		t.Fatalf("claims: %v count %d", err, len(claims.Items))
	}
	if err = m.Client.List(ctx, &jobs); err != nil || len(jobs.Items) != 1 {
		t.Fatalf("jobs: %v count %d", err, len(jobs.Items))
	}
	if err = m.cleanupTimedOutStorageCheck(ctx, m.Client, pvc, job); err != nil {
		t.Fatal(err)
	}
	if err = m.Client.Get(ctx, client.ObjectKeyFromObject(job), &batchv1.Job{}); !apierrors.IsNotFound(err) {
		t.Fatalf("timed out Job survived: %v", err)
	}
	if err = m.Client.Get(ctx, client.ObjectKeyFromObject(pvc), &corev1.PersistentVolumeClaim{}); !apierrors.IsNotFound(err) {
		t.Fatalf("timed out PVC survived: %v", err)
	}
}

func TestFailedFilesystemProbeReturnsItsErrorAndAllowsSameGenerationRetry(t *testing.T) {
	m := storageFixture(t)
	ctx := context.Background()
	claim, job, err := m.storageCheckObjects()
	if err != nil {
		t.Fatal(err)
	}
	previous := job.DeepCopy()
	previous.UID = "previous-failed-job"
	previous.Annotations = map[string]string{storageClaimUID: "previous-claim"}
	previous.Status.Conditions = []batchv1.JobCondition{{Type: batchv1.JobFailed, Status: corev1.ConditionTrue, Message: "filesystem rejects fallocate"}}
	failedClaim := claim.DeepCopy()
	failedClaim.UID = "previous-claim"
	m = storageFixture(t, previous, failedClaim)
	rec := record.NewFakeRecorder(2)
	m.Recorder = telemetry.NewRecorder(ctx, rec)
	defer m.Recorder.Done()
	done, err := m.storageCheckStep(ctx, m.Client, claim, job)
	if done || err == nil || !strings.Contains(err.Error(), string(previous.UID)) || !strings.Contains(err.Error(), "filesystem rejects fallocate") {
		t.Fatalf("lost original failure: %v %v", done, err)
	}
	select {
	case event := <-rec.Events:
		if !strings.Contains(event, "FilesystemProbeFailed") || !strings.Contains(event, string(previous.UID)) {
			t.Fatalf("incorrect event: %s", event)
		}
	default:
		t.Fatal("missing Warning for failed attempt")
	}
	if done, err = m.storageCheckStep(ctx, m.Client, claim, job); done || err != nil {
		t.Fatalf("retry: %v %v", done, err)
	}
	replacement := &batchv1.Job{}
	if err = m.Client.Get(ctx, client.ObjectKeyFromObject(job), replacement); err != nil {
		t.Fatal(err)
	}
	if replacement.UID == previous.UID || replacement.Annotations[storageClaimUID] == string(failedClaim.UID) {
		t.Fatalf("reused failed attempt: %+v", replacement.ObjectMeta)
	}
}
