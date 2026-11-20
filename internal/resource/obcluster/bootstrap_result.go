package obcluster

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/connector"
	dbconfig "github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/config"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

// A lost connection does not establish whether ALTER SYSTEM BOOTSTRAP took
// effect. Observe the system tenant and its bootstrap members without sending
// another bootstrap statement.
func (m *OBClusterManager) ObserveLostTask(name tasktypes.TaskName) tasktypes.TaskFunc {
	if name != tBootstrap || m.OBCluster.Status.Status != clusterstatus.New ||
		m.OBCluster.Spec.DeploymentMode == obconst.DeploymentModeSharedStorage {
		return nil
	}
	return func() tasktypes.TaskError {
		err := m.waitForBootstrapResult()
		return errors.Wrap(err, "could not verify the original database bootstrap")
	}
}

func (m *OBClusterManager) bootstrapResultReady(ctx context.Context) (bool, error) {
	reader := m.APIReader
	if reader == nil {
		reader = m.Client
	}
	current := &v1alpha1.OBCluster{}
	if err := reader.Get(ctx, client.ObjectKeyFromObject(m.OBCluster), current); err != nil {
		return false, err
	}
	if current.UID != m.OBCluster.UID || current.DeletionTimestamp != nil || current.Generation != m.OBCluster.Generation {
		return false, errors.New("cluster identity or desired state changed during bootstrap")
	}
	return m.verifyBootstrapDatabase(ctx, current)
}

func (m *OBClusterManager) verifyBootstrapDatabase(ctx context.Context, current *v1alpha1.OBCluster) (bool, error) {
	zones, err := m.readyBootstrapZones(ctx, serverstatus.BootstrapReady)
	if err != nil || len(zones) == 0 {
		return false, err
	}
	standalone := current.Annotations[obconst.AnnotationsMode] == obconst.ModeStandalone
	byZone := make(map[string]map[string]bool, len(zones))
	for _, zone := range zones {
		allowed := make(map[string]bool)
		for _, server := range zone.Status.OBServerStatus {
			addr := server.GetConnectAddr()
			if standalone {
				addr = "127.0.0.1"
			}
			allowed[addr] = true
		}
		byZone[zone.Spec.Topology.Zone] = allowed
	}
	// Only one Observer per zone executes bootstrap. Other ready replicas
	// may accept a connection before their system views become available.
	for _, zone := range zones {
		for _, server := range zone.Status.OBServerStatus {
			if err := ctx.Err(); err != nil {
				return false, err
			}
			manager, err := m.bootstrapClient(server.GetConnectAddr())
			if err != nil {
				continue
			}
			probeCtx, cancel := context.WithTimeout(ctx, dbconfig.DefaultSqlTimeout)
			ready := bootstrapDatabaseReady(probeCtx, manager, current, byZone)
			cancel()
			if ready {
				return true, nil
			}
		}
	}
	return false, nil
}

func (m *OBClusterManager) bootstrapClient(addr string) (*operation.OceanbaseOperationManager, error) {
	source := connector.NewOceanBaseDataSource(addr, obconst.SqlPort, obconst.RootUser, obconst.SysTenant, "", "")
	manager, err := operation.GetOceanbaseOperationManager(source)
	if err == nil {
		manager.Logger = m.Logger
	}
	return manager, err
}

func bootstrapSeeds(zones []v1alpha1.OBZone, connected string, standalone bool) []model.BootstrapServerInfo {
	seeds := make([]model.BootstrapServerInfo, 0, len(zones))
	for _, zone := range zones {
		addr := zone.Status.OBServerStatus[0].GetConnectAddr()
		if standalone {
			addr = "127.0.0.1"
		} else {
			for _, server := range zone.Status.OBServerStatus {
				if server.GetConnectAddr() == connected {
					addr = connected
					break
				}
			}
		}
		seeds = append(seeds, model.BootstrapServerInfo{Zone: zone.Spec.Topology.Zone, Server: &model.ServerInfo{Ip: addr, Port: obconst.RpcPort}})
	}
	return seeds
}

func bootstrapDatabaseReady(ctx context.Context, manager *operation.OceanbaseOperationManager, cluster *v1alpha1.OBCluster, byZone map[string]map[string]bool) bool {
	var identity int
	// A fresh connection intentionally has no default database or password.
	// Confirm the identity and system tenant before trusting server records.
	const identitySQL = `SELECT COUNT(DISTINCT name) FROM oceanbase.__all_virtual_sys_parameter_stat
WHERE ((name='cluster' AND value=?) OR (name='cluster_id' AND value=?))
AND EXISTS (SELECT 1 FROM oceanbase.__all_tenant WHERE tenant_id=1 AND tenant_name='sys')`
	if err := manager.QueryRow(ctx, &identity, identitySQL, cluster.Spec.ClusterName, strconv.FormatInt(cluster.Spec.ClusterId, 10)); err != nil || identity != 2 {
		return false
	}
	var members []model.OBServer
	err := manager.QueryList(ctx, &members, "SELECT zone, svr_ip, svr_port, LOWER(status) AS status, start_service_time FROM oceanbase.__all_server")
	return err == nil && bootstrapMembersReady(byZone, members)
}

func bootstrapMembersReady(byZone map[string]map[string]bool, members []model.OBServer) bool {
	if len(byZone) == 0 || len(members) != len(byZone) {
		return false
	}
	seen := make(map[string]bool, len(members))
	for _, member := range members {
		if seen[member.Zone] || !byZone[member.Zone][member.Ip] || member.Port != obconst.RpcPort ||
			member.Status != "active" || member.StartServiceTime <= 0 {
			return false
		}
		seen[member.Zone] = true
	}
	return true
}
