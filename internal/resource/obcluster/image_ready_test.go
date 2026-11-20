package obcluster

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestImageCheckRecoversOnlyItsOwnProbe(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	if err := batchv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name     string
		modify   func(*batchv1.Job)
		wantErr  string
		wantJobs int
	}{
		{"completed", func(*batchv1.Job) {}, "", 1},
		{"legacy completed", func(j *batchv1.Job) { j.Name = imageJobPrefix + "12345abc" }, "", 1},
		{"wrong owner UID", func(j *batchv1.Job) { j.OwnerReferences[0].UID = "other" }, "does not match", 1},
		{"no completions", func(j *batchv1.Job) { j.Spec.Completions = ptr.To[int32](0) }, "does not match", 1},
		{"no actual successful Pod", func(j *batchv1.Job) { j.Status.Succeeded = 0 }, "without a successful Pod", 1},
		{"suspended", func(j *batchv1.Job) { j.Spec.Suspend = ptr.To(true) }, "does not match", 1},
		{"no parallelism", func(j *batchv1.Job) { j.Spec.Parallelism = ptr.To[int32](0) }, "does not match", 1},
		{"bound to a different node", func(j *batchv1.Job) { j.Spec.Template.Spec.NodeName = "worker-b" }, "does not match", 1},
		{"overridden helper path", func(j *batchv1.Job) {
			j.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{{Name: "other", MountPath: "/home/admin/oceanbase/bin"}}
		}, "does not match", 1},
		{"wrong image", func(j *batchv1.Job) { j.Spec.Template.Spec.Containers[0].Image = "another" }, "does not match", 1},
		{"wrong command", func(j *batchv1.Job) { j.Spec.Template.Spec.Containers[0].Command = []string{"false"} }, "does not match", 1},
		{"pull policy", func(j *batchv1.Job) { j.Spec.Template.Spec.Containers[0].ImagePullPolicy = corev1.PullNever }, "does not match", 1},
		{"wrong restart policy", func(j *batchv1.Job) { j.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyOnFailure }, "does not match", 1},
		{"wrong placement", func(j *batchv1.Job) { j.Spec.Template.Spec.NodeSelector["other"] = "x" }, "does not match", 1},
		{"running", func(j *batchv1.Job) { j.Status.Conditions = nil; j.Status.Active = 1 }, "context deadline exceeded", 1},
		{"failed", func(j *batchv1.Job) {
			j.Status.Conditions[0].Type = batchv1.JobFailed
			j.Status.Conditions[0].Message = "helper exited 1"
		}, "helper exited 1", 1},
		{"old job for another cluster", func(j *batchv1.Job) { j.Name = imageJobPrefix + "12345abc"; j.OwnerReferences[0].UID = "other" }, "context deadline exceeded", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, _ := schedulingFixture()
			if !m.CanReplayTask(tCheckImageReady) || m.CanReplayTask(tCheckEnvironment) || m.CanReplayTask(tBootstrap) {
				t.Fatal("unexpected bootstrap replay policy")
			}
			existing := m.imageCheckJob()
			existing.UID = "known-job"
			existing.Status.Succeeded = 1
			existing.Status.Conditions = []batchv1.JobCondition{{Type: batchv1.JobComplete, Status: corev1.ConditionTrue}}
			tc.modify(existing)
			m.Client = fake.NewClientBuilder().WithScheme(scheme).WithObjects(existing).Build()
			m.APIReader = m.Client
			var cancel context.CancelFunc
			m.Ctx, cancel = context.WithTimeout(ctx, 250*time.Millisecond)
			defer cancel()
			err := CheckImageReady(m)
			if tc.wantErr == "" && err != nil || tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)) {
				t.Fatalf("got %v, want error containing %q", err, tc.wantErr)
			}
			if tc.wantErr == "context deadline exceeded" && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatal(err)
			}
			var jobs batchv1.JobList
			if err := m.Client.List(ctx, &jobs, client.InNamespace(existing.Namespace)); err != nil || len(jobs.Items) != tc.wantJobs {
				t.Fatalf("jobs=%v err=%v", jobs.Items, err)
			}
			got := &batchv1.Job{}
			if err := m.Client.Get(ctx, client.ObjectKeyFromObject(existing), got); err != nil || got.UID != existing.UID {
				t.Fatalf("lost original Job %s: %v", existing.Name, err)
			}
		})
	}
}

func TestImageCheckRejectsDeletionAndAcceptsAPIDefaults(t *testing.T) {
	m, _ := schedulingFixture()
	want := m.imageCheckJob()
	want.Spec.Template.Spec.ServiceAccountName = ""
	want.Spec.Template.Spec.SchedulerName = ""
	got := want.DeepCopy()
	got.Spec.Template.Spec.ServiceAccountName = "default"
	got.Spec.Template.Spec.SchedulerName = corev1.DefaultSchedulerName
	if !matchesImageCheckJob(want, got) {
		t.Fatal("API defaults do not change the desired probe")
	}
	now := metav1.Now()
	got.DeletionTimestamp = &now
	if matchesImageCheckJob(want, got) {
		t.Fatal("accepted deleting Job")
	}
}

func TestImageCheckLegacyFailedPodBelongsToTheJob(t *testing.T) {
	m, _ := schedulingFixture()
	job := m.imageCheckJob()
	if ptr.Deref(job.Spec.BackoffLimit, int32(-1)) != 0 {
		t.Fatal("helper failures must be reported promptly")
	}
	job.UID = "known-job"
	job.Spec.BackoffLimit = ptr.To[int32](32)
	job.Status.Failed = 1
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "old-failed", Namespace: job.Namespace, Labels: map[string]string{"job-name": job.Name},
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "batch/v1", Kind: "Job", Name: job.Name, UID: "previous-owner", Controller: ptr.To(true)}}},
		Status: corev1.PodStatus{Phase: corev1.PodFailed, Message: "exit status 1"},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod).Build()
	if done, err := imageCheckResult(context.Background(), cl, job); done || err != nil {
		t.Fatalf("used another Job's Pod: %v %v", done, err)
	}
	pod.OwnerReferences[0].UID = job.UID
	cl = fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod).Build()
	if done, err := imageCheckResult(context.Background(), cl, job); done || err == nil || !strings.Contains(err.Error(), "exit status 1") {
		t.Fatalf("failed Job did not report its failed Pod: %v %v", done, err)
	}
}
