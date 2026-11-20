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
	patches, staleReads, prematureWrites := 0, 0, 0
	var staleZone *v1alpha1.OBZone
	var staleChild *v1alpha1.OBServer
	cacheBehind := false
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).WithInterceptorFuncs(interceptor.Funcs{
		Get: func(ctx context.Context, cl client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			if target, ok := obj.(*v1alpha1.OBZone); ok && staleZone != nil && key.Name == staleZone.Name {
				cacheBehind = staleReads < 2
				if cacheBehind {
					staleReads++
					staleZone.DeepCopyInto(target)
					return nil
				}
			}
			return cl.Get(ctx, key, obj, opts...)
		},
		List: func(ctx context.Context, cl client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
			if err := cl.List(ctx, list, opts...); err != nil {
				return err
			}
			if servers, ok := list.(*v1alpha1.OBServerList); ok && cacheBehind {
				for i := range servers.Items {
					if servers.Items[i].Name == staleChild.Name {
						staleChild.DeepCopyInto(&servers.Items[i])
						break
					}
				}
			}
			return nil
		},
		Update: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
			if zone, ok := obj.(*v1alpha1.OBZone); ok {
				patches++
				s := &v1alpha1.OBServer{}
				if err := cl.Get(ctx, client.ObjectKeyFromObject(zone), s); err != nil {
					return err
				}
				if zone.Name == "a-idle" {
					staleZone = &v1alpha1.OBZone{}
					if err := cl.Get(ctx, client.ObjectKeyFromObject(zone), staleZone); err != nil {
						return err
					}
					staleChild = s.DeepCopy()
				} else if staleZone != nil && staleReads < 2 {
					prematureWrites++
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
	cfg.Time.ServerDeleteTimeoutSeconds = 5
	if err := RollingUpdateOBZones(m); err != nil || patches != 2 {
		t.Fatalf("completed second class: got %v, issued %d zone updates", err, patches)
	}
	if staleReads != 2 || prematureWrites != 0 {
		t.Fatalf("consumed %d stale cache snapshots and updated %d later zones prematurely", staleReads, prematureWrites)
	}
	if !m.CanReplayTask(tRollingUpdateOBZones) {
		t.Fatal("persisted rolling stage must resume after a manager restart")
	}
}
