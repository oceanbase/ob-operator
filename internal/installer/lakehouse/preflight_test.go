package lakehouse

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func pfTestValues() map[string]interface{} {
	var values map[string]interface{}
	_ = json.Unmarshal([]byte(`{
	 "storageClass":"local-path", "scheduling":{"nodeSelector":{},"tolerations":[]},
	 "objectStorage":{"mode":"bundled"},
	 "minio":{"resources":{"requests":{"cpu":"100m","memory":"512Mi"}}},
	 "initialization":{"resources":{"requests":{"cpu":"50m","memory":"64Mi"}}},
	 "logService":{"resource":{"cpu":"1","memory":"4Gi"},"topology":[{"zone":"z1","replica":1},{"zone":"z2","replica":1},{"zone":"z3","replica":1}]},
	 "observer":{"clusterId":1002,"resource":{"cpu":"2","memory":"10Gi"},"topology":[{"zone":"z1","replica":1}]},
	 "dashboard":{"enabled":true}
	}`), &values)
	return values
}

func pfTestOpts() PreflightOptions {
	return PreflightOptions{Release: "demo", Namespace: "demo-ns", Values: pfTestValues()}
}

func pfTestNode() corev1.Node {
	return corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "worker-1", Labels: map[string]string{"kubernetes.io/hostname": "worker-1"}}, Status: corev1.NodeStatus{
		Allocatable: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("8"), corev1.ResourceMemory: resource.MustParse("32Gi")},
		Conditions:  []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}},
		NodeInfo:    corev1.NodeSystemInfo{ContainerRuntimeVersion: "containerd://2.2.7", SystemUUID: "uuid"},
		Addresses:   []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "10.0.0.1"}, {Type: corev1.NodeExternalIP, Address: "192.0.2.1"}},
	}}
}

func pfTestResources(cpu, memory string) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse(cpu), corev1.ResourceMemory: resource.MustParse(memory)}}
}

func pfTestPod(cpu string, own bool) pfPod {
	p := pfPod{Pod: corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "workload", Namespace: "other"}, Spec: corev1.PodSpec{NodeName: "worker-1", Containers: []corev1.Container{{Name: "main", Resources: pfTestResources(cpu, "1Gi")}}}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}}
	if own {
		p.Namespace = "demo-ns"
		p.Labels = map[string]string{"ref-obcluster": "demo-ob"}
	}
	return p
}

func pfTestOperator() ([]apiextv1.CustomResourceDefinition, []appsv1.Deployment, []admissionv1.MutatingWebhookConfiguration, []admissionv1.ValidatingWebhookConfiguration, []corev1.Endpoints) {
	var crds []apiextv1.CustomResourceDefinition
	for _, name := range []string{"obclusters", "oblogserviceclusters", "oblogservicezones", "oblogservicenodes"} {
		crd := apiextv1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: name + ".oceanbase.oceanbase.com"}, Spec: apiextv1.CustomResourceDefinitionSpec{Group: "oceanbase.oceanbase.com", Names: apiextv1.CustomResourceDefinitionNames{Plural: name}, Versions: []apiextv1.CustomResourceDefinitionVersion{{Name: "v1alpha1", Served: true, Schema: &apiextv1.CustomResourceValidation{OpenAPIV3Schema: &apiextv1.JSONSchemaProps{Properties: map[string]apiextv1.JSONSchemaProps{"spec": {Properties: map[string]apiextv1.JSONSchemaProps{"deploymentMode": {Enum: []apiextv1.JSON{{Raw: []byte(`"shared_storage"`)}}}}}}}}}}}, Status: apiextv1.CustomResourceDefinitionStatus{Conditions: []apiextv1.CustomResourceDefinitionCondition{{Type: apiextv1.Established, Status: apiextv1.ConditionTrue}}}}
		crds = append(crds, crd)
	}
	deployments := []appsv1.Deployment{{ObjectMeta: metav1.ObjectMeta{Name: "oceanbase-controller-manager", Namespace: "oceanbase-system", Generation: 1}, Status: appsv1.DeploymentStatus{ObservedGeneration: 1, ReadyReplicas: 1, AvailableReplicas: 1}}}
	client := admissionv1.WebhookClientConfig{CABundle: []byte("ca"), Service: &admissionv1.ServiceReference{Name: "webhook", Namespace: "oceanbase-system"}}
	rules := []admissionv1.RuleWithOperations{{Operations: []admissionv1.OperationType{admissionv1.Create, admissionv1.Update}, Rule: admissionv1.Rule{APIGroups: []string{"oceanbase.oceanbase.com"}, APIVersions: []string{"v1alpha1"}, Resources: []string{"obclusters", "oblogserviceclusters"}}}}
	mutating := []admissionv1.MutatingWebhookConfiguration{{Webhooks: []admissionv1.MutatingWebhook{{ClientConfig: client, Rules: rules}}}}
	validating := []admissionv1.ValidatingWebhookConfiguration{{Webhooks: []admissionv1.ValidatingWebhook{{ClientConfig: client, Rules: rules}}}}
	endpoints := []corev1.Endpoints{{ObjectMeta: metav1.ObjectMeta{Name: "webhook", Namespace: "oceanbase-system"}, Subsets: []corev1.EndpointSubset{{Addresses: []corev1.EndpointAddress{{IP: "10.0.0.2"}}, Ports: []corev1.EndpointPort{{Port: 9443}}}}}}
	return crds, deployments, mutating, validating, endpoints
}

func pfTestFixture() map[string]interface{} {
	crds, deployments, mutating, validating, endpoints := pfTestOperator()
	return map[string]interface{}{
		"kubectl version -o json":                             map[string]interface{}{"serverVersion": map[string]string{"gitVersion": "v1.34.12+k3s1"}},
		"helm version --template {{.Version}}":                "v3.22.1",
		"kubectl get storageclasses -o json":                  storagev1.StorageClassList{Items: []storagev1.StorageClass{{ObjectMeta: metav1.ObjectMeta{Name: "local-path"}}}},
		"kubectl get nodes -o json":                           corev1.NodeList{Items: []corev1.Node{pfTestNode()}},
		"kubectl get pods --all-namespaces -o json":           corev1.PodList{},
		"kubectl get pvc --all-namespaces -o json":            corev1.PersistentVolumeClaimList{},
		"kubectl get pv -o json":                              corev1.PersistentVolumeList{},
		"kubectl get crds -o json":                            apiextv1.CustomResourceDefinitionList{Items: crds},
		"kubectl get deployments --all-namespaces -o json":    appsv1.DeploymentList{Items: deployments},
		"kubectl get mutatingwebhookconfigurations -o json":   admissionv1.MutatingWebhookConfigurationList{Items: mutating},
		"kubectl get validatingwebhookconfigurations -o json": admissionv1.ValidatingWebhookConfigurationList{Items: validating},
		"kubectl get endpoints --all-namespaces -o json":      corev1.EndpointsList{Items: endpoints},
		"kubectl get obclusters -n demo-ns -o json":           map[string]interface{}{"items": []interface{}{}},
	}
}

func TestCheckClusterRejectsSharedStorageObserverImageChangeBeforeMutation(t *testing.T) {
	for _, changed := range []string{"qa/observer:new", "docker.io/qa/observer:original"} {
		t.Run(changed, func(t *testing.T) {
			fixture := pfTestFixture()
			fixture["kubectl get obclusters -n demo-ns -o json"] = map[string]interface{}{"items": []interface{}{map[string]interface{}{
				"metadata": map[string]interface{}{"name": "demo-ob", "namespace": "demo-ns"},
				"spec":     map[string]interface{}{"deploymentMode": "shared_storage", "observer": map[string]interface{}{"image": "qa/observer:original"}},
			}}}
			opts := pfTestOpts()
			opts.Existing = true
			setValue(opts.Values, changed, "observer", "image")
			_, err := CheckCluster(context.Background(), pfTestRunner(t, fixture), opts)
			if err == nil || !strings.Contains(err.Error(), "shared_storage") || !strings.Contains(err.Error(), "observer.image") {
				t.Fatalf("shared-storage image change was not rejected read-only: %v", err)
			}
			setValue(opts.Values, "qa/observer:original", "observer", "image")
			if _, err := CheckCluster(context.Background(), pfTestRunner(t, fixture), opts); err != nil {
				t.Fatalf("unchanged Observer image rejected: %v", err)
			}
		})
	}
}

func TestCheckClusterDoesNotTreatObserverReadDeniedAsAbsent(t *testing.T) {
	fixture := pfTestFixture()
	fixture["kubectl get obclusters -n demo-ns -o json"] = errors.New("Forbidden")
	if _, err := CheckCluster(context.Background(), pfTestRunner(t, fixture), pfTestOpts()); err == nil || !strings.Contains(err.Error(), "cannot read") {
		t.Fatalf("failed Observer read was ignored: %v", err)
	}
}

func pfTestRunner(t *testing.T, fixture map[string]interface{}) RunCommand {
	t.Helper()
	return func(ctx context.Context, tool string, args ...string) ([]byte, error) {
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) > pfCommandTimeout {
			t.Fatal("missing bounded command context")
		}
		if tool == "kubectl" {
			if len(args) == 0 || args[0] != "--request-timeout=10s" {
				t.Fatal("missing API request timeout")
			}
			args = args[1:]
		}
		key := tool + " " + strings.Join(args, " ")
		value, ok := fixture[key]
		if !ok {
			t.Fatalf("unexpected or mutating command: %s", key)
		}
		if err, ok := value.(error); ok {
			return nil, err
		}
		if s, ok := value.(string); ok {
			return []byte(s), nil
		}
		return json.Marshal(value)
	}
}

func TestCheckClusterReady(t *testing.T) {
	got, err := CheckCluster(context.Background(), pfTestRunner(t, pfTestFixture()), pfTestOpts())
	if err != nil {
		t.Fatal(err)
	}
	if got.NodeName != "worker-1" || got.RequiredCPUMilli != 5850 || got.AvailableCPUMilli != 8000 || got.OperatorMissing {
		t.Fatalf("unexpected result: %+v", got)
	}
	if got.NodeRuntimeVersion != "containerd://2.2.7" || got.NodeInternalIP != "10.0.0.1" || got.NodeExternalIP != "192.0.2.1" || got.NodeSystemUUID != "uuid" {
		t.Fatalf("node identity missing: %+v", got)
	}
}

func TestCheckClusterCollectsVersionStorageAndOperatorProblems(t *testing.T) {
	fixture := pfTestFixture()
	fixture["kubectl version -o json"] = map[string]interface{}{"serverVersion": map[string]string{"gitVersion": "v1.26.9"}}
	fixture["helm version --template {{.Version}}"] = "v3.11.4"
	fixture["kubectl get storageclasses -o json"] = storagev1.StorageClassList{}
	fixture["kubectl get endpoints --all-namespaces -o json"] = corev1.EndpointsList{}
	_, err := CheckCluster(context.Background(), pfTestRunner(t, fixture), pfTestOpts())
	for _, part := range []string{"Kubernetes >=1.27", "Helm >=3.12", "StorageClass", "admission webhooks"} {
		if err == nil || !strings.Contains(err.Error(), part) {
			t.Fatalf("missing %s in %v", part, err)
		}
	}
}

func TestCheckClusterCannotTreatAccessDeniedAsMissingOperator(t *testing.T) {
	fixture := pfTestFixture()
	fixture["kubectl get crds -o json"] = errors.New("forbidden")
	opts := pfTestOpts()
	opts.PrepareOperator = true
	got, err := CheckCluster(context.Background(), pfTestRunner(t, fixture), opts)
	if err == nil || got.OperatorMissing {
		t.Fatalf("denied read allowed bootstrap: %+v %v", got, err)
	}
}

func TestCheckClusterAbsentOperatorPreparation(t *testing.T) {
	fixture := pfTestFixture()
	fixture["kubectl get crds -o json"] = apiextv1.CustomResourceDefinitionList{}
	fixture["kubectl get deployments --all-namespaces -o json"] = appsv1.DeploymentList{}
	fixture["kubectl get mutatingwebhookconfigurations -o json"] = admissionv1.MutatingWebhookConfigurationList{}
	fixture["kubectl get validatingwebhookconfigurations -o json"] = admissionv1.ValidatingWebhookConfigurationList{}
	for _, prepare := range []bool{false, true} {
		opts := pfTestOpts()
		opts.PrepareOperator = prepare
		got, err := CheckCluster(context.Background(), pfTestRunner(t, fixture), opts)
		if !got.OperatorMissing || (err == nil) != prepare {
			t.Fatalf("prepare %v: %+v %v", prepare, got, err)
		}
	}
}

func TestCheckClusterRejectsAutomaticOperatorPreparationBeforeMultiNodeMutation(t *testing.T) {
	fixture := pfTestFixture()
	fixture["kubectl get crds -o json"] = apiextv1.CustomResourceDefinitionList{}
	fixture["kubectl get deployments --all-namespaces -o json"] = appsv1.DeploymentList{}
	fixture["kubectl get mutatingwebhookconfigurations -o json"] = admissionv1.MutatingWebhookConfigurationList{}
	fixture["kubectl get validatingwebhookconfigurations -o json"] = admissionv1.ValidatingWebhookConfigurationList{}
	second := pfTestNode()
	second.Name = "worker-2"
	second.Labels = map[string]string{"kubernetes.io/hostname": "worker-2"}
	fixture["kubectl get nodes -o json"] = corev1.NodeList{Items: []corev1.Node{pfTestNode(), second}}
	opts := pfTestOpts()
	opts.NodeName = "worker-1"
	opts.PrepareOperator = true
	_, err := CheckCluster(context.Background(), pfTestRunner(t, fixture), opts)
	if err == nil || !strings.Contains(err.Error(), "Operator preparation supports only") {
		t.Fatalf("uninstallable prerequisites passed --check: %v", err)
	}
}

func TestOperatorPartialOldUnreadyAndMissingWebhooks(t *testing.T) {
	for _, scenario := range []string{"partial", "old", "unready", "unestablished", "empty-ca", "delete-only", "missing-endpoint"} {
		t.Run(scenario, func(t *testing.T) {
			crds, deployments, mutating, validating, endpoints := pfTestOperator()
			switch scenario {
			case "partial":
				crds = crds[:1]
			case "old":
				crds[0].Spec.Versions[0].Schema = nil
			case "unready":
				deployments[0].Status.AvailableReplicas = 0
			case "unestablished":
				crds[1].Status.Conditions = nil
			case "empty-ca":
				mutating[0].Webhooks[0].ClientConfig.CABundle = nil
			case "delete-only":
				mutating[0].Webhooks[0].Rules[0].Operations = []admissionv1.OperationType{admissionv1.Delete}
			case "missing-endpoint":
				endpoints = nil
			}
			missing, err := pfCheckOperator(crds, deployments, mutating, validating, endpoints)
			if missing || err == nil {
				t.Fatalf("%s incorrectly accepted as absent/ready: %v %v", scenario, missing, err)
			}
		})
	}
}

func TestCapacityUsesActualValuesAndExistingDelta(t *testing.T) {
	for _, existing := range []bool{false, true} {
		opts := pfTestOpts()
		opts.Existing = existing
		pod := pfTestPod("5", true)
		result := PreflightResult{StorageClass: "local-path"}
		err := pfCheckCapacity(opts, []corev1.Node{pfTestNode()}, []pfPod{pod, pfTestPod("1", false)}, nil, nil, &result)
		if existing {
			if err != nil || result.AvailableCPUMilli != 7000 {
				t.Fatalf("existing counted twice: %+v %v", result, err)
			}
		} else if err == nil {
			t.Fatal("fresh accepted existing resources")
		}
	}
	opts := pfTestOpts()
	pfMap(pfGet(opts.Values, "logService", "resource"))["cpu"] = "2"
	result := PreflightResult{StorageClass: "local-path"}
	if err := pfCheckCapacity(opts, []corev1.Node{pfTestNode()}, nil, nil, nil, &result); err == nil || result.RequiredCPUMilli != 8850 {
		t.Fatalf("custom CPU ignored: %+v %v", result, err)
	}
}

func TestNodePlacementAndTaintsFailClosed(t *testing.T) {
	for _, scenario := range []string{"not-ready", "cordoned", "tainted", "wrong-selector", "two-workers", "migrate", "multiple-existing-nodes"} {
		t.Run(scenario, func(t *testing.T) {
			opts := pfTestOpts()
			node := pfTestNode()
			nodes := []corev1.Node{node}
			var pods []pfPod
			switch scenario {
			case "not-ready":
				nodes[0].Status.Conditions[0].Status = corev1.ConditionFalse
			case "cordoned":
				nodes[0].Spec.Unschedulable = true
			case "tainted":
				nodes[0].Spec.Taints = []corev1.Taint{{Key: "dedicated", Value: "db", Effect: corev1.TaintEffectNoSchedule}}
			case "wrong-selector":
				pfMap(pfGet(opts.Values, "scheduling"))["nodeSelector"] = map[string]interface{}{"pool": "different"}
			case "two-workers":
				node.Name = "worker-2"
				nodes = append(nodes, node)
			case "migrate":
				opts.Existing = true
				opts.NodeName = "worker-2"
				pods = []pfPod{pfTestPod("2", true)}
			case "multiple-existing-nodes":
				opts.Existing = true
				second := pfTestPod("2", true)
				second.Spec.NodeName = "worker-2"
				pods = []pfPod{pfTestPod("2", true), second}
			}
			if err := pfCheckCapacity(opts, nodes, pods, nil, nil, &PreflightResult{StorageClass: "local-path"}); err == nil {
				t.Fatal("unsupported placement accepted")
			}
		})
	}
}

func TestRetainedPVCPreventsMovingPodlessInstallation(t *testing.T) {
	opts := pfTestOpts()
	opts.Existing = true
	opts.NodeName = "worker-2"
	claim := corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "data", Namespace: opts.Namespace, Labels: map[string]string{"app.kubernetes.io/instance": opts.Release}, Annotations: map[string]string{"volume.kubernetes.io/selected-node": "worker-1"}}}
	err := pfCheckCapacity(opts, []corev1.Node{pfTestNode()}, nil, []corev1.PersistentVolumeClaim{claim}, nil, &PreflightResult{StorageClass: "local-path"})
	if err == nil || !strings.Contains(err.Error(), "migration") {
		t.Fatalf("retained volume moved: %v", err)
	}
}

func TestPodRequestsInitSidecarOverheadAndNewPodResources(t *testing.T) {
	p := pfTestPod("1", false)
	always := corev1.ContainerRestartPolicyAlways
	p.Spec.InitContainers = []corev1.Container{{Name: "sidecar", RestartPolicy: &always, Resources: pfTestResources("500m", "100Mi")}, {Name: "setup", Resources: pfTestResources("2", "2Gi")}}
	p.Spec.Overhead = corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m"), corev1.ResourceMemory: resource.MustParse("10Mi")}
	got := pfPodRequests(p)
	if got.cpu != 2600 || got.memory != 2*1024*1024*1024+110*1024*1024 {
		t.Fatalf("init request calculation: %+v", got)
	}
	var newer pfPod
	if err := json.Unmarshal([]byte(`{"spec":{"containers":[{"name":"app","resources":{"requests":{"cpu":"1","memory":"1Gi"}}}],"resources":{"requests":{"cpu":"3","memory":"4Gi"}}}}`), &newer); err != nil {
		t.Fatal(err)
	}
	if got := pfPodRequests(newer); got.cpu != 3000 || got.memory != 4*1024*1024*1024 {
		t.Fatalf("pod-level requests ignored: %+v", got)
	}
}

func TestStorageClassDefaultSelection(t *testing.T) {
	values := pfTestValues()
	values["storageClass"] = ""
	sc := storagev1.StorageClass{ObjectMeta: metav1.ObjectMeta{Name: "default", Annotations: map[string]string{"storageclass.kubernetes.io/is-default-class": "true"}}}
	if got, err := pfStorageClass(values, []storagev1.StorageClass{sc}); err != nil || got != "default" {
		t.Fatalf("default selection: %q %v", got, err)
	}
	if _, err := pfStorageClass(values, []storagev1.StorageClass{sc, sc}); err == nil {
		t.Fatal("ambiguous default accepted")
	}
}

func TestDisabledDashboardDoesNotRequireItsStorageClass(t *testing.T) {
	values := pfTestValues()
	pfMap(values["dashboard"])["enabled"] = false
	pfMap(values["dashboard"])["prometheus"] = map[string]interface{}{"persistence": map[string]interface{}{"storageClassName": "unused"}}
	if _, err := pfStorageClass(values, []storagev1.StorageClass{{ObjectMeta: metav1.ObjectMeta{Name: "local-path"}}}); err != nil {
		t.Fatalf("disabled Dashboard required missing storage class: %v", err)
	}
}

func TestExistingClaimPresenceAndPlacement(t *testing.T) {
	opts := pfTestOpts()
	pfMap(opts.Values["minio"])["existingClaim"] = "provided"
	node := pfTestNode()
	result := PreflightResult{StorageClass: "local-path"}
	if err := pfCheckCapacity(opts, []corev1.Node{node}, nil, nil, nil, &result); err == nil || !strings.Contains(err.Error(), "existingClaim") {
		t.Fatalf("missing supplied claim accepted: %v", err)
	}
	claim := corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "provided", Namespace: opts.Namespace, Annotations: map[string]string{"volume.kubernetes.io/selected-node": "worker-2"}}}
	opts.NodeName = "worker-1"
	if err := pfCheckCapacity(opts, []corev1.Node{node}, nil, []corev1.PersistentVolumeClaim{claim}, nil, &result); err == nil || !strings.Contains(err.Error(), "migration") {
		t.Fatalf("supplied claim moved: %v", err)
	}
	claim.Annotations["volume.kubernetes.io/selected-node"] = "worker-1"
	if err := pfCheckCapacity(opts, []corev1.Node{node}, nil, []corev1.PersistentVolumeClaim{claim}, nil, &result); err != nil {
		t.Fatalf("valid user claim rejected: %v", err)
	}
}

func TestDashboardPlacementAndSeparateStorageClass(t *testing.T) {
	opts := pfTestOpts()
	pfMap(opts.Values["dashboard"])["nodeSelector"] = map[string]interface{}{"kubernetes.io/hostname": "different"}
	result := PreflightResult{StorageClass: "local-path"}
	if err := pfCheckCapacity(opts, []corev1.Node{pfTestNode()}, nil, nil, nil, &result); err == nil {
		t.Fatal("Dashboard selector ignored")
	}
	pfMap(opts.Values["dashboard"])["nodeSelector"] = map[string]interface{}{}
	pfMap(opts.Values["dashboard"])["prometheus"] = map[string]interface{}{"persistence": map[string]interface{}{"storageClassName": "monitoring"}}
	opts.Existing = true
	class := "monitoring"
	claim := corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "demo-prometheus-pvc", Namespace: opts.Namespace, Labels: map[string]string{"app.kubernetes.io/instance": opts.Release, "app.kubernetes.io/name": "oceanbase-lakehouse"}}, Spec: corev1.PersistentVolumeClaimSpec{StorageClassName: &class}}
	if err := pfCheckCapacity(opts, []corev1.Node{pfTestNode()}, nil, []corev1.PersistentVolumeClaim{claim}, nil, &result); err != nil {
		t.Fatalf("valid dashboard class override rejected: %v", err)
	}
}

func TestExternalReadOnlyStorageDoesNotReserveMinioOrInit(t *testing.T) {
	values := pfTestValues()
	pfMap(values["objectStorage"])["mode"] = "external"
	demand, _, _, err := pfDemand(values)
	if err != nil || demand.cpu != 5700 {
		t.Fatalf("external storage demand: %+v %v", demand, err)
	}
}

func TestLargeQuantitiesCannotOverflowIntoCapacitySuccess(t *testing.T) {
	values := pfTestValues()
	pfMap(pfGet(values, "logService", "resource"))["memory"] = "9223372036854775807"
	pfMap(pfGet(values, "logService", "topology").([]interface{})[0])["replica"] = 2
	if _, _, _, err := pfDemand(values); err == nil {
		t.Fatal("quantity multiplication overflow accepted")
	}
}

func TestVersionParsing(t *testing.T) {
	for _, v := range []string{"v1.27.0", "v1.34.12+k3s1", "1.28.2"} {
		if !pfVersionAtLeast(v, 1, 27) {
			t.Fatal(v)
		}
	}
	for _, v := range []string{"v1.26.9", "", "garbage", "v1.x.0"} {
		if pfVersionAtLeast(v, 1, 27) {
			t.Fatal(v)
		}
	}
}
