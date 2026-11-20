package obcluster

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/obzone"
)

func TestBootstrapWaitRequiresOwnedReadyReplicas(t *testing.T) {
	const ns = "db"
	cluster := &v1alpha1.OBCluster{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "cluster", UID: "cluster-uid"}}
	topology := apitypes.OBZoneTopology{Zone: "zone1", Replica: 1}
	cluster.Spec.Topology = []apitypes.OBZoneTopology{topology}
	zone := &v1alpha1.OBZone{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "cluster-1-zone1", UID: "zone-uid", Labels: map[string]string{obconst.LabelRefOBCluster: cluster.Name}, OwnerReferences: []metav1.OwnerReference{{Kind: "OBCluster", Name: cluster.Name, UID: cluster.UID}}}}
	zone.Spec.Topology = topology
	// A controller prior to the fix could already have published Running here.
	zone.Status.Status = zonestatus.Running
	zone.Status.OBServerStatus = []apitypes.OBServerReplicaStatus{{Status: serverstatus.BootstrapReady, Server: "192.0.2.1"}}
	server := &v1alpha1.OBServer{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "observer", UID: "observer-uid", Labels: map[string]string{obconst.LabelRefOBCluster: cluster.Name, obconst.LabelRefOBZone: zone.Name}, OwnerReferences: []metav1.OwnerReference{{Kind: "OBZone", Name: zone.Name, UID: zone.UID}}}}
	server.Status.Status = serverstatus.BootstrapReady
	server.Status.PodIp = "192.0.2.1"
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	target := serverstatus.BootstrapReady
	check := func(want bool, children ...client.Object) {
		t.Helper()
		m := OBClusterManager{OBCluster: cluster, APIReader: fake.NewClientBuilder().WithScheme(scheme).WithObjects(children...).Build()}
		got, err := m.bootstrapZonesReady(context.Background(), target)
		if err != nil || got != want {
			t.Fatalf("bootstrap ready=%v, want %v, err=%v", got, want, err)
		}
	}
	check(false)
	check(false, zone)
	check(true, zone, server)
	foreign := server.DeepCopy()
	foreign.OwnerReferences[0].UID = "earlier-zone"
	check(false, zone, foreign)
	foreignZone := zone.DeepCopy()
	foreignZone.OwnerReferences[0].UID = "earlier-cluster"
	check(false, foreignZone, server)
	partial := zone.DeepCopy()
	partial.Status.OBServerStatus = nil
	check(false, partial, server)
	stale := zone.DeepCopy()
	stale.Status.OBServerStatus[0].Server = "192.0.2.2"
	check(false, stale, server)
	failed := server.DeepCopy()
	failed.Status.Status = serverstatus.Unrecoverable
	check(false, zone, failed)
	cluster.Spec.Topology[0].Replica = 2
	check(false, zone, server)
	cluster.Spec.Topology[0].Replica = 1
	secondZone := zone.DeepCopy()
	secondZone.Name = "cluster-1-zone2"
	secondZone.UID = "zone2-uid"
	secondZone.Spec.Topology.Zone = "zone2"
	secondServer := server.DeepCopy()
	secondServer.Name, secondServer.UID = "second", "second-uid"
	secondServer.OwnerReferences[0].Name = secondZone.Name
	secondServer.OwnerReferences[0].UID = secondZone.UID
	secondServer.Labels[obconst.LabelRefOBZone] = secondZone.Name
	cluster.Spec.Topology = append(cluster.Spec.Topology, secondZone.Spec.Topology)
	check(false, zone, server, secondZone, secondServer) // the same address cannot identify two servers
	secondServer.Status.PodIp = "192.0.2.3"
	secondZone.Status.OBServerStatus[0].Server = secondServer.Status.PodIp
	check(true, zone, server, secondZone, secondServer)
	secondZone.Spec.Topology.Zone = zone.Spec.Topology.Zone
	check(false, zone, server, secondZone, secondServer)
	cluster.Spec.Topology = cluster.Spec.Topology[:1]
	target = serverstatus.Running
	check(false, zone, server)
	server.Status.Status = serverstatus.Running
	zone.Status.OBServerStatus[0].Status = serverstatus.Running
	check(true, zone, server)
}
