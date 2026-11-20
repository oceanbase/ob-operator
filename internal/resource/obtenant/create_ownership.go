package obtenant

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/const/status/tenantstatus"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
)

var errAllocationAssigned = errors.New("owned pools already assigned to the previously created tenant")

var errForeignCluster = errors.New("original tenant cluster is gone or replaced")

var errStaleTenant = errors.New("tenant changed since the task was dispatched")

var errCreationBindingChanged = errors.New("ordinary tenant creation no longer addresses its original database or owner")

func (m *OBTenantManager) reader() client.Reader {
	if m.APIReader != nil {
		return m.APIReader
	}
	return m.Client
}

func checkTenantClusterOwner(cluster *v1alpha1.OBCluster, live *v1alpha1.OBTenant) error {
	var owner *metav1.OwnerReference
	for i := range live.OwnerReferences {
		ref := &live.OwnerReferences[i]
		if ref.APIVersion != v1alpha1.GroupVersion.String() || ref.Kind != "OBCluster" {
			continue
		}
		if owner != nil {
			return errCreationBindingChanged
		}
		owner = ref
	}
	if owner == nil || owner.Name != cluster.Name || owner.UID == "" || cluster.UID == "" {
		return errCreationBindingChanged
	}
	if owner.UID != cluster.UID {
		return errForeignCluster
	}
	return nil
}

func (m *OBTenantManager) readCreationObjects(ctx context.Context) (*v1alpha1.OBTenant, *v1alpha1.OBCluster, error) {
	var live v1alpha1.OBTenant
	if err := m.reader().Get(ctx, client.ObjectKeyFromObject(m.OBTenant), &live); err != nil {
		return nil, nil, err
	}
	if live.UID == "" || live.UID != m.OBTenant.UID || live.Generation != m.OBTenant.Generation || !live.DeletionTimestamp.Equal(m.OBTenant.DeletionTimestamp) {
		return nil, nil, errStaleTenant
	}
	var cluster v1alpha1.OBCluster
	if err := m.reader().Get(ctx, client.ObjectKey{Namespace: live.Namespace, Name: live.Spec.ClusterName}, &cluster); err != nil {
		return nil, nil, err
	}
	if !cluster.DeletionTimestamp.IsZero() || live.Status.Creation != nil && live.Status.Creation.ClusterUID != cluster.UID {
		return nil, nil, errForeignCluster
	}
	if err := checkTenantClusterOwner(&cluster, &live); err != nil {
		return nil, nil, err
	}
	return &live, &cluster, nil
}

// Bind before the first named-object check, including checks that can reject the
// creation. Legacy running, restored and standby tenants retain their existing flow.
func shouldBindOrdinaryCreation(t *v1alpha1.OBTenant) bool {
	if t.Status.Creation != nil || !t.DeletionTimestamp.IsZero() || (t.Spec.Source != nil && (t.Spec.Source.Tenant != nil || t.Spec.Source.Restore != nil)) {
		return false
	}
	// The coordinator first initializes a structurally valid status. Bind before
	// it schedules the first database task. A pre-upgrade workflow may already
	// have provisioned some or all of its unmarked allocations.
	if t.Status.Status != tenantstatus.CreatingTenant {
		return false
	}
	op := t.Status.OperationContext
	return op == nil || op.Task == "" && op.Idx == 0 || op.Task == tCheckTenant || op.Task == tCheckPoolAndConfig
}

func (m *OBTenantManager) BindOrdinaryCreation() (bool, error) {
	t := m.OBTenant
	if !shouldBindOrdinaryCreation(t) {
		return false, nil
	}
	live, cluster, err := m.readCreationObjects(m.Ctx)
	if err != nil {
		return false, err
	}
	if live.Status.Creation != nil {
		return true, nil
	}
	if !shouldBindOrdinaryCreation(live) {
		return false, errStaleTenant
	}
	before := live.DeepCopy()
	live.Status.Creation = &v1alpha1.TenantCreation{
		TenantUID: live.UID, ClusterUID: cluster.UID, ClusterName: cluster.Name, TenantName: live.Spec.TenantName,
		DatabaseCluster: cluster.Spec.ClusterName, DatabaseID: cluster.Spec.ClusterId,
	}
	if err := m.persistCreation(before, live); err != nil {
		return false, err
	}
	return true, nil
}

// Kubernetes can accept a status patch while pruning fields from an older CRD.
func (m *OBTenantManager) persistCreation(before, live *v1alpha1.OBTenant) error {
	expected := live.Status.Creation.DeepCopy()
	if err := m.Client.Status().Patch(m.Ctx, live, client.MergeFromWithOptions(before, client.MergeFromWithOptimisticLock{})); err != nil {
		return err
	}
	var persisted v1alpha1.OBTenant
	if err := m.reader().Get(m.Ctx, client.ObjectKeyFromObject(live), &persisted); err != nil {
		return err
	}
	if persisted.UID != live.UID {
		return errStaleTenant
	}
	actual := persisted.Status.Creation
	if actual == nil {
		return errors.New("API server did not persist status.creation; install the matching OBTenant CRD")
	}
	if !sameCreationSnapshot(*expected, *actual) {
		return errors.New("API server did not preserve the tenant creation identities; check the OBTenant CRD and concurrent writes")
	}
	return nil
}

func sameCreationSnapshot(expected, actual v1alpha1.TenantCreation) bool {
	resourcesMatch := maps.Equal(expected.Resources, actual.Resources)
	expected.Resources, actual.Resources = nil, nil
	return resourcesMatch && reflect.DeepEqual(expected, actual)
}

func (m *OBTenantManager) currentCreation(ctx context.Context) (*v1alpha1.OBTenant, *v1alpha1.OBCluster, error) {
	live, cluster, err := m.readCreationObjects(ctx)
	if err != nil {
		return nil, nil, err
	}
	b := live.Status.Creation
	if b == nil || b.ClusterUID != cluster.UID || b.TenantUID != live.UID || b.ClusterName != cluster.Name ||
		b.TenantName != live.Spec.TenantName || b.DatabaseCluster != cluster.Spec.ClusterName || b.DatabaseID != cluster.Spec.ClusterId {
		return nil, nil, errCreationBindingChanged
	}
	return live, cluster, nil
}

func (m *OBTenantManager) creationDB(ctx context.Context) (*v1alpha1.OBTenant, *operation.OceanbaseOperationManager, error) {
	live, cluster, err := m.currentCreation(ctx)
	if err != nil {
		return nil, nil, err
	}
	db, err := resourceutils.GetSysOperationClient(m.Client, m.Logger, cluster)
	if err != nil {
		return nil, nil, err
	}
	if err := verifyCreationCluster(ctx, db, live.Status.Creation); err != nil {
		return nil, nil, err
	}
	return live, db, nil
}

func verifyCreationCluster(ctx context.Context, db *operation.OceanbaseOperationManager, b *v1alpha1.TenantCreation) error {
	var matches int
	const query = "SELECT COUNT(DISTINCT name) FROM oceanbase.GV$OB_PARAMETERS WHERE (name='cluster' AND value=?) OR (name='cluster_id' AND value=?)"
	if err := db.QueryRow(ctx, &matches, query, b.DatabaseCluster, strconv.FormatInt(b.DatabaseID, 10)); err != nil {
		return err
	}
	if matches != 2 {
		return errors.New("database cluster identity differs from the tenant creation record")
	}
	return nil
}

func (m *OBTenantManager) recordResource(zone string, old, ids v1alpha1.TenantCreationResources) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		live, _, err := m.currentCreation(m.Ctx)
		if err != nil {
			return err
		}
		before := live.DeepCopy()
		if live.Status.Creation.Resources[zone] != old {
			return errors.New("recorded resource identities changed")
		}
		if live.Status.Creation.Resources == nil {
			live.Status.Creation.Resources = make(map[string]v1alpha1.TenantCreationResources)
		}
		live.Status.Creation.Resources[zone] = ids
		return m.persistCreation(before, live)
	})
}

// An acknowledged strict CREATE is the only point at which a new physical ID
// may enter the record. If its result is lost, leave the allocation untouched.
func (m *OBTenantManager) createOwnedResources(pool v1alpha1.ResourcePoolSpec) error {
	if !m.OBTenant.DeletionTimestamp.IsZero() {
		return errors.New("tenant is being deleted")
	}
	live, db, err := m.creationDB(m.Ctx)
	if err != nil {
		return err
	}
	unitName, poolName := m.generateUnitName(pool.Zone), m.generatePoolName(pool.Zone)
	ids := live.Status.Creation.Resources[pool.Zone]
	exists, err := db.CheckUnitConfigExistByName(m.Ctx, unitName)
	if err != nil {
		return err
	}
	if !exists {
		unit := m.generateModelUnitConfigV4SQLParam(unitName, m.generateModelUnitConfigV4(pool.UnitConfig))
		if err := db.CreateUnitConfig(m.Ctx, unit); err != nil {
			return fmt.Errorf("unit %s creation was not acknowledged; any unrecorded allocation is retained: %w", unitName, err)
		}
	}
	unit, err := db.GetUnitConfigV4ByName(m.Ctx, unitName)
	if err != nil {
		return err
	}
	if unit.UnitConfigID <= 0 || exists && unit.UnitConfigID != ids.UnitConfigID {
		return fmt.Errorf("unit configuration %s is not owned by this creation", unitName)
	}
	if !exists {
		next := v1alpha1.TenantCreationResources{UnitConfigID: unit.UnitConfigID}
		if err := m.recordResource(pool.Zone, ids, next); err != nil {
			return fmt.Errorf("could not confirm the persisted identity of unit %s: %w", unitName, err)
		}
		ids = next
	}
	exists, err = db.CheckPoolExistByName(m.Ctx, poolName)
	if err != nil {
		return err
	}
	if !exists {
		if _, _, err := m.currentCreation(m.Ctx); err != nil {
			return err
		}
		if err := db.CreatePool(m.Ctx, model.PoolSQLParam{PoolName: poolName, UnitName: unitName, ZoneList: pool.Zone, UnitNum: int64(live.Spec.UnitNumber)}); err != nil {
			return fmt.Errorf("pool %s creation was not acknowledged; any unrecorded allocation is retained: %w", poolName, err)
		}
	}
	actual, err := db.GetPoolByName(m.Ctx, poolName)
	if err != nil {
		return err
	}
	if actual.ResourcePoolID <= 0 || exists && actual.ResourcePoolID != ids.PoolID || actual.ZoneList != pool.Zone ||
		actual.UnitConfigID != ids.UnitConfigID || actual.UnitNum != int64(live.Spec.UnitNumber) {
		return fmt.Errorf("resource pool %s differs from the owned allocation", poolName)
	}
	if exists {
		return nil
	}
	next := ids
	next.PoolID = actual.ResourcePoolID
	return m.recordResource(pool.Zone, ids, next)
}

func (m *OBTenantManager) checkCreateAllocation(db *operation.OceanbaseOperationManager) error {
	live, _, err := m.currentCreation(m.Ctx)
	if err != nil {
		return err
	}
	if err := verifyCreationCluster(m.Ctx, db, live.Status.Creation); err != nil {
		return err
	}
	for _, pool := range live.Spec.Pools {
		actual, err := db.GetPoolByName(m.Ctx, m.generatePoolName(pool.Zone))
		if err != nil {
			return err
		}
		ids := live.Status.Creation.Resources[pool.Zone]
		if ids.PoolID == 0 || ids.UnitConfigID == 0 || actual.ResourcePoolID != ids.PoolID || actual.UnitConfigID != ids.UnitConfigID ||
			actual.ZoneList != pool.Zone || actual.UnitNum != int64(live.Spec.UnitNumber) {
			return fmt.Errorf("pool %s is not the allocation provisioned for this tenant", actual.Name)
		}
		if actual.TenantID.Valid {
			rows, err := createdRows(m.Ctx, db, live.Status.Creation)
			if err != nil {
				return err
			}
			if len(rows) > 0 && rows[0].Owned && rows[0].TenantID == actual.TenantID.Int64 {
				return errAllocationAssigned
			}
			return fmt.Errorf("pool %s is attached to another tenant", actual.Name)
		}
	}
	_, _, err = m.currentCreation(m.Ctx)
	return err
}

// Finalization only removes resources whose physical identity is still recorded.
// Foreign tenants, pools that were reassigned, and shared units survive retries.
func (m *OBTenantManager) deleteCreatedTenant() error {
	live, db, err := m.creationDB(m.Ctx)
	if errors.Is(err, errForeignCluster) {
		return nil
	}
	if err != nil {
		return err
	}
	b := live.Status.Creation
	rows, err := createdRows(m.Ctx, db, b)
	if err != nil {
		return err
	}
	if len(rows) == 0 || rows[0].ClusterMatch != 2 {
		return errCreationBindingChanged
	}
	observed := rows[0]
	observed.Owned = observed.Owned || (b.LegacyTenantID > 1 && b.LegacyTenantID == observed.TenantID)
	if observed.Owned && observed.TenantID > 1 && observed.TenantType == "USER" {
		if err := db.DeleteTenant(m.Ctx, b.TenantName, live.Spec.ForceDelete); err != nil {
			return err
		}
	}
	return m.cleanupCreatedResources(db, b, observed)
}

type createdResourceCleaner interface {
	GetPoolList(context.Context) ([]model.Pool, error)
	GetUnitConfigV4List(context.Context) ([]model.UnitConfigV4, error)
	DeletePool(context.Context, string) error
	DeleteUnitConfig(context.Context, string) error
}

func (m *OBTenantManager) cleanupCreatedResources(db createdResourceCleaner, b *v1alpha1.TenantCreation, observed createdTenantRow) error {
	for zone, ids := range b.Resources {
		if ids.Borrowed {
			continue
		}
		pools, err := db.GetPoolList(m.Ctx)
		if err != nil {
			return err
		}
		for _, pool := range pools {
			if pool.ResourcePoolID != ids.PoolID || pool.Name != m.generatePoolName(zone) || pool.ZoneList != zone || pool.UnitConfigID != ids.UnitConfigID {
				continue
			}
			if pool.TenantID.Valid {
				if observed.Owned && pool.TenantID.Int64 == observed.TenantID {
					return fmt.Errorf("waiting for pool %s to detach from deleted tenant", pool.Name)
				}
				continue
			}
			if err := db.DeletePool(m.Ctx, pool.Name); err != nil {
				return err
			}
		}
		// Query again after a drop. Other pools can still reference the unit.
		pools, err = db.GetPoolList(m.Ctx)
		if err != nil {
			return err
		}
		configs, err := db.GetUnitConfigV4List(m.Ctx)
		if err != nil {
			return err
		}
		name := m.generateUnitName(zone)
		if ids.UnitConfigID > 0 && removableUnitConfig(ids.UnitConfigID, name, configs, pools) {
			if err := db.DeleteUnitConfig(m.Ctx, name); err != nil {
				return err
			}
		}
	}
	return nil
}

// Snapshot pools while the database still attests their association with the
// legacy tenant. Once DROP detaches them, names alone cannot authorize cleanup.
func (m *OBTenantManager) finalizeLegacyOrdinaryTenant() error {
	expectedID := int64(m.OBTenant.Status.TenantRecordInfo.TenantID)
	if expectedID <= 1 {
		return nil
	}
	live, cluster, err := m.readCreationObjects(m.Ctx)
	if errors.Is(err, errForeignCluster) {
		return nil
	}
	if err != nil {
		return err
	}
	if live.Status.Creation != nil {
		return m.deleteCreatedTenant()
	}
	if live.DeletionTimestamp.IsZero() || int64(live.Status.TenantRecordInfo.TenantID) != expectedID {
		return errors.New("legacy tenant changed before deletion")
	}
	b := &v1alpha1.TenantCreation{
		ClusterUID: cluster.UID, TenantUID: live.UID, ClusterName: cluster.Name, TenantName: live.Spec.TenantName,
		DatabaseCluster: cluster.Spec.ClusterName, DatabaseID: cluster.Spec.ClusterId, LegacyTenantID: expectedID,
		Resources: make(map[string]v1alpha1.TenantCreationResources),
	}
	db, err := resourceutils.GetSysOperationClient(m.Client, m.Logger, cluster)
	if err != nil {
		return err
	}
	rows, err := createdRows(m.Ctx, db, b)
	if err != nil {
		return err
	}
	if len(rows) == 0 || rows[0].ClusterMatch != 2 {
		return errors.New("cannot verify legacy tenant database identity")
	}
	if rows[0].TenantID != expectedID || rows[0].TenantType != "USER" {
		return nil
	}
	for _, pool := range live.Spec.Pools {
		for _, row := range rows {
			if row.Zone == pool.Zone && row.Pool == m.generatePoolName(pool.Zone) && row.UnitConfig == m.generateUnitName(pool.Zone) && row.PoolID > 0 && row.UnitID > 0 {
				b.Resources[pool.Zone] = v1alpha1.TenantCreationResources{PoolID: row.PoolID, UnitConfigID: row.UnitID}
			}
		}
	}
	if err := m.persistLegacyDeletion(b); err != nil {
		return err
	}
	return m.deleteCreatedTenant()
}

// The allocation query can span coordinator status writes. Take a new resource
// version on every retry and patch only the ownership snapshot. The delete step
// rechecks both the tenant and allocation IDs against the database.
func (m *OBTenantManager) persistLegacyDeletion(snapshot *v1alpha1.TenantCreation) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		live, cluster, err := m.readCreationObjects(m.Ctx)
		if err != nil {
			return err
		}
		if live.DeletionTimestamp.IsZero() || live.UID != snapshot.TenantUID || cluster.UID != snapshot.ClusterUID ||
			live.Spec.TenantName != snapshot.TenantName || cluster.Name != snapshot.ClusterName ||
			cluster.Spec.ClusterName != snapshot.DatabaseCluster || cluster.Spec.ClusterId != snapshot.DatabaseID ||
			int64(live.Status.TenantRecordInfo.TenantID) != snapshot.LegacyTenantID {
			return errStaleTenant
		}
		if live.Status.Creation != nil {
			if !sameCreationSnapshot(*live.Status.Creation, *snapshot) {
				return errCreationBindingChanged
			}
			return nil
		}
		before := live.DeepCopy()
		live.Status.Creation = snapshot.DeepCopy()
		return m.persistCreation(before, live)
	})
}

// Only a worker dispatched from a persisted legacy Pending CREATE can enter
// here. A lost Running CREATE is observed separately and is never submitted.
func (m *OBTenantManager) bindPreparedLegacyAllocation() error {
	return retry.RetryOnConflict(retry.DefaultRetry, m.recordPreparedLegacyAllocation)
}

func (m *OBTenantManager) recordPreparedLegacyAllocation() error {
	live, cluster, err := m.readCreationObjects(m.Ctx)
	if err != nil {
		return err
	}
	if live.Status.Creation != nil {
		return nil
	}
	if !legacyCreateDispatched(live) {
		return errors.New("tenant has no recorded creation or completed legacy pool provisioning")
	}
	b := &v1alpha1.TenantCreation{ClusterUID: cluster.UID, TenantUID: live.UID, ClusterName: cluster.Name, TenantName: live.Spec.TenantName,
		DatabaseCluster: cluster.Spec.ClusterName, DatabaseID: cluster.Spec.ClusterId, Resources: make(map[string]v1alpha1.TenantCreationResources)}
	db, err := resourceutils.GetSysOperationClient(m.Client, m.Logger, cluster)
	if err != nil {
		return err
	}
	if err := verifyCreationCluster(m.Ctx, db, b); err != nil {
		return err
	}
	exists, err := db.CheckTenantExistByName(m.Ctx, b.TenantName)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("pre-upgrade tenant already exists without creation ownership evidence")
	}
	pools, err := db.GetPoolList(m.Ctx)
	if err != nil {
		return err
	}
	configs, err := db.GetUnitConfigV4List(m.Ctx)
	if err != nil {
		return err
	}
	if err := m.inspectLegacyAllocation(b, pools, configs); err != nil {
		return err
	}
	before := live.DeepCopy()
	live.Status.Creation = b
	return m.persistCreation(before, live)
}

func legacyCreateDispatched(t *v1alpha1.OBTenant) bool {
	op := t.Status.OperationContext
	return t.Status.Status == tenantstatus.CreatingTenant && op != nil && op.Task == tCreateTenantWithClear &&
		(op.TaskStatus == taskstatus.Pending || op.TaskStatus == taskstatus.Running) && op.Idx >= 2 && op.Idx <= len(op.Tasks) &&
		op.Tasks[op.Idx-2] == tCreateResourcePoolAndConfig && op.Tasks[op.Idx-1] == tCreateTenantWithClear
}

func (m *OBTenantManager) inspectLegacyAllocation(b *v1alpha1.TenantCreation, pools []model.Pool, configs []model.UnitConfigV4) error {
	if len(m.OBTenant.Spec.Pools) == 0 {
		return errors.New("no prepared legacy resource pools")
	}
	for _, desired := range m.OBTenant.Spec.Pools {
		name := m.generatePoolName(desired.Zone)
		unitName := m.generateUnitName(desired.Zone)
		var unitID int64
		for _, unit := range configs {
			if unit.Name == unitName && desired.UnitConfig != nil && IsUnitConfigV4Equal(m.generateModelUnitConfigV4(desired.UnitConfig), &unit) {
				unitID = unit.UnitConfigID
			}
		}
		var poolID int64
		for _, pool := range pools {
			if pool.UnitConfigID != unitID || unitID <= 0 {
				continue
			}
			if pool.Name != name || pool.ZoneList != desired.Zone || pool.UnitNum != int64(m.OBTenant.Spec.UnitNumber) || pool.TenantID.Valid {
				return fmt.Errorf("legacy allocation %s is shared, assigned or differs from its specification", name)
			}
			poolID = pool.ResourcePoolID
		}
		if poolID <= 0 || b.Resources[desired.Zone].PoolID != 0 {
			return fmt.Errorf("no unique matching prepared legacy resource pool %s", name)
		}
		b.Resources[desired.Zone] = v1alpha1.TenantCreationResources{UnitConfigID: unitID, PoolID: poolID, Borrowed: true}
	}
	return nil
}
