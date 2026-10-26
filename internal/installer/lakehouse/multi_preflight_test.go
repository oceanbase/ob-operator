package lakehouse

import (
	"context"
	"fmt"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func pfMultiFixture() (PreflightOptions, []corev1.Node) {
	opts := pfTestOpts()
	var nodes []corev1.Node
	for i := 1; i <= 3; i++ {
		name := fmt.Sprintf("worker-%d", i)
		hostname := fmt.Sprintf("host-%d", i)
		opts.NodeNames = append(opts.NodeNames, name)
		n := pfTestNode()
		n.Name = name
		n.Labels["kubernetes.io/hostname"] = hostname
		n.Status.NodeInfo.OperatingSystem = "linux"
		n.Status.NodeInfo.Architecture = "amd64"
		n.Status.NodeInfo.SystemUUID = fmt.Sprintf("uuid-%d", i)
		n.Status.Addresses = []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: fmt.Sprintf("10.0.0.%d", i)}}
		nodes = append(nodes, n)
	}
	for _, component := range []string{"observer", "logService"} {
		var topology []interface{}
		for i := 1; i <= 3; i++ {
			topology = append(topology, map[string]interface{}{"zone": fmt.Sprintf("z%d", i), "replica": 1, "nodeSelector": map[string]interface{}{"kubernetes.io/hostname": fmt.Sprintf("host-%d", i)}})
		}
		setValue(opts.Values, topology, component, "topology")
	}
	setValue(opts.Values, map[string]interface{}{"kubernetes.io/hostname": "host-1"}, "scheduling", "nodeSelector")
	setValue(opts.Values, map[string]interface{}{"kubernetes.io/hostname": "host-1"}, "dashboard", "nodeSelector")
	return opts, nodes
}

func TestMultiPreflightAccountsPerNodeAndRetainsFirstNode(t *testing.T) {
	opts, nodes := pfMultiFixture()
	other := pfTestPod("2", false)
	other.Spec.NodeName = "worker-2"
	result := PreflightResult{StorageClass: "local-path"}
	if err := pfCheckMultiCapacity(opts, nodes, []pfPod{other}, nil, nil, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Nodes) != 3 || result.NodeName != "worker-1" || result.NodeHostname != "host-1" {
		t.Fatalf("missing per-node result: %+v", result)
	}
	if result.RequiredCPUMilli != 3850 || result.Nodes[1].RequiredCPUMilli != 3000 || result.Nodes[2].RequiredCPUMilli != 3000 {
		t.Fatalf("incorrect component placement: %+v", result.Nodes)
	}
	if result.Nodes[1].AvailableCPUMilli != 6000 || result.Nodes[2].AvailableCPUMilli != 8000 {
		t.Fatalf("workload charged to wrong node: %+v", result.Nodes)
	}
	nodes[1].Status.Allocatable[corev1.ResourceCPU] = resource.MustParse("4")
	result = PreflightResult{StorageClass: "local-path"}
	err := pfCheckMultiCapacity(opts, nodes, []pfPod{other}, nil, nil, &result)
	if err == nil || !strings.Contains(err.Error(), "node worker-2 lacks resources") {
		t.Fatalf("borrowed another node's capacity: %v", err)
	}
}

func TestMultiPreflightNodeAndComponentGuards(t *testing.T) {
	cases := []struct {
		name, want string
		change     func(*PreflightOptions, []corev1.Node)
	}{
		{"duplicate uuid", "same system UUID", func(o *PreflightOptions, n []corev1.Node) { n[2].Status.NodeInfo.SystemUUID = "UUID-1" }},
		{"duplicate ip", "same InternalIP", func(o *PreflightOptions, n []corev1.Node) { n[2].Status.Addresses[0].Address = "10.0.0.1" }},
		{"duplicate", "distinct nodes", func(o *PreflightOptions, n []corev1.Node) { o.NodeNames[2] = o.NodeNames[1] }},
		{"missing", "does not exist", func(o *PreflightOptions, n []corev1.Node) { o.NodeNames[2] = "missing" }},
		{"not ready", "Ready", func(o *PreflightOptions, n []corev1.Node) { n[2].Status.Conditions = nil }},
		{"cordon", "schedulable", func(o *PreflightOptions, n []corev1.Node) { n[2].Spec.Unschedulable = true }},
		{"arm", "linux/amd64", func(o *PreflightOptions, n []corev1.Node) { n[2].Status.NodeInfo.Architecture = "arm64" }},
		{"windows", "linux/amd64", func(o *PreflightOptions, n []corev1.Node) { n[2].Status.NodeInfo.OperatingSystem = "windows" }},
		{"missing hostname", "hostname", func(o *PreflightOptions, n []corev1.Node) { delete(n[2].Labels, "kubernetes.io/hostname") }},
		{"duplicate hostname", "shared by nodes", func(o *PreflightOptions, n []corev1.Node) { n[2].Labels["kubernetes.io/hostname"] = "host-2" }},
		{"database taint", "not tolerated by logService", func(o *PreflightOptions, n []corev1.Node) {
			n[2].Spec.Taints = []corev1.Taint{{Key: "dedicated", Effect: corev1.TaintEffectNoSchedule}}
		}},
		{"dashboard taint", "Dashboard placement", func(o *PreflightOptions, n []corev1.Node) {
			n[0].Spec.Taints = []corev1.Taint{{Key: "dedicated", Effect: corev1.TaintEffectNoSchedule}}
			setValue(o.Values, []interface{}{map[string]interface{}{"key": "dedicated", "operator": "Exists", "effect": "NoSchedule"}}, "scheduling", "tolerations")
		}},
		{"duplicate zone placement", "distinct requested nodes", func(o *PreflightOptions, n []corev1.Node) {
			top := pfGet(o.Values, "observer", "topology").([]interface{})
			pfMap(top[2])["nodeSelector"] = map[string]interface{}{"kubernetes.io/hostname": "host-2"}
		}},
		{"selector", "select exactly one", func(o *PreflightOptions, n []corev1.Node) {
			top := pfGet(o.Values, "observer", "topology").([]interface{})
			pfMap(top[2])["nodeSelector"] = map[string]interface{}{"kubernetes.io/hostname": "host-3", "disk": "ssd"}
		}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			opts, nodes := pfMultiFixture()
			tt.change(&opts, nodes)
			err := pfCheckMultiCapacity(opts, nodes, nil, nil, nil, &PreflightResult{StorageClass: "local-path"})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("want %q, got %v", tt.want, err)
			}
		})
	}
}

func pfMultiClaimAndPV() (corev1.PersistentVolumeClaim, corev1.PersistentVolume) {
	class := "local-path"
	claim := corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "demo-ob-1002-z2-abcdef-data", Namespace: "demo-ns", UID: "claim-uid", OwnerReferences: []metav1.OwnerReference{{Kind: "OBServer", Name: "demo-ob-1002-z2-abcdef"}}, Annotations: map[string]string{"volume.kubernetes.io/selected-node": "worker-2"}}, Spec: corev1.PersistentVolumeClaimSpec{StorageClassName: &class, VolumeName: "pv-data"}}
	pv := corev1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "pv-data"}, Spec: corev1.PersistentVolumeSpec{ClaimRef: &corev1.ObjectReference{Name: claim.Name, Namespace: claim.Namespace, UID: claim.UID}, NodeAffinity: &corev1.VolumeNodeAffinity{Required: &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "kubernetes.io/hostname", Operator: corev1.NodeSelectorOpIn, Values: []string{"host-2"}}}}}}}}}
	return claim, pv
}

func TestMultiPreflightRetainsPerClaimBindingWithoutPods(t *testing.T) {
	opts, nodes := pfMultiFixture()
	opts.Existing = true
	claim, pv := pfMultiClaimAndPV()
	if err := pfCheckMultiCapacity(opts, nodes, nil, []corev1.PersistentVolumeClaim{claim}, []corev1.PersistentVolume{pv}, &PreflightResult{StorageClass: "local-path"}); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, want string
		change     func(*corev1.PersistentVolumeClaim, *corev1.PersistentVolume)
	}{
		{"selected node moved", "automatic migration", func(c *corev1.PersistentVolumeClaim, p *corev1.PersistentVolume) {
			c.Annotations["volume.kubernetes.io/selected-node"] = "worker-3"
		}},
		{"pv affinity moved", "affinity excludes", func(c *corev1.PersistentVolumeClaim, p *corev1.PersistentVolume) {
			p.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Values = []string{"host-3"}
		}},
		{"unknown zone", "unconfigured database zone", func(c *corev1.PersistentVolumeClaim, p *corev1.PersistentVolume) {
			c.OwnerReferences[0].Name = "demo-ob-1002-oldzone-abcdef"
		}},
		{"storage changed", "changing existing storage", func(c *corev1.PersistentVolumeClaim, p *corev1.PersistentVolume) {
			different := "other"
			c.Spec.StorageClassName = &different
		}},
		{"pv uid changed", "inconsistent PV binding", func(c *corev1.PersistentVolumeClaim, p *corev1.PersistentVolume) { p.Spec.ClaimRef.UID = "other-claim" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, p := pfMultiClaimAndPV()
			tt.change(&c, &p)
			err := pfCheckMultiCapacity(opts, nodes, nil, []corev1.PersistentVolumeClaim{c}, []corev1.PersistentVolume{p}, &PreflightResult{StorageClass: "local-path"})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("want %q, got %v", tt.want, err)
			}
		})
	}
	opts.Existing = false
	if err := pfCheckMultiCapacity(opts, nodes, nil, []corev1.PersistentVolumeClaim{claim}, []corev1.PersistentVolume{pv}, &PreflightResult{StorageClass: "local-path"}); err == nil || !strings.Contains(err.Error(), "retained PVC") {
		t.Fatalf("orphan claim accepted: %v", err)
	}
}

func TestMultiPreflightExistingPodPlacementAndOrphans(t *testing.T) {
	opts, nodes := pfMultiFixture()
	opts.Existing = true
	pod := pfTestPod("2", true)
	pod.Name = "demo-ob-1002-z2-abcdef"
	pod.Labels["ref-obzone"] = "demo-ob-1002-z2"
	pod.Spec.NodeName = "worker-2"
	if err := pfCheckMultiCapacity(opts, nodes, []pfPod{pod}, nil, nil, &PreflightResult{StorageClass: "local-path"}); err != nil {
		t.Fatal(err)
	}
	pod.Spec.NodeName = "worker-1"
	if err := pfCheckMultiCapacity(opts, nodes, []pfPod{pod}, nil, nil, &PreflightResult{StorageClass: "local-path"}); err == nil || !strings.Contains(err.Error(), "automatic migration") {
		t.Fatalf("Pod relocation accepted: %v", err)
	}
	opts.Existing = false
	if err := pfCheckMultiCapacity(opts, nodes, []pfPod{pod}, nil, nil, &PreflightResult{StorageClass: "local-path"}); err == nil || !strings.Contains(err.Error(), "already exist") {
		t.Fatalf("orphan Pod accepted: %v", err)
	}
}

func TestMultiPreflightPendingPodChargedOnlyToEligibleNode(t *testing.T) {
	opts, nodes := pfMultiFixture()
	pod := pfTestPod("4", false)
	pod.Spec.NodeName = ""
	pod.Spec.Affinity = &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchFields: []corev1.NodeSelectorRequirement{{Key: "metadata.name", Operator: corev1.NodeSelectorOpIn, Values: []string{"worker-3"}}}}}}}}
	result := PreflightResult{StorageClass: "local-path"}
	if err := pfCheckMultiCapacity(opts, nodes, []pfPod{pod}, nil, nil, &result); err != nil {
		t.Fatal(err)
	}
	if result.Nodes[0].AvailableCPUMilli != 8000 || result.Nodes[1].AvailableCPUMilli != 8000 || result.Nodes[2].AvailableCPUMilli != 4000 {
		t.Fatalf("incorrect pending demand: %+v", result.Nodes)
	}
}

func TestMultiCheckClusterReportsCompatibleOperatorRequirement(t *testing.T) {
	opts, nodes := pfMultiFixture()
	fixture := pfTestFixture()
	fixture["kubectl get oblogserviceclusters -n demo-ns -o json"] = map[string]interface{}{"items": []interface{}{}}
	fixture["kubectl get nodes -o json"] = corev1.NodeList{Items: nodes}
	result, err := CheckCluster(context.Background(), pfTestRunner(t, fixture), opts)
	if err != nil || len(result.Nodes) != 3 {
		t.Fatalf("three node check failed: %+v %v", result, err)
	}
	fixture["kubectl get crds -o json"] = apiextv1.CustomResourceDefinitionList{}
	for key := range fixture {
		if strings.Contains(key, "deployments") || strings.Contains(key, "webhookconfigurations") {
			fixture[key] = map[string]interface{}{"items": []interface{}{}}
		}
	}
	opts.PrepareOperator = true
	_, err = CheckCluster(context.Background(), pfTestRunner(t, fixture), opts)
	if err == nil || !strings.Contains(err.Error(), "install a compatible Operator before using this multi-node cluster") {
		t.Fatalf("automatic multi-node preparation accepted: %v", err)
	}
}

func TestMultiCheckClusterRejectsOrphanDatabaseCRs(t *testing.T) {
	for _, kind := range []string{"obclusters", "oblogserviceclusters"} {
		t.Run(kind, func(t *testing.T) {
			opts, nodes := pfMultiFixture()
			fixture := pfTestFixture()
			fixture["kubectl get nodes -o json"] = corev1.NodeList{Items: nodes}
			fixture["kubectl get oblogserviceclusters -n demo-ns -o json"] = map[string]interface{}{"items": []interface{}{}}
			name := "demo-ob"
			if kind == "oblogserviceclusters" {
				name = "demo-ls"
			}
			fixture["kubectl get "+kind+" -n demo-ns -o json"] = map[string]interface{}{"items": []interface{}{map[string]interface{}{"metadata": map[string]interface{}{"name": name}}}}
			_, err := CheckCluster(context.Background(), pfTestRunner(t, fixture), opts)
			if err == nil || !strings.Contains(err.Error(), "adopting orphaned resources") {
				t.Fatalf("orphan CR accepted: %v", err)
			}
		})
	}
}

func TestMultiPreflightRetainedIndependentClaimAndSuppliedClaim(t *testing.T) {
	opts, nodes := pfMultiFixture()
	opts.Existing = true
	claim, pv := pfMultiClaimAndPV()
	claim.OwnerReferences = nil
	claim.Labels = map[string]string{"ref-obcluster": "demo-ob", "ref-obzone": "demo-ob-1002-z2"}
	if err := pfCheckMultiCapacity(opts, nodes, nil, []corev1.PersistentVolumeClaim{claim}, []corev1.PersistentVolume{pv}, &PreflightResult{StorageClass: "local-path"}); err != nil {
		t.Fatalf("independent-lifecycle PVC rejected: %v", err)
	}
	claim.Annotations["volume.kubernetes.io/selected-node"] = "worker-1"
	if err := pfCheckMultiCapacity(opts, nodes, nil, []corev1.PersistentVolumeClaim{claim}, []corev1.PersistentVolume{pv}, &PreflightResult{StorageClass: "local-path"}); err == nil || !strings.Contains(err.Error(), "automatic migration") {
		t.Fatalf("independent-lifecycle relocation accepted: %v", err)
	}
	opts.Existing = false
	setValue(opts.Values, "provided-minio", "minio", "existingClaim")
	supplied := corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "provided-minio", Namespace: opts.Namespace, Annotations: map[string]string{"volume.kubernetes.io/selected-node": "worker-1"}}}
	if err := pfCheckMultiCapacity(opts, nodes, nil, []corev1.PersistentVolumeClaim{supplied}, nil, &PreflightResult{StorageClass: "local-path"}); err != nil {
		t.Fatalf("compatible supplied claim rejected: %v", err)
	}
	supplied.Annotations["volume.kubernetes.io/selected-node"] = "worker-2"
	if err := pfCheckMultiCapacity(opts, nodes, nil, []corev1.PersistentVolumeClaim{supplied}, nil, &PreflightResult{StorageClass: "local-path"}); err == nil || !strings.Contains(err.Error(), "automatic migration") {
		t.Fatalf("supplied claim relocation accepted: %v", err)
	}
	if err := pfCheckMultiCapacity(opts, nodes, nil, nil, nil, &PreflightResult{StorageClass: "local-path"}); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("missing supplied claim accepted: %v", err)
	}
}

func TestMultiPreflightRejectsRetainedPVWithoutClaim(t *testing.T) {
	opts, nodes := pfMultiFixture()
	_, pv := pfMultiClaimAndPV()
	err := pfCheckMultiCapacity(opts, nodes, nil, nil, []corev1.PersistentVolume{pv}, &PreflightResult{StorageClass: "local-path"})
	if err == nil || !strings.Contains(err.Error(), "references missing PVC") {
		t.Fatalf("orphan retained PV accepted: %v", err)
	}
}

func TestMultiPreflightExternalStorageDoesNotChargeMinio(t *testing.T) {
	opts, nodes := pfMultiFixture()
	setValue(opts.Values, "external", "objectStorage", "mode")
	setValue(opts.Values, false, "objectStorage", "createBuckets")
	result := PreflightResult{StorageClass: "local-path"}
	if err := pfCheckMultiCapacity(opts, nodes, nil, nil, nil, &result); err != nil {
		t.Fatal(err)
	}
	if result.RequiredCPUMilli != 3700 {
		t.Fatalf("external storage still charged MinIO or init: %+v", result)
	}
}
