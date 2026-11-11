package oblogservicecluster

import (
	"context"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNativeStorageDependencyBeforeZones(t *testing.T) {
	for _, success := range []bool{false, true} {
		t.Run(map[bool]string{false: "failed-init-creates-no-zones", true: "complete-init-resolves-runtime-only"}[success], func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = v1alpha1.AddToScheme(scheme)
			_ = corev1.AddToScheme(scheme)
			_ = batchv1.AddToScheme(scheme)
			condition := batchv1.JobFailed
			if success {
				condition = batchv1.JobComplete
			}
			job := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "storage-init", Namespace: "demo"}, Status: batchv1.JobStatus{Conditions: []batchv1.JobCondition{{Type: condition, Status: corev1.ConditionTrue, Reason: "Fixture"}}}}
			svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "minio", Namespace: "demo"}, Spec: corev1.ServiceSpec{ClusterIP: "10.43.1.5"}}
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(job, svc).Build()
			ls := &v1alpha1.OBLogServiceCluster{ObjectMeta: metav1.ObjectMeta{Name: "ls", Namespace: "demo", Annotations: map[string]string{resourceutils.StorageInitJobAnnotation: job.Name, resourceutils.ResolveStorageServiceAnnotation: "true"}}, Spec: v1alpha1.OBLogServiceClusterSpec{LogService: &apitypes.LogServiceTemplate{Image: "ls:test"}, ObjectStoreURL: apitypes.ObjectStoreConfig{BucketURL: "s3://logs?host=http://minio.demo.svc:9000"}}}
			// Use the actual topology field type from the API without duplicating it.
			ls.Spec.Topology = append(ls.Spec.Topology, apitypes.LogServiceZoneTopology{Zone: "z1", Replica: 1})
			logger := logr.Discard()
			err := CreateZones(&OBLogServiceClusterManager{Ctx: context.Background(), Client: c, Resource: ls, Logger: &logger})
			zones := &v1alpha1.OBLogServiceZoneList{}
			if listErr := c.List(context.Background(), zones, client.InNamespace("demo")); listErr != nil {
				t.Fatal(listErr)
			}
			if !success {
				if err == nil || len(zones.Items) != 0 {
					t.Fatalf("failed initialization bootstrapped: %v", err)
				}
				return
			}
			if err != nil || len(zones.Items) != 1 {
				t.Fatalf("bootstrap failed: %v, zones=%d", err, len(zones.Items))
			}
			if zones.Items[0].Spec.ObjectStoreURL.BucketURL != "s3://logs?host=http://10.43.1.5:9000" {
				t.Fatal("zone endpoint not resolved")
			}
			if !strings.Contains(ls.Spec.ObjectStoreURL.BucketURL, "minio.demo.svc") {
				t.Fatal("parent spec mutated")
			}
		})
	}
}
