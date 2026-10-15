package oblogservicecluster

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oc "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

type captureBootstrapJobClient struct {
	client.Client
	job *batchv1.Job
}

func (c *captureBootstrapJobClient) Create(_ context.Context, obj client.Object, _ ...client.CreateOption) error {
	c.job = obj.(*batchv1.Job).DeepCopy()
	return errors.New("stop before polling")
}

func TestLogServiceBootstrapJobInheritsActualNodePlacement(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
	zone := &v1alpha1.OBLogServiceZone{ObjectMeta: metav1.ObjectMeta{Name: "ls-z1", Namespace: "qa", Labels: map[string]string{oc.LabelRefOBLogServiceCluster: "ls"}}}
	node := &v1alpha1.OBLogServiceNode{ObjectMeta: metav1.ObjectMeta{Name: "node", Namespace: "qa", Labels: map[string]string{oc.LabelRefOBLogServiceZone: zone.Name}}, Spec: v1alpha1.OBLogServiceNodeSpec{Zone: "z1", NodeSelector: map[string]string{"pool": "authorized"}, Affinity: &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "kubernetes.io/hostname", Operator: corev1.NodeSelectorOpIn, Values: []string{"worker-a"}}}}}}}}, Tolerations: []corev1.Toleration{{Key: "pool", Value: "authorized", Operator: corev1.TolerationOpEqual, Effect: corev1.TaintEffectNoSchedule}}}, Status: v1alpha1.OBLogServiceNodeStatus{PodIP: "10.42.1.1"}}
	c := &captureBootstrapJobClient{Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(zone, node).Build()}
	logger := logr.Discard()
	ls := &v1alpha1.OBLogServiceCluster{ObjectMeta: metav1.ObjectMeta{Name: "ls", Namespace: "qa"}, Spec: v1alpha1.OBLogServiceClusterSpec{ServiceAccount: "ls-workload", LogService: &apitypes.LogServiceTemplate{Image: "ls:test"}, ObjectStoreURL: apitypes.ObjectStoreConfig{BucketURL: "s3://bucket?host=http://10.43.1.1:9000", SecretRef: corev1.LocalObjectReference{Name: "s3-secret"}}}}
	_ = BootstrapLogService(&OBLogServiceClusterManager{Ctx: context.Background(), Resource: ls, Client: c, Logger: &logger})
	if c.job == nil {
		t.Fatal("no bootstrap Job captured")
	}
	pod := c.job.Spec.Template.Spec
	if !reflect.DeepEqual(pod.NodeSelector, node.Spec.NodeSelector) || !reflect.DeepEqual(pod.Affinity, node.Spec.Affinity) || !reflect.DeepEqual(pod.Tolerations, node.Spec.Tolerations) || pod.ServiceAccountName != "ls-workload" || pod.NodeName != "" {
		t.Fatalf("LS bootstrap lost restricted placement/credentials: %+v", pod)
	}
	if len(pod.Volumes) != 1 || pod.Volumes[0].Secret.SecretName != "s3-secret" {
		t.Fatal("credential mount changed")
	}
}
