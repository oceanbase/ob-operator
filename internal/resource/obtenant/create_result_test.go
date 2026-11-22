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

package obtenant

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/const/status/tenantstatus"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func TestCreateTenantResultPreservesDefinitiveErrorsAndVerifiesLostWrites(t *testing.T) {
	if err := confirmTenantCreate(false, io.EOF, func() error { t.Fatal("verified an unsent CREATE"); return nil }); err != io.EOF {
		t.Fatal("lost pre-dispatch error")
	}
	rejected := &mysql.MySQLError{Number: 1045}
	probeErr := errors.New("ownership could not be verified")
	for _, tc := range []struct {
		name      string
		createErr error
		check     bool
	}{
		{"success", nil, false},
		{"rejected", rejected, false},
		{"server timeout", &mysql.MySQLError{Number: 4012}, true},
		{"root service changed", &mysql.MySQLError{Number: 4638}, true},
		{"tenant already exists", &mysql.MySQLError{Number: 5156}, true},
		{"connection lost", io.EOF, true},
		{"driver lost", driver.ErrBadConn, true},
		{"deadline", fmt.Errorf("execute: %w", context.DeadlineExceeded), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, result := range []error{nil, probeErr} {
				observed := 0
				err := confirmTenantCreate(true, tc.createErr, func() error { observed++; return result })
				if (observed == 1) != tc.check {
					t.Fatalf("unexpected result check count %d", observed)
				}
				switch {
				case !tc.check && err != tc.createErr:
					t.Fatalf("changed database result: %v", err)
				case tc.check && result != nil && (!errors.Is(err, tc.createErr) || !errors.Is(err, probeErr)):
					t.Fatalf("original CREATE result or verification error lost: %v", err)
				case tc.check && result == nil && err != nil:
					t.Fatalf("verified completion failed: %v", err)
				}
			}
		})
	}
}

func TestCreatedTenantNeedsExactOwnedAllocation(t *testing.T) {
	m := &OBTenantManager{OBTenant: &v1alpha1.OBTenant{ObjectMeta: metav1.ObjectMeta{UID: "tenant-uid", Generation: 3},
		Spec: v1alpha1.OBTenantSpec{TenantName: "example", UnitNumber: 2, Pools: []v1alpha1.ResourcePoolSpec{{Zone: "zone1"}, {Zone: "zone2"}}}}}
	m.OBTenant.Status.Creation = &v1alpha1.TenantCreation{TenantUID: "tenant-uid", ClusterUID: "cluster-uid",
		Resources: map[string]v1alpha1.TenantCreationResources{"zone1": {PoolID: 11, UnitConfigID: 21}, "zone2": {PoolID: 12, UnitConfigID: 22}}}
	owner := createOwner(m.OBTenant.Status.Creation)
	m.OBTenant.Generation++
	if createOwner(m.OBTenant.Status.Creation) != owner || owner != "ob-operator/cluster-uid/tenant-uid" {
		t.Fatalf("owner must survive subsequent generations: %s", owner)
	}

	if m.ObserveLostTask(tCreateTenantWithClear) == nil || m.CanReplayTask(tCreateTenantWithClear) || m.ObserveLostTask(tDeletePool) != nil {
		t.Fatal("lost tenant creation must verify without replaying CREATE")
	}
	row := func(zone string) createdTenantRow {
		ids := m.OBTenant.Status.Creation.Resources[zone]
		return createdTenantRow{PoolID: ids.PoolID, UnitID: ids.UnitConfigID, ClusterMatch: 2, TenantID: 1002, Owned: true, TenantType: "USER", Status: "NORMAL", Role: "PRIMARY",
			Pool: m.generatePoolName(zone), Zone: zone, Units: 2, UnitConfig: m.generateUnitName(zone)}
	}
	for _, tc := range []struct {
		name string
		edit func([]createdTenantRow) []createdTenantRow
	}{
		{"wrong database", func(r []createdTenantRow) []createdTenantRow { r[0].ClusterMatch = 1; return r }},
		{"existing tenant without marker", func(r []createdTenantRow) []createdTenantRow { r[0].Owned = false; return r }},
		{"metadata tenant", func(r []createdTenantRow) []createdTenantRow { r[0].TenantType = "META"; return r }},
		{"standby", func(r []createdTenantRow) []createdTenantRow { r[0].Role = "STANDBY"; return r }},
		{"dropping", func(r []createdTenantRow) []createdTenantRow { r[0].Status = "DROPPING"; return r }},
		{"replaced pool", func(r []createdTenantRow) []createdTenantRow { r[0].PoolID++; return r }},
		{"replaced config", func(r []createdTenantRow) []createdTenantRow { r[0].UnitID++; return r }},
		{"missing pool", func(r []createdTenantRow) []createdTenantRow { return r[:1] }},
		{"extra pool", func(r []createdTenantRow) []createdTenantRow { return append(r, row("zone3")) }},
		{"wrong pool", func(r []createdTenantRow) []createdTenantRow { r[0].Pool = "unrelated"; return r }},
		{"wrong zone", func(r []createdTenantRow) []createdTenantRow { r[1] = row("zone1"); return r }},
		{"wrong unit", func(r []createdTenantRow) []createdTenantRow { r[1].UnitConfig = "unrelated"; return r }},
		{"wrong unit number", func(r []createdTenantRow) []createdTenantRow { r[1].Units = 1; return r }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var id int64
			ready, err := m.verifyCreatedTenant(tc.edit([]createdTenantRow{row("zone1"), row("zone2")}), &id)
			if ready || err == nil {
				t.Fatalf("accepted inconsistent result: %v %v", ready, err)
			}
		})
	}
	var id int64
	creating := row("zone1")
	creating.Status, creating.Role = "CREATING", ""
	for _, pending := range [][]createdTenantRow{{{ClusterMatch: 2}}, {creating}} {
		if ready, err := m.verifyCreatedTenant(pending, &id); ready || err != nil {
			t.Fatalf("expected to wait for server: %v %v", ready, err)
		}
	}
	good := []createdTenantRow{row("zone2"), row("zone1")}
	if ready, err := m.verifyCreatedTenant(good, &id); !ready || err != nil {
		t.Fatalf("expected successful verification: %v %v", ready, err)
	}
	id++
	if ready, err := m.verifyCreatedTenant(good, &id); ready || err == nil {
		t.Fatalf("accepted a replacement tenant: %v %v", ready, err)
	}
}

func TestCreationBindingSurvivesStatusWritesAndRejectsReplacement(t *testing.T) {
	m, _ := newPoolRemovalTest(t)
	tcr := m.OBTenant.DeepCopy()
	tcr.Status = v1alpha1.OBTenantStatus{Status: tenantstatus.CreatingTenant, Pools: []v1alpha1.ResourcePoolStatus{}}
	cluster, err := m.getOBCluster()
	if err != nil {
		t.Fatal(err)
	}
	tcr.OwnerReferences = []metav1.OwnerReference{{APIVersion: v1alpha1.GroupVersion.String(), Kind: "OBCluster", Name: cluster.Name, UID: cluster.UID}}
	m.Client = fake.NewClientBuilder().WithScheme(m.Client.Scheme()).WithStatusSubresource(tcr).WithObjects(cluster, tcr.DeepCopy()).Build()
	m.OBTenant = tcr
	if bound, err := m.BindOrdinaryCreation(); !bound || err != nil {
		t.Fatalf("binding: %v %v", bound, err)
	}
	live, _, err := m.currentCreation(m.Ctx)
	if err != nil || live.Status.Creation.TenantUID != tcr.UID {
		t.Fatalf("missing persisted binding: %v", err)
	}
	m.OBTenant = live
	allocated := v1alpha1.TenantCreationResources{UnitConfigID: 101, PoolID: 201}
	if err := m.recordResource("zone1", v1alpha1.TenantCreationResources{}, allocated); err != nil {
		t.Fatal(err)
	}
	m.StatusAtRead = live.DeepCopy().Status
	m.OBTenant.Status.Status = tenantstatus.Failed
	if err := m.retryUpdateStatus(); err != nil {
		t.Fatal(err)
	}
	live, _, err = m.currentCreation(m.Ctx)
	if err != nil || live.Status.Creation.Resources["zone1"] != allocated {
		t.Fatalf("stale status erased resource proof: %v", err)
	}
	if err := m.retryUpdateStatus(); err == nil {
		t.Fatal("stale task cursor overwrote newer status")
	}

	m.OBTenant = live.DeepCopy()
	foreign := live.DeepCopy()
	foreign.Spec.TenantName = "replacement"
	if err := m.Client.Update(m.Ctx, foreign); err != nil {
		t.Fatal(err)
	}
	if _, _, err := m.currentCreation(m.Ctx); err == nil {
		t.Fatal("mutable tenant name redirected creation")
	}
	foreign.Spec.TenantName = live.Spec.TenantName
	foreign.OwnerReferences[0].UID = "replacement-cluster"
	if err := m.Client.Update(m.Ctx, foreign); err != nil {
		t.Fatal(err)
	}
	if _, _, err := m.currentCreation(m.Ctx); err == nil {
		t.Fatal("replacement owner accepted")
	}
}

func TestFailedCreationCleanupRequiresRecordedPhysicalIDs(t *testing.T) {
	m, db := newPoolRemovalTest(t)
	initial := v1alpha1.TenantCreationResources{PoolID: 1003, UnitConfigID: 1007}
	b := &v1alpha1.TenantCreation{Resources: map[string]v1alpha1.TenantCreationResources{}}
	if err := m.cleanupCreatedResources(db, b, createdTenantRow{}); err != nil || len(db.actions) != 0 {
		t.Fatalf("unowned cleanup: %v %v", db.actions, err)
	}
	b.Resources["zone3"] = initial
	if err := m.cleanupCreatedResources(db, b, createdTenantRow{TenantID: 1002}); err != nil || len(db.actions) != 0 {
		t.Fatalf("touched foreign allocated pool: %v %v", db.actions, err)
	}
	db.pools[1].TenantID.Valid = false
	db.pools[1].ResourcePoolID++
	if err := m.cleanupCreatedResources(db, b, createdTenantRow{}); err != nil || len(db.actions) != 0 {
		t.Fatalf("touched replaced allocation: %v %v", db.actions, err)
	}
	db.pools[1].ResourcePoolID = initial.PoolID
	if err := m.cleanupCreatedResources(db, b, createdTenantRow{}); !errors.Is(err, context.DeadlineExceeded) || len(db.actions) != 1 {
		t.Fatalf("lost pool DROP acknowledgment: %v %v", db.actions, err)
	}
	if err := m.cleanupCreatedResources(db, b, createdTenantRow{}); err != nil || len(db.actions) != 2 {
		t.Fatalf("did not finish owned cleanup: %v %v", db.actions, err)
	}
}

func TestCompletedLegacyCreateKeepsUnmarkedDeletionPath(t *testing.T) {
	cr := &v1alpha1.OBTenant{Status: v1alpha1.OBTenantStatus{Status: tenantstatus.CreatingTenant}}
	if !shouldBindOrdinaryCreation(cr) {
		t.Fatal("new ordinary tenant missing binding")
	}
	for _, task := range []tasktypes.TaskName{tCreateTenantWithClear, tCreateUserWithCredentialSecrets, tOptimizeTenantByScenario} {
		cr.Status.OperationContext = &tasktypes.OperationContext{Task: task, TaskStatus: taskstatus.Successful}
		if shouldBindOrdinaryCreation(cr) {
			t.Fatalf("retrofit an already created unmarked tenant at %s", task)
		}
	}
}

func TestFinalizationRejectsStaleGenerationWithoutFinishing(t *testing.T) {
	m, _ := newPoolRemovalTest(t)
	m.OBTenant.Generation++
	if err := m.deleteCreatedTenant(); !errors.Is(err, errStaleTenant) {
		t.Fatalf("newer generation skipped cleanup: %v", err)
	}
	m.OBTenant.Status.TenantRecordInfo.TenantID = 1002
	if err := m.finalizeLegacyOrdinaryTenant(); !errors.Is(err, errStaleTenant) {
		t.Fatalf("legacy generation skipped cleanup: %v", err)
	}
}

func TestOldPendingCreateOnlyBorrowsMatchingDetachedAllocation(t *testing.T) {
	m, db := newPoolRemovalTest(t)
	m.OBTenant.Status.Status = tenantstatus.CreatingTenant
	m.OBTenant.Spec.Pools = m.OBTenant.Spec.Pools[:1]
	m.OBTenant.Spec.UnitNumber = 1
	m.OBTenant.Spec.Pools[0].UnitConfig = &v1alpha1.UnitConfig{MaxCPU: resource.MustParse("1"), MemorySize: resource.MustParse("1Gi")}
	tasks := []tasktypes.TaskName{tCheckTenant, tCheckPoolAndConfig, tCreateResourcePoolAndConfig, tCreateTenantWithClear}
	m.OBTenant.Status.OperationContext = &tasktypes.OperationContext{Tasks: tasks, Task: tCreateTenantWithClear, Idx: 4, TaskStatus: taskstatus.Pending}
	if shouldBindOrdinaryCreation(m.OBTenant) || !legacyCreateDispatched(m.OBTenant) {
		t.Fatal("pending classic CREATE received an empty new binding")
	}
	m.OBTenant.Status.OperationContext.Task = tCreateResourcePoolAndConfig
	m.OBTenant.Status.OperationContext.Idx = 3
	m.OBTenant.Status.OperationContext.TaskStatus = taskstatus.Running
	if shouldBindOrdinaryCreation(m.OBTenant) || !m.CanReplayTask(tCreateResourcePoolAndConfig) {
		t.Fatal("pre-upgrade provisioning cannot finish preparation")
	}
	if m.ObserveLostTask(tCreateTenantWithClear) == nil || m.CanReplayTask(tCreateTenantWithClear) {
		t.Fatal("lost Running CREATE must never resend")
	}
	res := &v1alpha1.TenantCreation{Resources: map[string]v1alpha1.TenantCreationResources{}}
	db.pools = db.pools[:1]
	db.pools[0].TenantID.Valid = false
	db.pools[0].UnitNum = 1
	unit := m.generateModelUnitConfigV4(m.OBTenant.Spec.Pools[0].UnitConfig)
	unit.Name = m.generateUnitName("zone1")
	unit.UnitConfigID = db.pools[0].UnitConfigID
	if err := m.inspectLegacyAllocation(res, db.pools, []model.UnitConfigV4{*unit}); err != nil {
		t.Fatal(err)
	}
	if got := res.Resources["zone1"]; !got.Borrowed || got.PoolID != db.pools[0].ResourcePoolID || got.UnitConfigID != unit.UnitConfigID {
		t.Fatalf("unverified borrowed allocation: %+v", got)
	}
	if err := m.cleanupCreatedResources(db, res, createdTenantRow{}); err != nil || len(db.actions) != 0 {
		t.Fatalf("borrowed allocation removed: %v %v", db.actions, err)
	}
	intent := poolRemovalIntent{KeepResources: true, Zone: "zone1", PoolID: db.pools[0].ResourcePoolID, UnitConfigID: unit.UnitConfigID}
	if done, err := m.advancePoolRemoval(db, nil, db.pools, intent); !done || err != nil || len(db.actions) != 0 {
		t.Fatalf("zone removal dropped borrowed assets: %v %v", err, db.actions)
	}
	db.pools[0].TenantID.Valid = true
	if err := m.inspectLegacyAllocation(&v1alpha1.TenantCreation{Resources: map[string]v1alpha1.TenantCreationResources{}}, db.pools, []model.UnitConfigV4{*unit}); err == nil {
		t.Fatal("foreign allocated pool was borrowed")
	}
	if shouldBindOrdinaryCreation(&v1alpha1.OBTenant{}) {
		t.Fatal("attempted partial status patch before InitStatus")
	}
	m.OBTenant.Status.Creation = res
	if m.CanReplayTask(tCreateResourcePoolAndConfig) || m.ObserveLostTask(tCreateResourcePoolAndConfig) == nil {
		t.Fatal("new strict provisioning replayed an unknown allocation")
	}
}

type pruneCreationReader struct{ client.Reader }

func (r pruneCreationReader) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if err := r.Reader.Get(ctx, key, obj, opts...); err != nil {
		return err
	}
	if tenant, ok := obj.(*v1alpha1.OBTenant); ok {
		tenant.Status.Creation = nil
	}
	return nil
}
func TestCreationWriteFailsWhenInstalledCRDPrunesIdentity(t *testing.T) {
	m, _ := newPoolRemovalTest(t)
	cluster, err := m.getOBCluster()
	if err != nil {
		t.Fatal(err)
	}
	cr := m.OBTenant.DeepCopy()
	cr.OwnerReferences = []metav1.OwnerReference{{APIVersion: v1alpha1.GroupVersion.String(), Kind: "OBCluster", Name: cluster.Name, UID: cluster.UID}}
	cr.Status = v1alpha1.OBTenantStatus{Status: tenantstatus.CreatingTenant, Pools: []v1alpha1.ResourcePoolStatus{}}
	m.Client = fake.NewClientBuilder().WithScheme(m.Client.Scheme()).WithStatusSubresource(cr).WithObjects(cr, cluster).Build()
	m.APIReader = pruneCreationReader{m.Client}
	m.OBTenant = cr
	bound, err := m.BindOrdinaryCreation()
	if bound || err == nil || !strings.Contains(err.Error(), "matching OBTenant CRD") {
		t.Fatalf("pruned status advanced or spun: %v %v", bound, err)
	}
}

func TestUnmarkedLegacyFinalizerRequiresOneIdentifiedClusterOwner(t *testing.T) {
	group := v1alpha1.GroupVersion.String()
	old := metav1.OwnerReference{APIVersion: group, Kind: "OBCluster", Name: "cluster", UID: "old-cluster"}
	current := old
	current.UID = "cluster-uid"
	missingUID := old
	missingUID.UID = ""
	unrelated := old
	unrelated.Name = "elsewhere"
	for _, tc := range []struct {
		name string
		refs []metav1.OwnerReference
		want error
	}{
		{"original cluster replaced", []metav1.OwnerReference{old}, nil},
		{"owner missing", nil, errCreationBindingChanged},
		{"owner UID missing", []metav1.OwnerReference{missingUID}, errCreationBindingChanged},
		{"owner name differs", []metav1.OwnerReference{unrelated}, errCreationBindingChanged},
		{"two conflicting owners", []metav1.OwnerReference{old, current}, errCreationBindingChanged},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, _ := newPoolRemovalTest(t)
			cluster, err := m.getOBCluster()
			if err != nil {
				t.Fatal(err)
			}
			cr := m.OBTenant.DeepCopy()
			stamp := metav1.Now()
			cr.DeletionTimestamp = &stamp
			cr.Status.TenantRecordInfo.TenantID = 1002
			cr.OwnerReferences = tc.refs
			m.Client = fake.NewClientBuilder().WithScheme(m.Client.Scheme()).WithObjects(cluster, cr.DeepCopy()).Build()
			m.OBTenant = &v1alpha1.OBTenant{}
			if err := m.Client.Get(m.Ctx, client.ObjectKeyFromObject(cr), m.OBTenant); err != nil {
				t.Fatal(err)
			}
			if err := m.finalizeLegacyOrdinaryTenant(); !errors.Is(err, tc.want) {
				t.Fatalf("finalizer returned %v, want %v", err, tc.want)
			}
		})
	}
}

func TestLegacyDeletionRecordsAllocationAcrossCoordinatorConflict(t *testing.T) {
	m, _ := newPoolRemovalTest(t)
	cluster, err := m.getOBCluster()
	if err != nil {
		t.Fatal(err)
	}
	cr := m.OBTenant.DeepCopy()
	stamp := metav1.Now()
	cr.DeletionTimestamp = &stamp
	cr.Status.Status = tenantstatus.DeletingTenant
	cr.Status.TenantRecordInfo.TenantID = 1002
	cr.OwnerReferences = []metav1.OwnerReference{{APIVersion: v1alpha1.GroupVersion.String(), Kind: "OBCluster", Name: cluster.Name, UID: cluster.UID}}
	writes := 0
	m.Client = fake.NewClientBuilder().WithScheme(m.Client.Scheme()).WithStatusSubresource(cr).WithObjects(cluster, cr.DeepCopy()).WithInterceptorFuncs(interceptor.Funcs{
		SubResourcePatch: func(ctx context.Context, c client.Client, sub string, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
			writes++
			if writes == 1 {
				var current v1alpha1.OBTenant
				if err := c.Get(ctx, client.ObjectKeyFromObject(cr), &current); err != nil {
					return err
				}
				current.Status.OperationContext = &tasktypes.OperationContext{Task: tDeleteTenant, TaskStatus: taskstatus.Running}
				if err := c.Status().Update(ctx, &current); err != nil {
					return err
				}
				// The fake status writer does not enforce the API server's merge-patch resource version check.
				return kubeerrors.NewConflict(v1alpha1.GroupVersion.WithResource("obtenants").GroupResource(), cr.Name, errors.New("coordinator advanced status"))
			}
			return c.SubResource(sub).Patch(ctx, obj, patch, opts...)
		},
	}).Build()
	key := client.ObjectKeyFromObject(cr)
	m.OBTenant = &v1alpha1.OBTenant{}
	if err := m.Client.Get(m.Ctx, key, m.OBTenant); err != nil {
		t.Fatal(err)
	}
	b := &v1alpha1.TenantCreation{TenantUID: cr.UID, ClusterUID: cluster.UID, ClusterName: cluster.Name, DatabaseCluster: cluster.Spec.ClusterName, DatabaseID: cluster.Spec.ClusterId,
		TenantName: cr.Spec.TenantName, LegacyTenantID: 1002, Resources: map[string]v1alpha1.TenantCreationResources{"zone1": {UnitConfigID: 101, PoolID: 201}}}
	if err := m.persistLegacyDeletion(b); err != nil {
		t.Fatal(err)
	}
	if writes != 2 {
		t.Fatalf("expected a resource version conflict and one retry; observed %d writes", writes)
	}
	var live v1alpha1.OBTenant
	if err := m.Client.Get(m.Ctx, key, &live); err != nil {
		t.Fatal(err)
	}
	if live.Status.OperationContext == nil || live.Status.OperationContext.TaskStatus != taskstatus.Running ||
		live.Status.Creation == nil || live.Status.Creation.LegacyTenantID != 1002 || live.Status.Creation.Resources["zone1"] != b.Resources["zone1"] {
		t.Fatalf("lost coordinator status or verified allocation: %+v", live.Status)
	}
}
