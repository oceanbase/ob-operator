package utils

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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func storageTestClient(objects ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = batchv1.AddToScheme(scheme)
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}
func TestResolveStorageServiceURL(t *testing.T) {
	obj := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: "demo", Annotations: map[string]string{ResolveStorageServiceAnnotation: "true"}}}
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "store", Namespace: "demo"}, Spec: corev1.ServiceSpec{ClusterIP: "10.43.2.3"}}
	for _, tc := range []struct {
		name, input, want string
		bad               bool
	}{
		{"short", "s3://data?host=http://store.demo.svc:9000&s3_region=us-east-1", "s3://data?host=http://10.43.2.3:9000&s3_region=us-east-1", false},
		{"fqdn", "s3://data?host=http://store.demo.svc.cluster.local:9000&s3_region=us-east-1", "s3://data?host=http://10.43.2.3:9000&s3_region=us-east-1", false},
		{"encoded", "s3://data?s3_region=a&host=http%3A%2F%2Fstore.demo.svc%3A9000&other=a%2Bb", "s3://data?s3_region=a&host=http://10.43.2.3:9000&other=a%2Bb", false},
		{"external", "s3://data?host=http://storage.example.com:9000", "s3://data?host=http://storage.example.com:9000", false},
		{"tls", "s3://data?host=https://store.demo.svc:9000", "s3://data?host=https://store.demo.svc:9000", false},
		{"ip", "s3://data?host=http://10.0.0.1:9000", "s3://data?host=http://10.0.0.1:9000", false},
		{"cross-namespace", "s3://data?host=http://store.other.svc:9000", "", true},
		{"missing", "s3://data?host=http://absent.demo.svc:9000", "", true},
		{"duplicate", "s3://data?host=http://store.demo.svc&host=http://other", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolveStorageServiceURL(context.Background(), storageTestClient(svc), obj, tc.input)
			if (err != nil) != tc.bad || got != tc.want {
				t.Fatalf("got %q, %v", got, err)
			}
		})
	}
	for _, ip := range []string{"None", "", "fd00::1"} {
		svc := svc.DeepCopy()
		svc.Spec.ClusterIP = ip
		if _, err := ResolveStorageServiceURL(context.Background(), storageTestClient(svc), obj, "s3://data?host=http://store.demo.svc:9000"); err == nil {
			t.Fatalf("accepted unusable ClusterIP %s", ip)
		}
	}
	obj.Annotations = nil
	raw := "s3://data?host=http://missing.demo.svc:9000"
	if got, err := ResolveStorageServiceURL(context.Background(), nil, obj, raw); err != nil || got != raw {
		t.Fatal("legacy endpoint changed", got, err)
	}
}

func TestWaitForStorageInitialization(t *testing.T) {
	obj := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: "demo", Annotations: map[string]string{StorageInitJobAnnotation: "init"}}}
	for _, tc := range []struct {
		name      string
		condition batchv1.JobCondition
		want      string
	}{
		{"complete", batchv1.JobCondition{Type: batchv1.JobComplete, Status: corev1.ConditionTrue}, ""},
		{"failed", batchv1.JobCondition{Type: batchv1.JobFailed, Status: corev1.ConditionTrue, Reason: "BackoffLimitExceeded"}, "BackoffLimitExceeded"},
		{"not-complete", batchv1.JobCondition{Type: batchv1.JobComplete, Status: corev1.ConditionFalse}, "deadline exceeded"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			job := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "init", Namespace: "demo"}, Status: batchv1.JobStatus{Conditions: []batchv1.JobCondition{tc.condition}}}
			err := WaitForStorageInitialization(context.Background(), storageTestClient(job), obj, 30*time.Millisecond, time.Millisecond)
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("unexpected %v", err)
			}
		})
	}
	t.Run("missing-does-not-bootstrap", func(t *testing.T) {
		err := WaitForStorageInitialization(context.Background(), storageTestClient(), obj, time.Hour, time.Millisecond)
		if err == nil || !strings.Contains(err.Error(), "retry with the current resource configuration") {
			t.Fatal(err)
		}
	})
	t.Run("replaced-job-releases-stale-task", func(t *testing.T) {
		now := metav1.Now()
		job := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "init", Namespace: "demo", DeletionTimestamp: &now, Finalizers: []string{"test"}}}
		err := WaitForStorageInitialization(context.Background(), storageTestClient(job), obj, time.Hour, time.Millisecond)
		if err == nil || !strings.Contains(err.Error(), "being replaced") {
			t.Fatal(err)
		}
		fresh := obj.DeepCopy()
		fresh.Annotations[StorageInitJobAnnotation] = "init-2"
		job = &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "init-2", Namespace: "demo"}, Status: batchv1.JobStatus{Conditions: []batchv1.JobCondition{{Type: batchv1.JobComplete, Status: corev1.ConditionTrue}}}}
		if err := WaitForStorageInitialization(context.Background(), storageTestClient(job), fresh, time.Second, time.Millisecond); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := WaitForStorageInitialization(ctx, storageTestClient(), obj, time.Hour, time.Millisecond)
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	})
	obj.Annotations = nil
	if err := WaitForStorageInitialization(context.Background(), nil, obj, 0, 0); err != nil {
		t.Fatal("legacy install changed", err)
	}
}
