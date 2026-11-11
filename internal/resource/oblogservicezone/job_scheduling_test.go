package oblogservicezone

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

type captureZoneJobClient struct {
	client.Client
	jobs []*batchv1.Job
}

func (c *captureZoneJobClient) Create(_ context.Context, obj client.Object, _ ...client.CreateOption) error {
	c.jobs = append(c.jobs, obj.(*batchv1.Job).DeepCopy())
	return errors.New("stop before polling")
}

func TestLogServiceMembershipJobsInheritZonePlacement(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(*OBLogServiceZoneManager) error
	}{{"register", func(m *OBLogServiceZoneManager) error { return RegisterNodeToCluster(m) }}, {"unregister", func(m *OBLogServiceZoneManager) error { return UnregisterNodeFromCluster(m) }}, {"unregister-all", func(m *OBLogServiceZoneManager) error { return UnregisterAllNodesFromCluster(m) }}} {
		t.Run(tc.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = v1alpha1.AddToScheme(scheme)
			zone := &v1alpha1.OBLogServiceZone{ObjectMeta: metav1.ObjectMeta{Name: "ls-z1", Namespace: "qa"}, Spec: v1alpha1.OBLogServiceZoneSpec{ClusterName: "ls", Image: "ls:test", ServiceAccount: "ls-workload", Topology: apitypes.LogServiceZoneTopology{Zone: "z1", Replica: 1, NodeSelector: map[string]string{"pool": "authorized"}, Affinity: &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{}}, Tolerations: []corev1.Toleration{{Key: "pool", Value: "authorized", Operator: corev1.TolerationOpEqual, Effect: corev1.TaintEffectNoSchedule}}}}}
			a := &v1alpha1.OBLogServiceNode{ObjectMeta: metav1.ObjectMeta{Name: "a", Namespace: "qa", Labels: map[string]string{oc.LabelRefOBLogServiceCluster: "ls", oc.LabelRefOBLogServiceZone: zone.Name}, Annotations: map[string]string{oc.AnnotationsLogServiceNodeRegistered: "true"}}, Spec: v1alpha1.OBLogServiceNodeSpec{Zone: "z1"}, Status: v1alpha1.OBLogServiceNodeStatus{Status: "running", PodIP: "10.42.1.1"}}
			b := a.DeepCopy()
			b.Name = "b"
			b.Status.PodIP = "10.42.1.2"
			if tc.name == "register" {
				b.Annotations = nil
			}
			c := &captureZoneJobClient{Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(a, b).Build()}
			logger := logr.Discard()
			_ = tc.run(&OBLogServiceZoneManager{Ctx: context.Background(), Resource: zone, Client: c, Logger: &logger})
			if len(c.jobs) == 0 {
				t.Fatal("no membership Job captured")
			}
			for _, job := range c.jobs {
				pod := job.Spec.Template.Spec
				topo := zone.Spec.Topology
				if !reflect.DeepEqual(pod.NodeSelector, topo.NodeSelector) || !reflect.DeepEqual(pod.Affinity, topo.Affinity) || !reflect.DeepEqual(pod.Tolerations, topo.Tolerations) || pod.ServiceAccountName != "ls-workload" || pod.NodeName != "" {
					t.Fatalf("membership Job lost placement/SA: %+v", pod)
				}
			}
		})
	}
}
