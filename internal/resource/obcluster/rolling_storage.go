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
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/obzone"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
)

// A zone's spec may already be updated while its running Pods still use the
// old class. Verify every child before permitting another zone to allocate a spare.
func (m *OBClusterManager) waitForZoneStorage(zone *v1alpha1.OBZone, timeout time.Duration) error {
	expected := resourceutils.EffectiveZoneTemplate(zone)
	err := wait.PollUntilContextTimeout(m.Ctx, time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		current := &v1alpha1.OBZone{}
		if err := m.Client.Get(ctx, client.ObjectKeyFromObject(zone), current); err != nil {
			return false, err
		}
		if current.UID == "" || current.UID != zone.UID {
			return false, errors.Errorf("zone %s identity changed", zone.Name)
		}
		if current.Status.Status != zonestatus.Running || current.Generation < zone.Generation ||
			current.Spec.Topology.Replica != zone.Spec.Topology.Replica ||
			!resourceutils.SameObserverStorageClasses(current.Spec.OBServerTemplate, zone.Spec.OBServerTemplate, zone.Annotations) ||
			!resourceutils.SameObserverStorageClasses(resourceutils.EffectiveZoneTemplate(current), expected, zone.Annotations) {
			return false, nil
		}
		servers := &v1alpha1.OBServerList{}
		if err := m.Client.List(ctx, servers, client.InNamespace(zone.Namespace), client.MatchingLabels{oceanbaseconst.LabelRefOBZone: zone.Name}); err != nil {
			return false, err
		}
		if len(servers.Items) != zone.Spec.Topology.Replica {
			return false, nil
		}
		for _, s := range servers.Items {
			owned := false
			for _, ref := range s.OwnerReferences {
				if ref.Kind == "OBZone" && ref.Name == zone.Name && ref.UID == zone.UID {
					owned = true
					break
				}
			}
			if !owned || s.DeletionTimestamp != nil || s.Status.Status != serverstatus.Running || !resourceutils.SameObserverStorageClasses(s.Spec.OBServerTemplate, expected, zone.Annotations) {
				return false, nil
			}
		}
		return true, nil
	})
	return errors.Wrapf(err, "wait for storage classes on observers in zone %s", zone.Name)
}
