// Package obcluster reconciles OceanBase clusters and their zones.
package obcluster

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	obconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/obzone"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
)

func (m *OBClusterManager) waitForBootstrapZones(target string) error {
	timeout := time.Duration(obcfg.GetConfig().Time.DefaultStateWaitTimeout) * time.Second
	return wait.PollUntilContextTimeout(m.Ctx, time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		return m.bootstrapZonesReady(ctx, target)
	})
}

func (m *OBClusterManager) bootstrapZonesReady(ctx context.Context, target string) (bool, error) {
	zones, err := m.readyBootstrapZones(ctx, target)
	return len(zones) > 0, err
}

// readyBootstrapZones returns the exact owned member snapshot checked by the
// bootstrap wait. An incomplete or stale topology has no usable snapshot.
func (m *OBClusterManager) readyBootstrapZones(ctx context.Context, target string) ([]v1alpha1.OBZone, error) {
	reader := m.APIReader
	if reader == nil {
		reader = m.Client
	}
	byCluster := client.MatchingLabels{obconst.LabelRefOBCluster: m.OBCluster.Name}
	inNamespace := client.InNamespace(m.OBCluster.Namespace)
	zones := &v1alpha1.OBZoneList{}
	if err := reader.List(ctx, zones, inNamespace, byCluster); err != nil {
		return nil, err
	}
	if len(zones.Items) == 0 || len(zones.Items) != len(m.OBCluster.Spec.Topology) {
		return nil, nil
	}
	expected := make(map[string]int, len(zones.Items))
	for _, topology := range m.OBCluster.Spec.Topology {
		expected[topology.Zone] = topology.Replica
	}
	servers := &v1alpha1.OBServerList{}
	if err := reader.List(ctx, servers, inNamespace, byCluster); err != nil {
		return nil, err
	}
	byZone := make(map[string][]v1alpha1.OBServer, len(zones.Items))
	for _, server := range servers.Items {
		name := server.Labels[obconst.LabelRefOBZone]
		byZone[name] = append(byZone[name], server)
	}
	addresses := make(map[string]bool, len(servers.Items))
	for i := range zones.Items {
		zone := &zones.Items[i]
		topology := zone.Spec.Topology
		if !resourceutils.IsLiveChild(zone, m.OBCluster, "OBCluster") || topology.Replica < 1 || expected[topology.Zone] != topology.Replica ||
			(zone.Status.Status != zonestatus.Running && (target != serverstatus.BootstrapReady || zone.Status.Status != zonestatus.BootstrapReady)) ||
			!resourceutils.BootstrapReplicasReady(zone, byZone[zone.Name], target) {
			return nil, nil
		}
		for _, replica := range zone.Status.OBServerStatus {
			address := replica.GetConnectAddr()
			if addresses[address] {
				return nil, nil
			}
			addresses[address] = true
		}
		delete(byZone, zone.Name)
		delete(expected, topology.Zone)
	}
	if len(expected) != 0 || len(byZone) != 0 {
		return nil, nil
	}
	return zones.DeepCopy().Items, nil
}
