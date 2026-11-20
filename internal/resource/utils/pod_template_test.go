package utils

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obagentconst "github.com/oceanbase/ob-operator/internal/const/obagent"
	oc "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
)

func TestZonePodTemplateConvergence(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	zone := &v1alpha1.OBZone{ObjectMeta: metav1.ObjectMeta{Name: "z", Namespace: "n", UID: "zone"}}
	zone.Spec.Topology.Replica = 1
	zone.Spec.MonitorTemplate = &apitypes.MonitorTemplate{Image: "v2"}
	baseServer := &v1alpha1.OBServer{ObjectMeta: metav1.ObjectMeta{Name: "s", Namespace: "n", UID: "srv", Labels: map[string]string{oc.LabelRefOBZone: "z"}, OwnerReferences: []metav1.OwnerReference{{Name: "z", Kind: "OBZone", UID: "zone"}}, Annotations: map[string]string{oc.AnnotationsMonitorRevision: "second"}}}
	baseServer.Spec.MonitorTemplate = zone.Spec.MonitorTemplate.DeepCopy()
	baseServer.Status.Status, baseServer.Status.Ready = serverstatus.Running, true
	basePod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "s", Namespace: "n", UID: "pod", OwnerReferences: []metav1.OwnerReference{{Name: "s", Kind: "OBServer", UID: "srv"}}, Annotations: map[string]string{oc.AnnotationsMonitorRevision: "second"}}}
	basePod.Spec.Containers = []corev1.Container{{Name: obagentconst.ContainerName}}
	basePod.Status.Phase = corev1.PodRunning
	basePod.Status.Conditions = []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}
	for _, tc := range []struct {
		name   string
		mutate func(*v1alpha1.OBServer, *corev1.Pod)
		ready  bool
	}{
		{"already completed", func(*v1alpha1.OBServer, *corev1.Pod) {}, true},
		{"old pod", func(_ *v1alpha1.OBServer, p *corev1.Pod) { p.Annotations[oc.AnnotationsMonitorRevision] = "first" }, false},
		{"unrecoverable", func(s *v1alpha1.OBServer, _ *corev1.Pod) { s.Status.Status = serverstatus.Unrecoverable }, false},
		{"waiting for zone patch", func(s *v1alpha1.OBServer, _ *corev1.Pod) { s.Spec.MonitorTemplate.Image = "v1" }, false},
		{"stale owner", func(_ *v1alpha1.OBServer, p *corev1.Pod) { p.OwnerReferences[0].UID = "old" }, false},
		{"terminating pod", func(_ *v1alpha1.OBServer, p *corev1.Pod) {
			now := metav1.Now()
			p.DeletionTimestamp = &now
			p.Finalizers = []string{"test/hold"}
		}, false},
		{"not ready", func(_ *v1alpha1.OBServer, p *corev1.Pod) { p.Status.Conditions = nil }, false},
		{"backup removal pending", func(_ *v1alpha1.OBServer, p *corev1.Pod) {
			p.Spec.Containers = append(p.Spec.Containers, corev1.Container{Name: oc.ContainerName, VolumeMounts: []corev1.VolumeMount{{MountPath: oc.BackupPath}}})
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, p := baseServer.DeepCopy(), basePod.DeepCopy()
			tc.mutate(s, p)
			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(s, p).Build()
			got, err := ZonePodTemplatesReady(context.Background(), cl, zone)
			if err != nil || got != tc.ready {
				t.Fatalf("ready=%v, err=%v", got, err)
			}
		})
	}
}
