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
	"testing"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/obzone"
	"github.com/oceanbase/ob-operator/internal/telemetry"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
)

func TestMonitorRolloutUsesEffectiveZoneTemplate(t *testing.T) {
	original := &apitypes.MonitorTemplate{Image: "agent:v1", Resource: &apitypes.ResourceSpec{
		Cpu: resource.MustParse("500m"), Memory: resource.MustParse("512Mi"),
	}}
	updated := original.DeepCopy()
	updated.Image = "agent:v2"
	m := &OBClusterManager{OBCluster: &v1alpha1.OBCluster{}}
	m.OBCluster.Spec.MonitorTemplate = updated
	zone := &v1alpha1.OBZone{}
	zone.Spec.Topology.Zone = "z1"
	zone.Spec.MonitorTemplate = original
	if !m.checkIfMonitorMutated(zone) {
		t.Fatal("updated default image must propagate")
	}
	m.changeZonesWhenModifyingServerTemplate(zone)
	if zone.Spec.MonitorTemplate.Image != updated.Image || m.checkIfMonitorMutated(zone) {
		t.Fatal("zone did not converge on updated image")
	}
	m.OBCluster.Spec.Topology = []apitypes.OBZoneTopology{{Zone: "z1", MonitorTemplate: original}}
	if !m.checkIfMonitorMutated(zone) {
		t.Fatal("new override must propagate")
	}
	m.changeZonesWhenModifyingServerTemplate(zone)
	if !equality.Semantic.DeepEqual(zone.Spec.Topology.MonitorTemplate, original) {
		t.Fatal("zone override not saved")
	}

	m.OBCluster.Spec.MonitorTemplate = nil
	m.changeZonesWhenModifyingServerTemplate(zone)
	if zone.Spec.MonitorTemplate != nil || !equality.Semantic.DeepEqual(zone.Spec.Topology.MonitorTemplate, original) || m.checkIfMonitorMutated(zone) {
		t.Fatal("overridden zone must not enter a rollout on a default-only change")
	}
	m.OBCluster.Spec.MonitorTemplate = original.DeepCopy()
	m.OBCluster.Spec.Topology[0].MonitorTemplate = nil
	if !m.checkIfMonitorMutated(zone) {
		t.Fatal("removing an override equal to the new default must synchronize its provenance")
	}
	m.changeZonesWhenModifyingServerTemplate(zone)
	if zone.Spec.Topology.MonitorTemplate != nil {
		t.Fatal("obsolete equal override was retained")
	}
	m.OBCluster.Spec.MonitorTemplate = nil
	m.changeZonesWhenModifyingServerTemplate(zone)
	if zone.Spec.MonitorTemplate != nil || zone.Spec.Topology.MonitorTemplate != nil {
		t.Fatal("removing override must restore disabled default")
	}
}

func TestMonitorRolloutResumesDispatchedZoneBeforeRollingAnother(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	cluster := &v1alpha1.OBCluster{ObjectMeta: metav1.ObjectMeta{Name: "db", Namespace: "ns", UID: "cluster"}}
	cluster.Spec.MonitorTemplate = &apitypes.MonitorTemplate{Image: "v2"}
	var objects []client.Object
	for _, name := range []string{"a-next", "z-dispatched"} {
		zone := &v1alpha1.OBZone{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "ns", UID: types.UID(name), Labels: map[string]string{oceanbaseconst.LabelRefOBCluster: "db"}, OwnerReferences: []metav1.OwnerReference{{Kind: "OBCluster", Name: "db", UID: "cluster"}}}}
		zone.Spec.Topology = apitypes.OBZoneTopology{Zone: name, Replica: 1}
		zone.Spec.MonitorTemplate = &apitypes.MonitorTemplate{Image: "v1"}
		zone.Status.Status = zonestatus.Running
		server := &v1alpha1.OBServer{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "ns", UID: types.UID(name + "-server"), Labels: map[string]string{oceanbaseconst.LabelRefOBZone: name}, OwnerReferences: []metav1.OwnerReference{{Kind: "OBZone", Name: name, UID: zone.UID}}}}
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "ns", UID: types.UID(name + "-pod"), OwnerReferences: []metav1.OwnerReference{{Kind: "OBServer", Name: name, UID: server.UID}}}}
		pod.Spec.Containers = []corev1.Container{{Name: "obagent"}}
		pod.Status.Phase = corev1.PodRunning
		pod.Status.Conditions = []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}
		if name == "z-dispatched" {
			zone.Spec.MonitorTemplate = cluster.Spec.MonitorTemplate.DeepCopy()
			server.Annotations = map[string]string{oceanbaseconst.AnnotationsMonitorRevision: "in-flight"}
		}
		server.Spec.MonitorTemplate = zone.Spec.MonitorTemplate.DeepCopy()
		server.Status.Status, server.Status.Ready = serverstatus.Running, true
		objects = append(objects, zone, server, pod)
	}
	patches := 0
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).WithInterceptorFuncs(interceptor.Funcs{
		Update: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
			if zone, ok := obj.(*v1alpha1.OBZone); ok {
				patches++
				key := client.ObjectKeyFromObject(zone)
				server, pod := &v1alpha1.OBServer{}, &corev1.Pod{}
				if err := cl.Get(ctx, key, server); err != nil {
					return err
				}
				if err := cl.Get(ctx, key, pod); err != nil {
					return err
				}
				server.Spec.MonitorTemplate = zone.Spec.MonitorTemplate.DeepCopy()
				server.Annotations = map[string]string{oceanbaseconst.AnnotationsMonitorRevision: "complete"}
				pod.Annotations = map[string]string{oceanbaseconst.AnnotationsMonitorRevision: "complete"}
				if err := cl.Update(ctx, server); err != nil {
					return err
				}
				if err := cl.Update(ctx, pod); err != nil {
					return err
				}
			}
			return cl.Update(ctx, obj, opts...)
		},
	}).Build()
	cfg := obcfg.GetConfig()
	disabled := cfg.Telemetry.Disabled
	cfg.Telemetry.Disabled = true
	t.Cleanup(func() { cfg.Telemetry.Disabled = disabled })
	logger := logr.Discard()
	m := &OBClusterManager{Ctx: ctx, OBCluster: cluster, Client: cl, Logger: &logger, Recorder: telemetry.NewRecorder(ctx, record.NewFakeRecorder(10))}
	if !m.CanReplayTask(tModifyServerTemplate) || !m.CanReplayTask(tWaitOBZoneRunning) {
		t.Fatal("lost cluster rollout must resume and verify its last zone")
	}
	roll := m.rollingUpdateZones(m.changeZonesWhenModifyingServerTemplate, zonestatus.ModifyServerTemplate, zonestatus.Running, 1)
	if err := roll(); err == nil || patches != 0 {
		t.Fatalf("unfinished pod revision: err=%v zone updates=%d", err, patches)
	}
	pod := &corev1.Pod{}
	if err := cl.Get(ctx, client.ObjectKey{Namespace: "ns", Name: "z-dispatched"}, pod); err != nil {
		t.Fatal(err)
	}
	pod.Annotations = map[string]string{oceanbaseconst.AnnotationsMonitorRevision: "in-flight"}
	if err := cl.Update(ctx, pod); err != nil {
		t.Fatal(err)
	}
	if err := roll(); err != nil || patches != 1 {
		t.Fatalf("completed first rollout: err=%v zone updates=%d", err, patches)
	}
}
