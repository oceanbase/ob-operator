package obtenant

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/const/status/tenantstatus"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

type interruptedPoolDB struct {
	tenant       model.OBTenant
	pools        []model.Pool
	configs      []model.UnitConfigV4
	beforeDetach func()
	actions      []string
}

func (db *interruptedPoolDB) GetTenantByName(context.Context, string) (*model.OBTenant, error) {
	return &db.tenant, nil
}
func (db *interruptedPoolDB) GetPoolList(context.Context) ([]model.Pool, error) {
	return slices.Clone(db.pools), nil
}
func (db *interruptedPoolDB) GetUnitConfigV4List(context.Context) ([]model.UnitConfigV4, error) {
	return slices.Clone(db.configs), nil
}
func (db *interruptedPoolDB) CheckRsJobExistByTenantID(context.Context, int) (bool, error) {
	return false, nil
}
func (db *interruptedPoolDB) SetTenant(_ context.Context, param model.TenantSQLParam) error {
	if len(param.PoolList) == 0 {
		return errors.New("unexpected change to an already converged locality")
	}
	if db.beforeDetach != nil {
		db.beforeDetach()
	}
	db.actions = append(db.actions, "detach")
	for i, pool := range db.pools {
		if pool.TenantID.Valid && pool.TenantID.Int64 == db.tenant.TenantID && !slices.Contains(param.PoolList, pool.Name) {
			db.pools[i].TenantID = sql.NullInt64{}
		}
	}
	return context.DeadlineExceeded // SQL committed; caller never received the acknowledgment.
}
func (db *interruptedPoolDB) DeletePool(_ context.Context, name string) error {
	db.actions = append(db.actions, "drop "+name)
	db.pools = slices.DeleteFunc(db.pools, func(p model.Pool) bool { return p.Name == name })
	return context.DeadlineExceeded
}
func (db *interruptedPoolDB) DeleteUnitConfig(_ context.Context, name string) error {
	db.actions = append(db.actions, "drop "+name)
	db.configs = slices.DeleteFunc(db.configs, func(u model.UnitConfigV4) bool { return u.Name == name })
	return nil
}

func newPoolRemovalTest(t *testing.T) (*OBTenantManager, *interruptedPoolDB) {
	t.Helper()
	cr := &v1alpha1.OBTenant{ObjectMeta: metav1.ObjectMeta{Name: "orders", Namespace: "test", UID: "tenant-uid", Generation: 2, Finalizers: []string{"operator.oceanbase.com/test"}},
		Spec:   v1alpha1.OBTenantSpec{ClusterName: "cluster", TenantName: "orders", Pools: []v1alpha1.ResourcePoolSpec{{Zone: "zone1", Type: &v1alpha1.LocalityType{Name: "FULL", IsActive: true}}}},
		Status: v1alpha1.OBTenantStatus{Pools: []v1alpha1.ResourcePoolStatus{{ZoneList: "zone1"}, {ZoneList: "zone3"}}}}
	cluster := &v1alpha1.OBCluster{ObjectMeta: metav1.ObjectMeta{Name: "cluster", Namespace: "test", UID: "cluster-uid", Annotations: map[string]string{"example.org/keep": "yes", "oceanbase.oceanbase.com/pool-removal-other": "another tenant"}}}
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	logger := logr.Discard()
	m := &OBTenantManager{Ctx: context.Background(), OBTenant: cr, Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(cr.DeepCopy(), cluster).Build(), Logger: &logger}
	owner := sql.NullInt64{Valid: true, Int64: 1002}
	db := &interruptedPoolDB{tenant: model.OBTenant{TenantID: 1002, TenantName: "orders", Locality: "FULL{1}@zone1"},
		pools:   []model.Pool{{Name: "pool_orders_zone1", ZoneList: "zone1", ResourcePoolID: 1001, UnitConfigID: 1001, TenantID: owner}, {Name: "pool_orders_zone3", ZoneList: "zone3", ResourcePoolID: 1003, UnitConfigID: 1007, TenantID: owner}},
		configs: []model.UnitConfigV4{{Name: "unitconfig_orders_zone3", UnitConfigID: 1007}}}
	return m, db
}

func readRemovalCluster(t *testing.T, m *OBTenantManager) *v1alpha1.OBCluster {
	t.Helper()
	cluster, err := m.getOBCluster()
	if err != nil {
		t.Fatal(err)
	}
	return cluster
}

func TestInterruptedDetachedPoolResumesByPhysicalID(t *testing.T) {
	m, db := newPoolRemovalTest(t)
	db.beforeDetach = func() {
		cluster := readRemovalCluster(t, m)
		var intent poolRemovalIntent
		if err := json.Unmarshal([]byte(cluster.Annotations[m.poolRemovalKey()]), &intent); err != nil {
			t.Fatal(err)
		}
		if intent.PoolID != 1003 || intent.UnitConfigID != 1007 || intent.TenantID != 1002 || intent.ClusterUID != cluster.UID || intent.TenantUID != m.OBTenant.UID {
			t.Fatalf("invalid persisted identity %+v", intent)
		}
	}
	if err := m.removePool(m.OBTenant, readRemovalCluster(t, m), db); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("detach: %v", err)
	}
	if db.pools[1].TenantID.Valid {
		t.Fatal("database did not commit the detach")
	}
	fresh := *m
	fresh.OBTenant = m.OBTenant.DeepCopy()
	fresh.OBTenant.Status.Pools = fresh.OBTenant.Status.Pools[:1]
	if !fresh.CanReplayTask(tDeletePool) {
		t.Fatal("replacement manager cannot resume the recorded task")
	}
	if status, err := fresh.NextStatus(); err != nil || status != tenantstatus.DeletingResourcePool {
		t.Fatalf("new worker lost detached pool: %q %v", status, err)
	}
	if err := fresh.removePool(fresh.OBTenant, readRemovalCluster(t, &fresh), db); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("drop: %v", err)
	}
	if err := fresh.removePool(fresh.OBTenant, readRemovalCluster(t, &fresh), db); err != nil {
		t.Fatal(err)
	}
	if err := fresh.removePool(fresh.OBTenant, readRemovalCluster(t, &fresh), db); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(db.actions, []string{"detach", "drop pool_orders_zone3", "drop unitconfig_orders_zone3"}) || len(db.pools) != 1 || len(db.configs) != 0 {
		t.Fatalf("unexpected final database %+v", db)
	}
	annotations := readRemovalCluster(t, &fresh).Annotations
	if len(annotations) != 2 || annotations["example.org/keep"] != "yes" || annotations["oceanbase.oceanbase.com/pool-removal-other"] != "another tenant" {
		t.Fatalf("unexpected cluster annotations: %v", annotations)
	}
}

func TestPoolRemovalPreservesChangedOwners(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*OBTenantManager, *interruptedPoolDB)
	}{
		{"same name new ID", func(_ *OBTenantManager, db *interruptedPoolDB) { db.pools[1].ResourcePoolID = 2000 }},
		{"same ID new name", func(_ *OBTenantManager, db *interruptedPoolDB) { db.pools[1].Name = "manual" }},
		{"changed config", func(_ *OBTenantManager, db *interruptedPoolDB) { db.pools[1].UnitConfigID = 5000 }},
		{"assigned elsewhere", func(_ *OBTenantManager, db *interruptedPoolDB) {
			db.pools[1].TenantID = sql.NullInt64{Valid: true, Int64: 1004}
		}},
		{"replacement tenant", func(_ *OBTenantManager, db *interruptedPoolDB) { db.tenant.TenantID = 2002 }},
		{"shared configuration", func(_ *OBTenantManager, db *interruptedPoolDB) {
			db.pools = db.pools[:1]
			db.pools[0].UnitConfigID = 1007
		}},
		{"configuration name reused", func(_ *OBTenantManager, db *interruptedPoolDB) {
			db.pools = db.pools[:1]
			db.configs[0].UnitConfigID = 3007
		}},
		{"explicit inactive", func(m *OBTenantManager, _ *interruptedPoolDB) {
			m.OBTenant.Spec.Pools = append(m.OBTenant.Spec.Pools, v1alpha1.ResourcePoolSpec{Zone: "zone3", Type: &v1alpha1.LocalityType{}})
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			m, db := newPoolRemovalTest(t)
			cluster := readRemovalCluster(t, m)
			intent := poolRemovalIntent{TenantUID: m.OBTenant.UID, ClusterUID: cluster.UID, TenantID: 1002, TenantName: "orders", Zone: "zone3", PoolID: 1003, UnitConfigID: 1007}
			record, err := json.Marshal(intent)
			if err != nil {
				t.Fatal(err)
			}
			if err := m.writePoolRemoval(cluster, string(record)); err != nil {
				t.Fatal(err)
			}
			db.pools[1].TenantID = sql.NullInt64{}
			test.change(m, db)
			before := slices.Clone(db.pools)
			if err := m.removePool(m.OBTenant, readRemovalCluster(t, m), db); err != nil {
				t.Fatal(err)
			}
			if len(db.actions) != 0 || !reflect.DeepEqual(db.pools, before) || len(db.configs) != 1 || readRemovalCluster(t, m).Annotations[m.poolRemovalKey()] != "" {
				t.Fatalf("foreign or desired state changed: %+v", db)
			}
		})
	}
}

func TestPoolRemovalReservationConflictsBeforeSQL(t *testing.T) {
	m, db := newPoolRemovalTest(t)
	stale := readRemovalCluster(t, m)
	changed := stale.DeepCopy()
	changed.Annotations["example.org/other-change"] = "1"
	if err := m.Client.Patch(m.Ctx, changed, client.MergeFrom(stale)); err != nil {
		t.Fatal(err)
	}
	if err := m.removePool(m.OBTenant, stale, db); !apierrors.IsConflict(err) {
		t.Fatalf("expected conflict, got %v", err)
	}
	if len(db.actions) != 0 || readRemovalCluster(t, m).Annotations[m.poolRemovalKey()] != "" {
		t.Fatal("unreserved SQL executed")
	}
}

func TestPoolRemovalReservationClearedByTenantFinalizer(t *testing.T) {
	m, _ := newPoolRemovalTest(t)
	cluster := readRemovalCluster(t, m)
	if err := m.writePoolRemoval(cluster, `{"tenantUID":"replaced"}`); err != nil {
		t.Fatal(err)
	}
	if err := m.Client.Delete(m.Ctx, m.OBTenant); err != nil {
		t.Fatal(err)
	}
	current := new(v1alpha1.OBTenant)
	if err := m.Client.Get(m.Ctx, client.ObjectKeyFromObject(m.OBTenant), current); err != nil {
		t.Fatal(err)
	}
	m.OBTenant = current
	m.OBTenant.Status.Status = tenantstatus.FinalizerFinished
	if err := m.CheckAndUpdateFinalizers(); err != nil {
		t.Fatal(err)
	}
	annotations := readRemovalCluster(t, m).Annotations
	if annotations[m.poolRemovalKey()] != "" || annotations["oceanbase.oceanbase.com/pool-removal-other"] != "another tenant" {
		t.Fatalf("unexpected annotations after deleting tenant: %v", annotations)
	}
}

func TestTenantDeletionFinishesInterruptedPhysicalRemoval(t *testing.T) {
	m, db := newPoolRemovalTest(t)
	cluster := readRemovalCluster(t, m)
	intent := poolRemovalIntent{TenantUID: m.OBTenant.UID, ClusterUID: cluster.UID, TenantID: 1002, TenantName: "orders", Zone: "zone3", PoolID: 1003, UnitConfigID: 1007}
	record, err := json.Marshal(intent)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.writePoolRemoval(cluster, string(record)); err != nil {
		t.Fatal(err)
	}
	// The tenant DDL may not have released the pool yet. Never issue another
	// ALTER against a tenant that is already deleting.
	if err := m.finishDetachedPool(db, readRemovalCluster(t, m), intent); err == nil || len(db.actions) != 0 {
		t.Fatalf("attached pool was modified: %v %v", err, db.actions)
	}
	db.pools[1].TenantID = sql.NullInt64{}
	if err := m.finishDetachedPool(db, readRemovalCluster(t, m), intent); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected lost DROP acknowledgment: %v", err)
	}
	if len(readRemovalCluster(t, m).Annotations[m.poolRemovalKey()]) == 0 {
		t.Fatal("finalizer lost deletion record")
	}
	if err := m.finishDetachedPool(db, readRemovalCluster(t, m), intent); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(db.actions, []string{"drop pool_orders_zone3", "drop unitconfig_orders_zone3"}) || readRemovalCluster(t, m).Annotations[m.poolRemovalKey()] != "" {
		t.Fatalf("finalizer did not finish exactly the detached objects: %v", db.actions)
	}
}
