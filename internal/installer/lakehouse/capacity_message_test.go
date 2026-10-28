package lakehouse

import (
	"context"
	"os"
	"strings"
	"testing"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestCapacityFailureUsesCoresAndGiB(t *testing.T) {
	for _, multi := range []bool{false, true} {
		mode := "single"
		if multi {
			mode = "three workers"
		}
		for _, tc := range []struct {
			name, cpu, memory, available string
		}{
			{"CPU", "2500m", "32Gi", "2.25 CPU cores / 31.5 GiB memory"},
			{"memory", "8", "10Gi", "7.75 CPU cores / 9.5 GiB memory"},
			{"overcommitted", "100m", "256Mi", "-0.15 CPU cores / -0.2 GiB memory"},
		} {
			t.Run(mode+"/"+tc.name, func(t *testing.T) {
				opts := pfTestOpts()
				nodes := []corev1.Node{pfTestNode()}
				idx := 0
				needs := "5.85 CPU cores / 24.0 GiB memory"
				if multi {
					opts, nodes = pfMultiFixture()
					idx = 1
					needs = "3.00 CPU cores / 14.0 GiB memory"
				}
				nodes[idx].Status.Allocatable[corev1.ResourceCPU] = resource.MustParse(tc.cpu)
				nodes[idx].Status.Allocatable[corev1.ResourceMemory] = resource.MustParse(tc.memory)
				other := pfTestPod("250m", false)
				other.Spec.NodeName = nodes[idx].Name
				other.Spec.Containers[0].Resources = pfTestResources("250m", "512Mi")
				result := PreflightResult{StorageClass: "local-path"}
				check := pfCheckCapacity
				if multi {
					check = pfCheckMultiCapacity
				}
				err := check(opts, nodes, []pfPod{other}, nil, nil, &result)
				want := "node " + nodes[idx].Name + " lacks resources: installation needs " + needs + ", available after other workloads " + tc.available
				if err == nil || err.Error() != want {
					t.Fatalf("want readable capacity error %q, got %v", want, err)
				}
			})
		}
	}
}

func TestPreparationCapacityFailureUsesCoresAndGiB(t *testing.T) {
	for _, tc := range []struct{ name, cpu, memory string }{
		{"CPU", "6", "32Gi"},
		{"memory", "8", "25Gi"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := pfTestFixture()
			node := pfTestNode()
			node.Status.Allocatable[corev1.ResourceCPU] = resource.MustParse(tc.cpu)
			node.Status.Allocatable[corev1.ResourceMemory] = resource.MustParse(tc.memory)
			fixture["kubectl get nodes -o json"] = corev1.NodeList{Items: []corev1.Node{node}}
			fixture["kubectl get crds -o json"] = apiextv1.CustomResourceDefinitionList{}
			fixture["kubectl get deployments --all-namespaces -o json"] = appsv1.DeploymentList{}
			fixture["kubectl get mutatingwebhookconfigurations -o json"] = admissionv1.MutatingWebhookConfigurationList{}
			fixture["kubectl get validatingwebhookconfigurations -o json"] = admissionv1.ValidatingWebhookConfigurationList{}
			opts := pfTestOpts()
			opts.PrepareOperator = true
			_, err := CheckCluster(context.Background(), pfTestRunner(t, fixture), opts)
			want := "insufficient remaining capacity including the 0.50 CPU cores / 1.0 GiB memory prerequisite preparation allowance"
			if err == nil || err.Error() != want {
				t.Fatalf("want readable preparation error %q, got %v", want, err)
			}
		})
	}
}

func TestCapacityCheckAliasFailureUsesCoresAndGiBWithoutMutation(t *testing.T) {
	f := newResourceRunFixture(t)
	nodes := f.responses["kubectl get nodes -o json"].(corev1.NodeList)
	nodes.Items[0].Status.Allocatable[corev1.ResourceCPU] = resource.MustParse("8")
	f.responses["kubectl get nodes -o json"] = nodes
	f.syncCommands(t)
	out, err := f.run(t, "--oc", "10", "--check")
	want := "installation needs 13.15 CPU cores / 22.6 GiB memory, available after other workloads 8.00 CPU cores / 128.0 GiB memory"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("--oc 10 --check must report cores and GiB: %v\n%s", err, out)
	}
	trace, readErr := os.ReadFile(f.trace)
	if readErr != nil {
		t.Fatal(readErr)
	}
	for _, forbidden := range []string{"docker ", "ssh ", "scp ", "kubectl apply", "helm upgrade"} {
		if strings.Contains(string(trace), forbidden) {
			t.Fatalf("capacity check executed %q: %s", forbidden, trace)
		}
	}
	for _, path := range []string{f.state, f.marker, f.cache} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("capacity check wrote state or launched the chart", path, err)
		}
	}
}
