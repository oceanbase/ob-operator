package obzone

import (
	"context"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
)

// Only the parent cluster carries the opt-in annotation. A missing Service
// must stop expansion before opening SQL, rather than passing its DNS name.
func TestAddSharedStorageDestResolvesParentClusterEndpoint(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
	cluster := &v1alpha1.OBCluster{ObjectMeta: metav1.ObjectMeta{Name: "db", Namespace: "test", Annotations: map[string]string{resourceutils.ResolveStorageServiceAnnotation: "true"}}}
	cluster.Spec.DeploymentMode = oceanbaseconst.DeploymentModeSharedStorage
	cluster.Spec.SharedStorageInfo = &apitypes.SharedStorageSpec{BucketURL: "s3://data?host=http://missing.test.svc:9000", SecretRef: corev1.LocalObjectReference{Name: "storage"}}
	cluster.Spec.UserSecrets = &apitypes.OBUserSecrets{Operator: "operator"}
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "storage", Namespace: "test"}, Data: map[string][]byte{"access_id": []byte("user"), "access_key": []byte("secret")}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cluster, secret).Build()
	zone := &v1alpha1.OBZone{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Labels: map[string]string{oceanbaseconst.LabelRefOBCluster: "db"}}}
	logger := logr.Discard()
	err := AddSharedStorageDest(&OBZoneManager{Ctx: context.Background(), Client: c, OBZone: zone, Logger: &logger})
	if err == nil || !strings.Contains(err.Error(), "get storage Service test/missing") {
		t.Fatalf("expected storage resolution failure before SQL, got %v", err)
	}
	got := &v1alpha1.OBCluster{}
	if err := c.Get(context.Background(), client.ObjectKeyFromObject(cluster), got); err != nil {
		t.Fatal(err)
	}
	if got.Spec.SharedStorageInfo.BucketURL != cluster.Spec.SharedStorageInfo.BucketURL {
		t.Fatal("runtime resolution changed the declarative endpoint")
	}
}
