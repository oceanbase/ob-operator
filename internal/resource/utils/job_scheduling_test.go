package utils

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestBootstrapJobSchedulingCopiesOnlyOneTargetZone(t *testing.T) {
	sa := "override-workload"
	cluster := &v1alpha1.OBCluster{Spec: v1alpha1.OBClusterSpec{
		ServiceAccount:   "cluster-workload",
		OBServerTemplate: &apitypes.OBServerTemplate{PodFields: &apitypes.PodFieldsSpec{ServiceAccountName: &sa}},
		Topology: []apitypes.OBZoneTopology{
			{NodeSelector: map[string]string{"pool": "first"}, Tolerations: []corev1.Toleration{{Key: "first", Value: "allowed"}}, Affinity: &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "pool", Operator: corev1.NodeSelectorOpIn, Values: []string{"first"}}}}}}}}},
			{NodeSelector: map[string]string{"pool": "other"}, Tolerations: []corev1.Toleration{{Key: "other", Value: "forbidden"}}},
		},
	}}
	pod := corev1.PodSpec{}
	ApplyJobScheduling(&pod, BootstrapJobScheduling(cluster))
	if pod.ServiceAccountName != sa || pod.NodeSelector["pool"] != "first" || len(pod.Tolerations) != 1 || pod.Tolerations[0].Key != "first" {
		t.Fatalf("placement broadened or SA override lost: %+v", pod)
	}
	pod.NodeSelector["pool"] = "changed"
	pod.Tolerations[0].Key = "changed"
	pod.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values[0] = "changed"
	zone := cluster.Spec.Topology[0]
	if zone.NodeSelector["pool"] != "first" || zone.Tolerations[0].Key != "first" || zone.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values[0] != "first" {
		t.Fatal("Job placement aliases source CR")
	}
}

func TestBootstrapJobSchedulingWithoutTopology(t *testing.T) {
	cluster := &v1alpha1.OBCluster{Spec: v1alpha1.OBClusterSpec{ServiceAccount: "workload"}}
	pod := corev1.PodSpec{}
	ApplyJobScheduling(&pod, BootstrapJobScheduling(cluster))
	if pod.ServiceAccountName != "workload" || pod.NodeSelector != nil || pod.Affinity != nil || pod.Tolerations != nil || pod.NodeName != "" {
		t.Fatalf("unexpected implicit placement: %+v", pod)
	}
}

type captureLegacyJobClient struct {
	client.Client
	job *batchv1.Job
}

func (c *captureLegacyJobClient) Create(_ context.Context, obj client.Object, _ ...client.CreateOption) error {
	c.job = obj.(*batchv1.Job).DeepCopy()
	return errors.New("stop before polling")
}

func TestLegacyRunJobPlacementRemainsUnchanged(t *testing.T) {
	logger := logr.Discard()
	c := &captureLegacyJobClient{}
	_, _, _ = RunJob(context.Background(), c, &logger, "qa", "legacy", "image:test", nil, "true")
	if c.job == nil {
		t.Fatal("missing captured legacy Job")
	}
	pod := c.job.Spec.Template.Spec
	if pod.ServiceAccountName != "" || pod.NodeSelector != nil || pod.Affinity != nil || pod.Tolerations != nil || pod.SchedulerName != corev1.DefaultSchedulerName {
		t.Fatalf("legacy upgrade/other Job behavior changed: %+v", pod)
	}
}
