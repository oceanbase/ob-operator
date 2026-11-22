/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package obcluster

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

// Creation is intercepted before any Job polling/log access. No cluster config
// is loaded, and the client never performs network operations.
type capturePreflightClient struct {
	client.Client
	jobs []*batchv1.Job
}

func (c *capturePreflightClient) Get(_ context.Context, key client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
	return apierrors.NewNotFound(schema.GroupResource{Group: "batch", Resource: "jobs"}, key.Name)
}

func (c *capturePreflightClient) List(context.Context, client.ObjectList, ...client.ListOption) error {
	return nil
}

func (c *capturePreflightClient) Create(_ context.Context, obj client.Object, _ ...client.CreateOption) error {
	if pvc, ok := obj.(*corev1.PersistentVolumeClaim); ok {
		pvc.UID = "captured-claim"
	}
	if job, ok := obj.(*batchv1.Job); ok {
		c.jobs = append(c.jobs, job.DeepCopy())
		return errors.New("stop after capturing Job")
	}
	return nil
}

func (c *capturePreflightClient) Delete(context.Context, client.Object, ...client.DeleteOption) error {
	return nil
}

func schedulingFixture() (*OBClusterManager, *capturePreflightClient) {
	logger := logr.Discard()
	c := &capturePreflightClient{}
	storage := &apitypes.StorageSpec{StorageClass: "local-path", Size: resource.MustParse("1Gi")}
	cluster := &v1alpha1.OBCluster{
		TypeMeta:   metav1.TypeMeta{Kind: "OBCluster", APIVersion: "oceanbase.oceanbase.com/v1alpha1"},
		ObjectMeta: metav1.ObjectMeta{Name: "qa", Namespace: "qa-ns", UID: "test-cluster", Annotations: map[string]string{oceanbaseconst.AnnotationsMode: oceanbaseconst.ModeService}},
		Spec: v1alpha1.OBClusterSpec{
			ServiceAccount:   "qa-workload",
			OBServerTemplate: &apitypes.OBServerTemplate{Image: "registry.example/observer:test", Storage: &apitypes.OceanbaseStorageSpec{DataStorage: storage, RedoLogStorage: storage}},
			Topology: []apitypes.OBZoneTopology{{Zone: "z1", Replica: 1,
				NodeSelector: map[string]string{"capacity.oceanbase.dev/project-uid": "seekdb"},
				Affinity:     &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "kubernetes.io/hostname", Operator: corev1.NodeSelectorOpIn, Values: []string{"worker-a", "worker-b"}}}}}}}},
				Tolerations:  []corev1.Toleration{{Key: "capacity.oceanbase.dev/project-uid", Operator: corev1.TolerationOpEqual, Value: "seekdb", Effect: corev1.TaintEffectNoSchedule}},
			}},
		},
	}
	return &OBClusterManager{Ctx: context.Background(), OBCluster: cluster, Client: c, Logger: &logger}, c
}

func TestBootstrapJobsInheritRestrictedPlacement(t *testing.T) {
	for _, task := range []struct {
		name string
		run  func(*OBClusterManager) error
	}{
		{"image", func(m *OBClusterManager) error { return CheckImageReady(m) }},
		{"storage", func(m *OBClusterManager) error { return CheckEnvironment(m) }},
		{"mode", func(m *OBClusterManager) error { return CheckClusterMode(m) }},
		{"optimize", func(m *OBClusterManager) error { return OptimizeClusterByScenario(m) }},
	} {
		t.Run(task.name, func(t *testing.T) {
			m, c := schedulingFixture()
			_ = task.run(m)
			if len(c.jobs) == 0 {
				t.Fatal("no Job captured")
			}
			zone := m.OBCluster.Spec.Topology[0]
			for _, job := range c.jobs {
				pod := job.Spec.Template.Spec
				if !reflect.DeepEqual(pod.NodeSelector, zone.NodeSelector) || !reflect.DeepEqual(pod.Affinity, zone.Affinity) || !reflect.DeepEqual(pod.Tolerations, zone.Tolerations) {
					t.Fatalf("preflight placement differs from target zone: %+v", pod)
				}
				// Kubernetes ServiceAccount admission injects this SA's imagePullSecrets.
				if pod.ServiceAccountName != "qa-workload" {
					t.Fatalf("lost workload credentials/pull-secret identity: %q", pod.ServiceAccountName)
				}
				if pod.NodeName != "" {
					t.Fatal("must not bypass scheduler using nodeName")
				}
			}
		})
	}
}

func TestClusterModeReturnsJobCreationFailure(t *testing.T) {
	m, c := schedulingFixture()
	if err := CheckClusterMode(m); err == nil {
		t.Fatal("expected error creating the version job")
	}
	if len(c.jobs) != 1 {
		t.Fatalf("created %d jobs after version job creation failed, want 1", len(c.jobs))
	}
}
