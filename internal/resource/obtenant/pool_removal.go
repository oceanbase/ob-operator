package obtenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

func (m *OBTenantManager) poolRemovalKey() string {
	// Only administrators of the owning cluster may authorize cleanup after detachment.
	// A distinct key keeps multiple tenants' removals independent.
	return "oceanbase.oceanbase.com/pool-removal-" + string(m.OBTenant.UID)
}

// Capture the physical identities while the database still identifies the owner.
// A detached pool no longer contains its tenant ID.
type poolRemovalIntent struct {
	KeepResources bool      `json:"keepResources,omitempty"`
	TenantUID     types.UID `json:"tenantUID"`
	ClusterUID    types.UID `json:"clusterUID"`
	TenantID      int64     `json:"tenantID"`
	TenantName    string    `json:"tenantName"`
	Zone          string    `json:"zone"`
	PoolID        int64     `json:"poolID"`
	UnitConfigID  int64     `json:"unitConfigID"`
}

type poolRemovalDatabase interface {
	GetTenantByName(context.Context, string) (*model.OBTenant, error)
	GetPoolList(context.Context) ([]model.Pool, error)
	GetUnitConfigV4List(context.Context) ([]model.UnitConfigV4, error)
	SetTenant(context.Context, model.TenantSQLParam) error
	CheckRsJobExistByTenantID(context.Context, int) (bool, error)
	DeletePool(context.Context, string) error
	DeleteUnitConfig(context.Context, string) error
}

func (m *OBTenantManager) currentRemovalTenant() (*v1alpha1.OBTenant, error) {
	var live v1alpha1.OBTenant
	if err := m.Client.Get(m.Ctx, client.ObjectKeyFromObject(m.OBTenant), &live); err != nil {
		return nil, err
	}
	if live.UID == "" || live.UID != m.OBTenant.UID || live.Generation != m.OBTenant.Generation || !live.DeletionTimestamp.IsZero() {
		return nil, errors.New("tenant changed before pool removal")
	}
	return &live, nil
}

func (m *OBTenantManager) reconcilePoolRemoval() error {
	live, err := m.currentRemovalTenant()
	if err != nil {
		return err
	}
	cluster, err := m.getOBCluster()
	if err != nil {
		return err
	}
	db, err := resourceutils.GetSysOperationClient(m.Client, m.Logger, cluster)
	if err != nil {
		return err
	}
	return m.removePool(live, cluster, db)
}

func (i poolRemovalIntent) appliesTo(live *v1alpha1.OBTenant, cluster *v1alpha1.OBCluster) bool {
	return i.TenantUID == live.UID && i.ClusterUID == cluster.UID && i.TenantName == live.Spec.TenantName &&
		i.Zone != "" && i.TenantID > 0 && i.PoolID > 0 && i.UnitConfigID > 0
}

func (m *OBTenantManager) removePool(live *v1alpha1.OBTenant, cluster *v1alpha1.OBCluster, db poolRemovalDatabase) error {
	var intent poolRemovalIntent
	raw := cluster.Annotations[m.poolRemovalKey()]
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &intent); err != nil {
			return fmt.Errorf("invalid pending pool removal: %w", err)
		}
		if !intent.appliesTo(live, cluster) {
			return m.writePoolRemoval(cluster, "")
		}
		for _, desired := range live.Spec.Pools {
			if desired.Zone == intent.Zone {
				return m.writePoolRemoval(cluster, "")
			}
		}
	}
	tenant, err := db.GetTenantByName(m.Ctx, live.Spec.TenantName)
	if err != nil {
		return err
	}
	pools, err := db.GetPoolList(m.Ctx)
	if err != nil {
		return err
	}
	if raw == "" {
		for _, candidate := range m.getPoolsForDelete() {
			zone := candidate.ZoneList
			plan := planPoolRemoval(pools, tenant.TenantID, m.generatePoolName(zone), zone)
			if plan.target.ResourcePoolID == 0 {
				continue
			}
			intent = poolRemovalIntent{
				TenantUID: live.UID, ClusterUID: cluster.UID, TenantID: tenant.TenantID,
				TenantName: tenant.TenantName, Zone: zone, PoolID: plan.target.ResourcePoolID, UnitConfigID: plan.target.UnitConfigID,
			}
			if live.Status.Creation != nil {
				ids := live.Status.Creation.Resources[zone]
				intent.KeepResources = ids.Borrowed || ids.PoolID != intent.PoolID || ids.UnitConfigID != intent.UnitConfigID
			}
			data, err := json.Marshal(intent)
			if err != nil {
				return err
			}
			if err := m.writePoolRemoval(cluster, string(data)); err != nil {
				return err
			}
			break
		}
		if intent.PoolID == 0 {
			return nil
		}
	} else if intent.TenantID != tenant.TenantID {
		return m.writePoolRemoval(cluster, "")
	}
	done, err := m.advancePoolRemoval(db, tenant, pools, intent)
	if err != nil || !done {
		return err
	}
	return m.writePoolRemoval(cluster, "")
}

// The normal tenant deletion loop only visits zones still present in spec.
// If deletion crossed an in-flight pool removal, finish its proven detached
// objects before allowing the tenant finalizer to disappear.
func (m *OBTenantManager) finalizePoolRemoval(cluster *v1alpha1.OBCluster) error {
	var intent poolRemovalIntent
	if err := json.Unmarshal([]byte(cluster.Annotations[m.poolRemovalKey()]), &intent); err != nil {
		return fmt.Errorf("invalid pending pool removal: %w", err)
	}
	if !intent.appliesTo(m.OBTenant, cluster) {
		return m.writePoolRemoval(cluster, "")
	}
	db, err := resourceutils.GetSysOperationClient(m.Client, m.Logger, cluster)
	if err != nil {
		return err
	}
	return m.finishDetachedPool(db, cluster, intent)
}

func (m *OBTenantManager) finishDetachedPool(db poolRemovalDatabase, cluster *v1alpha1.OBCluster, intent poolRemovalIntent) error {
	// There are at most two remaining stages after deleting the tenant:
	// release its detached pool, then remove its unused unit configuration.
	for range 2 {
		pools, err := db.GetPoolList(m.Ctx)
		if err != nil {
			return err
		}
		done, err := m.advancePoolRemoval(db, nil, pools, intent)
		if err != nil {
			return err
		}
		if done {
			return m.writePoolRemoval(cluster, "")
		}
	}
	return fmt.Errorf("waiting for detached pool %s cleanup", m.generatePoolName(intent.Zone))
}

func (m *OBTenantManager) writePoolRemoval(cluster *v1alpha1.OBCluster, value string) error {
	base := cluster.DeepCopy()
	key := m.poolRemovalKey()
	if value == "" {
		delete(cluster.Annotations, key)
	} else {
		if cluster.Annotations == nil {
			cluster.Annotations = make(map[string]string)
		}
		cluster.Annotations[key] = value
	}
	return m.Client.Patch(m.Ctx, cluster, client.MergeFromWithOptions(base, client.MergeFromWithOptimisticLock{}))
}

// Advance one physical stage from the database's current state. A fresh worker
// can repeat this after a lost response, including after DROP of the pool itself.
func (m *OBTenantManager) advancePoolRemoval(db poolRemovalDatabase, tenant *model.OBTenant, pools []model.Pool, intent poolRemovalIntent) (bool, error) {
	poolName := m.generatePoolName(intent.Zone)
	for _, pool := range pools {
		if pool.Name != poolName && pool.ResourcePoolID != intent.PoolID {
			continue
		}
		if pool.Name != poolName || pool.ResourcePoolID != intent.PoolID || pool.ZoneList != intent.Zone || pool.UnitConfigID != intent.UnitConfigID {
			return true, nil
		}
		if !pool.TenantID.Valid {
			if intent.KeepResources {
				return true, nil
			}
			return false, db.DeletePool(m.Ctx, poolName)
		}
		if pool.TenantID.Int64 != intent.TenantID {
			return true, nil
		}
		if tenant == nil {
			return false, fmt.Errorf("waiting for pool %s to detach from deleted tenant %d", poolName, intent.TenantID)
		}
		plan := planPoolRemoval(pools, intent.TenantID, poolName, intent.Zone)
		return false, m.detachOwnedPool(db, tenant, plan, intent)
	}

	if intent.KeepResources {
		return true, nil
	}
	configs, err := db.GetUnitConfigV4List(m.Ctx)
	if err != nil {
		return false, err
	}
	unitName := m.generateUnitName(intent.Zone)
	if removableUnitConfig(intent.UnitConfigID, unitName, configs, pools) {
		if err := db.DeleteUnitConfig(m.Ctx, unitName); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (m *OBTenantManager) detachOwnedPool(db poolRemovalDatabase, tenant *model.OBTenant, plan poolRemoval, intent poolRemovalIntent) error {
	var locality []string
	removed := false
	for _, entry := range strings.Split(tenant.Locality, ",") {
		entry = strings.TrimSpace(entry)
		if !plan.sharedZone && strings.HasSuffix(entry, "@"+intent.Zone) {
			removed = true
		} else {
			locality = append(locality, entry)
		}
	}
	if removed {
		if err := db.SetTenant(m.Ctx, model.TenantSQLParam{TenantName: tenant.TenantName, Locality: strings.Join(locality, ",")}); err != nil {
			return err
		}
	}
	if err := resourceutils.CheckJobWithTimeout(func() (bool, error) {
		active, err := db.CheckRsJobExistByTenantID(m.Ctx, int(tenant.TenantID))
		return !active, err
	}); err != nil {
		return fmt.Errorf("wait for locality change before detaching pool: %w", err)
	}
	if _, err := m.currentRemovalTenant(); err != nil {
		return err
	}
	pools, err := db.GetPoolList(m.Ctx)
	if err != nil {
		return err
	}
	current := planPoolRemoval(pools, intent.TenantID, m.generatePoolName(intent.Zone), intent.Zone)
	if current.target.ResourcePoolID != intent.PoolID || current.target.UnitConfigID != intent.UnitConfigID {
		return fmt.Errorf("resource pool changed while removing it from tenant %s", tenant.TenantName)
	}
	if len(current.remaining) == 0 {
		return fmt.Errorf("cannot detach the last pool from tenant %s", tenant.TenantName)
	}
	return db.SetTenant(m.Ctx, model.TenantSQLParam{
		TenantName: tenant.TenantName, PoolList: current.remaining, PrimaryZone: m.generateSpecPrimaryZone(m.OBTenant.Spec.Pools),
	})
}
