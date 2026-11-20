package logservice

import (
	"context"
	api "github.com/oceanbase/ob-operator/api/types"
	ob "github.com/oceanbase/ob-operator/api/v1alpha1"
	obconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"testing"
)

func validCreate() *CreateRequest {
	return &CreateRequest{Namespace: "test", Name: "ls-test", Spec: ob.OBLogServiceClusterSpec{
		ClusterId: 123, LogService: &api.LogServiceTemplate{Image: "ls:1.3.0", Resource: api.ResourceSpec{Cpu: resource.MustParse("1"), Memory: resource.MustParse("8Gi")}, Storage: &api.LogServiceStorageSpec{StoreStorage: &api.StorageSpec{Size: resource.MustParse("30Gi")}, LogStorage: &api.StorageSpec{Size: resource.MustParse("15Gi")}}},
		Topology: []api.LogServiceZoneTopology{{Zone: "zone1", Replica: 3, RpcPort: 50051, HttpPort: 50052}}, ObjectStoreURL: api.ObjectStoreConfig{BucketURL: "s3://test/ls?host=minio:9000", SecretRef: corev1.LocalObjectReference{Name: "s3-secret"}},
	}}
}
func fixture(t *testing.T) (*Service, *unstructured.Unstructured) {
	t.Helper()
	p := validCreate()
	r := &ob.OBLogServiceCluster{TypeMeta: metav1.TypeMeta{APIVersion: "oceanbase.oceanbase.com/v1alpha1", Kind: "OBLogServiceCluster"}, ObjectMeta: metav1.ObjectMeta{Name: p.Name, Namespace: p.Namespace, ResourceVersion: "10"}, Spec: p.Spec, Status: ob.OBLogServiceClusterStatus{Status: "running"}}
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(r)
	if err != nil {
		t.Fatal(err)
	}
	u := &unstructured.Unstructured{Object: obj}
	d := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), map[schema.GroupVersionResource]string{ClusterGVR: "OBLogServiceClusterList", NodeGVR: "OBLogServiceNodeList", OBGVR: "OBClusterList"}, u)
	core := kubefake.NewSimpleClientset(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test"}}, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "s3-secret", Namespace: "test"}, Data: map[string][]byte{"access_id": []byte("id"), "access_key": []byte("do-not-leak")}})
	return &Service{Dynamic: d, Core: core}, u
}
func TestCreateValidation(t *testing.T) {
	cases := map[string]func(*CreateRequest){
		"namespace": func(p *CreateRequest) { p.Namespace = "" }, "zero ID": func(p *CreateRequest) { p.Spec.ClusterId = 0 }, "unsafe ID": func(p *CreateRequest) { p.Spec.ClusterId = 9007199254740992 },
		"no template": func(p *CreateRequest) { p.Spec.LogService = nil }, "zero CPU": func(p *CreateRequest) { p.Spec.LogService.Resource.Cpu = resource.Quantity{} },
		"zero storage": func(p *CreateRequest) { p.Spec.LogService.Storage.LogStorage.Size = resource.Quantity{} }, "no zones": func(p *CreateRequest) { p.Spec.Topology = nil },
		"duplicate zone": func(p *CreateRequest) { p.Spec.Topology = append(p.Spec.Topology, p.Spec.Topology[0]) }, "zero replica": func(p *CreateRequest) { p.Spec.Topology[0].Replica = 0 },
		"invalid port": func(p *CreateRequest) { p.Spec.Topology[0].HttpPort = 70000 }, "same port": func(p *CreateRequest) { p.Spec.Topology[0].HttpPort = 50051 },
		"inline secret": func(p *CreateRequest) { p.Spec.ObjectStoreURL.BucketURL += "&access_key=secret" }, "URL userinfo": func(p *CreateRequest) { p.Spec.ObjectStoreURL.BucketURL = "s3://user:pass@bucket" }, "missing Secret": func(p *CreateRequest) { p.Spec.ObjectStoreURL.SecretRef.Name = "" },
	}
	if err := ValidateCreate(validCreate()); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			p := validCreate()
			mutate(p)
			if ValidateCreate(p) == nil {
				t.Fatal("invalid request accepted")
			}
		})
	}
}
func TestCreateDependenciesAndNoSecretDisclosure(t *testing.T) {
	ctx := context.Background()
	s, _ := fixture(t)
	p := validCreate()
	p.Name = "new-ls"
	p.Spec.ClusterId = 124
	item, err := s.Create(ctx, p)
	if err != nil {
		t.Fatal(err)
	}
	if item.Spec.ObjectStoreURL.SecretRef.Name != "s3-secret" {
		t.Fatal("missing reference")
	}
	list, err := s.List(ctx, "test")
	if err != nil || len(list) != 2 {
		t.Fatalf("list: %v %v", list, err)
	}
	p.Name = "duplicate-id"
	if _, err := s.Create(ctx, p); err == nil {
		t.Fatal("duplicate ID accepted")
	}
	p.Spec.ClusterId = 125
	p.Spec.ObjectStoreURL.SecretRef.Name = "missing"
	if _, err := s.Create(ctx, p); err == nil {
		t.Fatal("missing Secret accepted")
	}
	list, _ = s.List(ctx, "test")
	if len(list) != 2 {
		t.Fatal("failed create mutated resources")
	}
}
func TestCreateBootstrapReplicaCount(t *testing.T) {
	for _, count := range []int{1, 2, 3, 4} {
		p := validCreate()
		p.Spec.Topology[0].Replica = count
		if err := ValidateCreate(p); (err == nil) != (count == 3) {
			t.Fatalf("replicas=%d: %v", count, err)
		}
	}
	p := validCreate()
	p.Spec.Topology = nil
	for _, zone := range []string{"zone1", "zone2", "zone3"} {
		p.Spec.Topology = append(p.Spec.Topology, api.LogServiceZoneTopology{Zone: zone, Replica: 1, RpcPort: 50051, HttpPort: 50052})
	}
	if err := ValidateCreate(p); err != nil {
		t.Fatal(err)
	}
	s, _ := fixture(t)
	p.Spec.Topology = p.Spec.Topology[:1]
	if _, err := s.Create(context.Background(), p); err == nil {
		t.Fatal("invalid bootstrap accepted")
	}
	if got := s.Dynamic.(*dynamicfake.FakeDynamicClient).Actions(); len(got) != 0 {
		t.Fatal("invalid topology performed Kubernetes requests")
	}
}
func TestScaleOnlyExistingReplicasAndVersion(t *testing.T) {
	for _, test := range []struct {
		name     string
		version  string
		replicas map[string]int
		wantErr  bool
	}{
		{"out", "10", map[string]int{"zone1": 4}, false}, {"in", "10", map[string]int{"zone1": 2}, false}, {"stale", "9", map[string]int{"zone1": 2}, true},
		{"zero", "10", map[string]int{"zone1": 0}, true}, {"add zone", "10", map[string]int{"zone1": 1, "zone2": 1}, true}, {"rename", "10", map[string]int{"zone2": 1}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			s, _ := fixture(t)
			v, err := s.Scale(context.Background(), "test", "ls-test", &ScaleRequest{ResourceVersion: test.version, Replicas: test.replicas})
			if (err != nil) != test.wantErr {
				t.Fatalf("%v", err)
			}
			if err == nil && (v.Spec.LogService.Image != "ls:1.3.0" || v.Spec.LogService.Storage.LogStorage.Size.String() != "15Gi") {
				t.Fatal("immutable template changed")
			}
		})
	}
}
func TestDeleteGuardAndPreconditions(t *testing.T) {
	ctx := context.Background()
	s, u := fixture(t)
	p := &DeleteRequest{ResourceVersion: "10", ConfirmName: "wrong"}
	if _, err := s.Delete(ctx, "test", "ls-test", p); err == nil {
		t.Fatal("name confirmation missing")
	}
	p.ConfirmName = "ls-test"
	p.ResourceVersion = "9"
	if _, err := s.Delete(ctx, "test", "ls-test", p); err == nil {
		t.Fatal("stale deletion accepted")
	}
	p.ResourceVersion = "10"
	ref := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "oceanbase.oceanbase.com/v1alpha1", "kind": "OBCluster", "metadata": map[string]interface{}{"name": "consumer", "namespace": "test"}, "spec": map[string]interface{}{"logServiceRef": map[string]interface{}{"name": "ls-test"}}}}
	s.Dynamic.Resource(OBGVR).Namespace("test").Create(ctx, ref, metav1.CreateOptions{})
	if _, err := s.Delete(ctx, "test", "ls-test", p); err == nil {
		t.Fatal("referenced LS deleted")
	}
	s.Dynamic.Resource(OBGVR).Namespace("test").Delete(ctx, "consumer", metav1.DeleteOptions{})
	u.SetAnnotations(map[string]string{obconst.AnnotationsIgnoreDeletion: "true"})
	s.Dynamic.Resource(ClusterGVR).Namespace("test").Update(ctx, u, metav1.UpdateOptions{})
	if _, err := s.Delete(ctx, "test", "ls-test", p); err == nil {
		t.Fatal("protected LS deleted")
	}
	u.SetAnnotations(nil)
	s.Dynamic.Resource(ClusterGVR).Namespace("test").Update(ctx, u, metav1.UpdateOptions{})
	if ok, err := s.Delete(ctx, "test", "ls-test", p); err != nil || !ok {
		t.Fatalf("safe delete: %v", err)
	}
}
func TestPublicURLRedaction(t *testing.T) {
	inline := validCreate()
	inline.Spec.ObjectStoreURL.BucketURL = "s3://bucket?host=http://user:pass@minio:9000"
	if ValidateCreate(inline) == nil {
		t.Fatal("endpoint credentials accepted")
	}
	if got := publicItem(&ob.OBLogServiceCluster{Spec: inline.Spec}); got.Spec.ObjectStoreURL.BucketURL != "s3://bucket" {
		t.Fatal("endpoint credentials exposed")
	}
	original := validCreate()
	original.Spec.ObjectStoreURL.BucketURL = "s3://bucket?host=http://minio:9000&s3_region=us-east-1"
	if got := publicItem(&ob.OBLogServiceCluster{Spec: original.Spec}); got.Spec.ObjectStoreURL.BucketURL != original.Spec.ObjectStoreURL.BucketURL {
		t.Fatal("safe bucket URL changed")
	}
	p := validCreate()
	p.Spec.ObjectStoreURL.BucketURL = "s3://user:pass@bucket/path?host=minio&access_key=secret"
	p.Spec.Parameters = []api.Parameter{{Name: "access_key", Value: "secret"}}
	item := publicItem(&ob.OBLogServiceCluster{Spec: p.Spec})
	if item.Spec.ObjectStoreURL.BucketURL != "s3://bucket/path?host=minio" || item.Spec.Parameters[0].Value != "<redacted>" {
		t.Fatal("credential exposure")
	}
}
