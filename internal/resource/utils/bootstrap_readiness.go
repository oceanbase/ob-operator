package utils

import (
	"slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
)

// IsLiveChild excludes terminating resources and children of a previous owner with the same name.
func IsLiveChild(obj metav1.Object, parent metav1.Object, kind string) bool {
	return parent.GetUID() != "" && obj.GetDeletionTimestamp() == nil && slices.ContainsFunc(obj.GetOwnerReferences(), func(ref metav1.OwnerReference) bool {
		return ref.Kind == kind && ref.Name == parent.GetName() && ref.UID == parent.GetUID()
	})
}

// BootstrapReplicasReady requires a complete set of live children and the same
// addresses in zone status. SQL bootstrap consumes this published address list.
func BootstrapReplicasReady(zone *v1alpha1.OBZone, servers []v1alpha1.OBServer, target string) bool {
	expected := zone.Spec.Topology.Replica
	if expected < 1 || len(servers) != expected || len(zone.Status.OBServerStatus) != expected {
		return false
	}
	matches := func(status string) bool {
		return status == target || (target == serverstatus.BootstrapReady && status == serverstatus.Running)
	}
	addresses := make(map[string]types.UID, expected)
	for i := range servers {
		server := &servers[i]
		if !IsLiveChild(server, zone, "OBZone") || !matches(server.Status.Status) {
			return false
		}
		address := server.Status.GetConnectAddr()
		if address == "" || server.UID == "" || addresses[address] != "" {
			return false
		}
		addresses[address] = server.UID
	}
	for _, published := range zone.Status.OBServerStatus {
		addr := published.GetConnectAddr()
		if !matches(published.Status) || addresses[addr] == "" {
			return false
		}
		delete(addresses, addr)
	}
	return true
}
