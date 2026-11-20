package obcluster

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/obzone"
	"github.com/oceanbase/ob-operator/internal/telemetry"
)

func TestRollingStorageFinishesPriorTargetBeforePatchingAnotherZone(t *testing.T) {
	template := func(sc string) *apitypes.OBServerTemplate {
		vol := &apitypes.StorageSpec{StorageClass: sc}
		return &apitypes.OBServerTemplate{Storage: &apitypes.OceanbaseStorageSpec{DataStorage: vol, LogStorage: vol}}
	}
	cfg := obcfg.GetConfig()
	before, disabled := cfg.Time.ServerDeleteTimeoutSeconds, cfg.Telemetry.Disabled
	cfg.Time.ServerDeleteTimeoutSeconds, cfg.Telemetry.Disabled = 1, true
	t.Cleanup(func() { cfg.Time.ServerDeleteTimeoutSeconds, cfg.Telemetry.Disabled = before, disabled })
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	cluster := &v1alpha1.OBCluster{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "db"}}
	cluster.Spec.OBServerTemplate = template("third")
	var objs []client.Object
	for _, name := range []string{"a-idle", "z-inflight"} {
		zone := &v1alpha1.OBZone{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: name, UID: types.UID(name), Labels: map[string]string{oceanbaseconst.LabelRefOBCluster: cluster.Name}}}
		zone.Spec.Topology.Replica = 1
		zone.Spec.OBServerTemplate = template("old")
		zone.Status.Status = zonestatus.Running
		if name == "z-inflight" {
			zone.Spec.OBServerTemplate = template("second")
		}
		s := &v1alpha1.OBServer{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: name, UID: types.UID(name + "-server"), Labels: map[string]string{oceanbaseconst.LabelRefOBZone: name}, OwnerReferences: []metav1.OwnerReference{{Kind: "OBZone", Name: name, UID: zone.UID}}}}
		s.Status.Status = serverstatus.Running
		s.Spec.OBServerTemplate = template("old")
		objs = append(objs, zone, s)
	}
	patches := 0
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).WithInterceptorFuncs(interceptor.Funcs{Update: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
		if zone, ok := obj.(*v1alpha1.OBZone); ok {
			patches++
			s := &v1alpha1.OBServer{}
			if err := cl.Get(ctx, client.ObjectKeyFromObject(zone), s); err != nil {
				return err
			}
			s.Spec.OBServerTemplate = zone.Spec.OBServerTemplate.DeepCopy()
			if err := cl.Update(ctx, s); err != nil {
				return err
			}
		}
		return cl.Update(ctx, obj, opts...)
	}}).Build()
	ctx := context.Background()
	l := logr.Discard()
	m := &OBClusterManager{Ctx: ctx, OBCluster: cluster, Client: cl, Logger: &l, Recorder: telemetry.NewRecorder(ctx, record.NewFakeRecorder(20))}
	if err := RollingUpdateOBZones(m); err == nil || patches != 0 {
		t.Fatalf("pending second class: got %v, issued %d zone updates", err, patches)
	}
	s := &v1alpha1.OBServer{}
	if err := cl.Get(ctx, client.ObjectKey{Namespace: "test", Name: "z-inflight"}, s); err != nil {
		t.Fatal(err)
	}
	s.Spec.OBServerTemplate = template("second")
	if err := cl.Update(ctx, s); err != nil {
		t.Fatal(err)
	}
	if err := RollingUpdateOBZones(m); err != nil || patches != 2 {
		t.Fatalf("completed second class: got %v, issued %d zone updates", err, patches)
	}
	if !m.CanReplayTask(tRollingUpdateOBZones) {
		t.Fatal("persisted rolling stage must resume after a manager restart")
	}
}
