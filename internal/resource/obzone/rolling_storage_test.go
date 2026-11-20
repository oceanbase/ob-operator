package obzone

import (
	"context"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	"github.com/oceanbase/ob-operator/internal/telemetry"
)

func TestRollingStorageWaitsForPreviousObserverDeletion(t *testing.T) {
	cfg := obcfg.GetConfig()
	oldTimeout, oldDisabled := cfg.Time.ServerDeleteTimeoutSeconds, cfg.Telemetry.Disabled
	cfg.Time.ServerDeleteTimeoutSeconds, cfg.Telemetry.Disabled = 1, true
	t.Cleanup(func() { cfg.Time.ServerDeleteTimeoutSeconds, cfg.Telemetry.Disabled = oldTimeout, oldDisabled })

	template := func(sc string) *apitypes.OBServerTemplate {
		storage := &apitypes.StorageSpec{StorageClass: sc}
		return &apitypes.OBServerTemplate{Storage: &apitypes.OceanbaseStorageSpec{DataStorage: storage, LogStorage: storage}}
	}
	zone := &v1alpha1.OBZone{ObjectMeta: metav1.ObjectMeta{Name: "zone1", Namespace: "test"},
		Spec: v1alpha1.OBZoneSpec{Topology: apitypes.OBZoneTopology{Zone: "zone1", Replica: 2}, OBServerTemplate: template("new")}}
	old := func(name string) *v1alpha1.OBServer {
		return &v1alpha1.OBServer{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: zone.Namespace,
			Labels: map[string]string{oceanbaseconst.LabelRefOBZone: zone.Name}, Finalizers: []string{"test.example/hold"}},
			Spec: v1alpha1.OBServerSpec{OBServerTemplate: template("old")}}
	}
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	created := 0
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(old("old-1"), old("old-2")).
		WithInterceptorFuncs(interceptor.Funcs{Create: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
			obj.(*v1alpha1.OBServer).Status.Status = serverstatus.Running
			created++
			return cl.Create(ctx, obj, opts...)
		}}).Build()
	ctx := context.Background()
	logger := logr.Discard()
	m := &OBZoneManager{Ctx: ctx, OBZone: zone, Client: cl, Logger: &logger, Recorder: telemetry.NewRecorder(ctx, record.NewFakeRecorder(10))}
	for attempt := 1; attempt <= 2; attempt++ {
		err := RollingReplaceOBServers(m)
		if err == nil || !strings.Contains(err.Error(), "old-1") || created != 1 {
			t.Fatalf("attempt %d, deletion held by finalizer: got %v, created %d replacements; want error identifying old-1 and one replacement", attempt, err, created)
		}
	}
	// Once both old finalizers release, the same task should keep the first
	// successful replacement and migrate the second server exactly once.
	for _, name := range []string{"old-1", "old-2"} {
		server := &v1alpha1.OBServer{}
		if err := cl.Get(ctx, client.ObjectKey{Namespace: zone.Namespace, Name: name}, server); err != nil {
			t.Fatal(err)
		}
		server.Finalizers = nil
		if err := cl.Update(ctx, server); err != nil {
			t.Fatal(err)
		}
	}
	if err := RollingReplaceOBServers(m); err != nil || created != 2 {
		t.Fatalf("after releasing the finalizers: got %v and %d replacements, want success and exactly two", err, created)
	}
	current, err := m.listOBServers()
	if err != nil || len(current.Items) != 2 {
		t.Fatalf("final servers: got %+v, error %v", current, err)
	}
	for _, server := range current.Items {
		if m.checkIfStorageClassChanged(&server) {
			t.Fatalf("old storage still present on %s", server.Name)
		}
	}
}
