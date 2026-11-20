// Package obzone reconciles OceanBase zones and their servers.
package obzone

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	obconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
)

func (m *OBZoneManager) waitForBootstrapReplicas(status string) error {
	timeout := time.Duration(obcfg.GetConfig().Time.DefaultStateWaitTimeout) * time.Second
	return wait.PollUntilContextTimeout(m.Ctx, time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		zone := &v1alpha1.OBZone{}
		if err := m.Client.Get(ctx, client.ObjectKeyFromObject(m.OBZone), zone); err != nil {
			return false, err
		}
		if zone.UID != m.OBZone.UID || zone.DeletionTimestamp != nil {
			return false, nil
		}
		servers := &v1alpha1.OBServerList{}
		if err := m.Client.List(ctx, servers, client.InNamespace(zone.Namespace), client.MatchingLabels{obconst.LabelRefOBZone: zone.Name}); err != nil {
			return false, err
		}
		return resourceutils.BootstrapReplicasReady(zone, servers.Items, status), nil
	})
}
