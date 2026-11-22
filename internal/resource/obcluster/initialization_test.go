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
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-sql-driver/mysql"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	"github.com/oceanbase/ob-operator/pkg/database"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func TestPostBootstrapResumesFailedInitialization(t *testing.T) {
	logger := logr.Discard()
	m := &OBClusterManager{OBCluster: &v1alpha1.OBCluster{}, Logger: &logger}
	m.OBCluster.Status.Status = clusterstatus.Bootstrapped
	fresh, err := m.GetTaskFlow()
	if err != nil || fresh.OperationContext.OnFailure.Strategy != strategy.RetryFromCurrent {
		t.Fatalf("post-bootstrap failure rule: %v, %v", fresh, err)
	}
	op := fresh.OperationContext
	for op.Task != tCreateUsers {
		if !fresh.HasNext() {
			t.Fatal("missing create users task")
		}
		fresh.NextTask()
	}
	op.TaskStatus = taskstatus.Failed
	op.OnFailure = tasktypes.FailureRule{Strategy: strategy.StartOver, NextTryStatus: clusterstatus.Running, RetryCount: 1}
	m.SetOperationContext(op)
	// A new manager reads the older persisted flow after a controller restart.
	resumed := &OBClusterManager{OBCluster: m.OBCluster.DeepCopy(), Logger: &logger}
	flow, err := resumed.GetTaskFlow()
	if err != nil {
		t.Fatal(err)
	}
	resumed.HandleFailure()
	current := flow.OperationContext
	if resumed.GetStatus() != clusterstatus.Bootstrapped || resumed.OBCluster.Status.OperationContext != current || current.Idx != op.Idx || current.Task != tCreateUsers || current.TaskStatus != taskstatus.Pending || current.OnFailure.RetryCount != 1 {
		t.Fatal("retry skipped pending user initialization or discarded its remaining tasks")
	}
	for _, task := range op.Tasks[op.Idx:] {
		if got := flow.NextTask(); got != task {
			t.Fatalf("next task = %q, want %q", got, task)
		}
	}
	resumed.FinishTask()
	if resumed.GetStatus() != clusterstatus.Running || resumed.OBCluster.Status.OperationContext != nil {
		t.Fatal("completed initializer did not finish")
	}
	for _, tc := range []struct {
		status   string
		name     tasktypes.FlowName
		strategy tasktypes.TaskFailureStrategy
		next     string
	}{
		{clusterstatus.Bootstrapped, postBootstrapFlow, strategy.Pause, clusterstatus.Running},
		{clusterstatus.Bootstrapped, postBootstrapFlow, strategy.StartOver, clusterstatus.Failed},
		{clusterstatus.MigrateFromExisting, "migrate obcluster from existing", strategy.StartOver, clusterstatus.Failed},
		{clusterstatus.Running, postBootstrapFlow, strategy.StartOver, clusterstatus.Running},
	} {
		previous := op.DeepCopy()
		previous.Name = tc.name
		previous.OnFailure.Strategy = tc.strategy
		previous.OnFailure.NextTryStatus = tc.next
		m.OBCluster.Status.Status = tc.status
		m.SetOperationContext(previous)
		f, err := m.GetTaskFlow()
		if err != nil || f.OperationContext.OnFailure != previous.OnFailure || previous.OnFailure.Strategy != tc.strategy || previous.OnFailure.NextTryStatus != tc.next {
			t.Fatal("unrelated or explicitly configured flow was modified")
		}
	}
}

func TestInitializationRootCredentialsAfterRestart(t *testing.T) {
	configured := t.Name()
	denied := fmt.Errorf("open: %w", &mysql.MySQLError{Number: 1045})
	for _, tc := range []struct {
		name        string
		creating    bool
		first       error
		initial     error
		wantInitial bool
		wantErr     bool
	}{
		{"fresh", true, denied, nil, true, false},
		{"root initialized before restart", true, nil, nil, false, false},
		{"configured and initial both rejected", true, denied, denied, true, true},
		{"temporary connection failure", true, context.DeadlineExceeded, nil, false, true},
		{"optimization", false, nil, nil, false, false},
		{"optimization cannot use initial credentials", false, denied, nil, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want := &operation.OceanbaseOperationManager{}
			var triedInitial []bool
			got, err := connectInitializationRoot(configured, tc.creating, func(value string) (*operation.OceanbaseOperationManager, error) {
				triedInitial = append(triedInitial, value == "")
				switch value {
				case configured:
					if tc.first != nil {
						return nil, tc.first
					}
				case "":
					if tc.initial != nil {
						return nil, tc.initial
					}
				default:
					t.Fatal("unexpected login candidate")
				}
				return want, nil
			})
			calls := []bool{false}
			if tc.wantInitial {
				calls = append(calls, true)
			}
			if (err != nil) != tc.wantErr || (err == nil && got != want) || !slices.Equal(triedInitial, calls) {
				t.Fatalf("initial credential attempts = %v, success = %t", triedInitial, err == nil)
			}
		})
	}
}

func TestInitializationRejectsStaleClusterTask(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	snapshot := &v1alpha1.OBCluster{ObjectMeta: metav1.ObjectMeta{Name: "cluster", Namespace: "db", UID: "original", Generation: 1}}
	snapshot.Status.Status = clusterstatus.Bootstrapped
	snapshot.Status.OperationContext = &tasktypes.OperationContext{Name: postBootstrapFlow, Task: tCreateUsers, TaskStatus: taskstatus.Running, TargetStatus: clusterstatus.Running}
	for _, tc := range []struct {
		name   string
		change func(*v1alpha1.OBCluster)
	}{
		{"replacement", func(c *v1alpha1.OBCluster) { c.UID = "replacement" }},
		{"changed spec", func(c *v1alpha1.OBCluster) { c.Generation++ }},
		{"paused", func(c *v1alpha1.OBCluster) {
			c.Annotations = map[string]string{obconst.AnnotationsPauseReconciling: "true"}
		}},
		{"already running", func(c *v1alpha1.OBCluster) { c.Status.Status = clusterstatus.Running; c.Status.OperationContext = nil }},
		{"different task", func(c *v1alpha1.OBCluster) { c.Status.OperationContext.Task = tMaintainOBParameter }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			live := snapshot.DeepCopy()
			tc.change(live)
			// No cached client or database is available: these must reject on the authoritative read.
			m := &OBClusterManager{Ctx: context.Background(), OBCluster: snapshot, APIReader: fake.NewClientBuilder().WithScheme(scheme).WithObjects(live).Build()}
			if err := m.withInitializationClient(tCreateUsers, nil); err == nil {
				t.Fatal("stale task was allowed to initialize users")
			}
		})
	}
}

func TestStandaloneServiceResumesAfterLostCreateResult(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	m := &OBClusterManager{Ctx: ctx, Client: c, Recorder: telemetry.NewRecorder(ctx, record.NewFakeRecorder(5)), OBCluster: &v1alpha1.OBCluster{ObjectMeta: metav1.ObjectMeta{Name: "standalone", Namespace: "db", UID: "original", Annotations: map[string]string{obconst.AnnotationsMode: obconst.ModeStandalone}}}}
	if err := CreateOBClusterService(m); err != nil {
		t.Fatal(err)
	}
	if err := CreateOBClusterService(m); err != nil {
		t.Fatalf("replayed create: %v", err)
	}
	svc := &corev1.Service{}
	if err := c.Get(ctx, client.ObjectKey{Namespace: "db", Name: "standalone-standalone-svc"}, svc); err != nil {
		t.Fatal(err)
	}
	svc.OwnerReferences[0].UID = "another-cluster"
	if err := c.Update(ctx, svc); err != nil {
		t.Fatal(err)
	}
	if err := CreateOBClusterService(m); err == nil {
		t.Fatal("foreign Service was adopted")
	}
}

func TestInitializationPasswordRequiresPresentKeyOnLiveSecret(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	old := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: "db", Name: "account"}, Data: map[string][]byte{"password": []byte(t.Name())}}
	for _, tc := range []struct {
		name    string
		data    map[string][]byte
		wantErr bool
	}{
		{"missing", nil, true}, {"explicitly empty", map[string][]byte{"password": {}}, false}, {"present", map[string][]byte{"password": []byte(t.Name() + "-new")}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			live := old.DeepCopy()
			live.Data = tc.data
			m := &OBClusterManager{Ctx: context.Background(), OBCluster: &v1alpha1.OBCluster{ObjectMeta: metav1.ObjectMeta{Namespace: "db"}}, Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(old).Build(), APIReader: fake.NewClientBuilder().WithScheme(scheme).WithObjects(live).Build()}
			got, err := m.initializationPassword(old.Name)
			if (err != nil) != tc.wantErr || got != string(tc.data["password"]) {
				t.Fatal("did not honor the live password field")
			}
		})
	}
}

func TestInitializationMembersMatchAllReadyReplicas(t *testing.T) {
	var zones []v1alpha1.OBZone
	var members []model.OBServer
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("zone%d", i+1)
		var z v1alpha1.OBZone
		z.Spec.Topology.Zone = name
		for j := 0; j < 2; j++ {
			pod := fmt.Sprintf("192.0.2.%d", 1+i*2+j)
			service := fmt.Sprintf("198.51.100.%d", 1+i*2+j)
			z.Status.OBServerStatus = append(z.Status.OBServerStatus, apitypes.OBServerReplicaStatus{Server: pod, ServiceIP: service})
			members = append(members, model.OBServer{Zone: name, Ip: service, Port: obconst.RpcPort, Status: "active", StartServiceTime: 1})
		}
		zones = append(zones, z)
	}
	if !initializationMembersMatch(zones, false, members) {
		t.Fatal("six valid service-addressed database members were rejected")
	}
	for _, mutate := range []func([]model.OBServer){
		func(m []model.OBServer) { m[0].Ip = "203.0.113.99" }, func(m []model.OBServer) { m[0] = m[1] },
		func(m []model.OBServer) { m[0].Zone = m[2].Zone }, func(m []model.OBServer) { m[0].Port++ },
		func(m []model.OBServer) { m[0].StartServiceTime = 0 },
	} {
		bad := slices.Clone(members)
		mutate(bad)
		if initializationMembersMatch(zones, false, bad) {
			t.Fatal("unowned, duplicated, or inactive database member accepted")
		}
	}
	if initializationMembersMatch(zones, false, members[:3]) {
		t.Fatal("incomplete member set accepted")
	}
	zones = zones[:1]
	zones[0].Status.OBServerStatus = zones[0].Status.OBServerStatus[:1]
	members = members[:1]
	members[0].Ip = "127.0.0.1"
	if !initializationMembersMatch(zones, true, members) || initializationMembersMatch(zones, false, members) {
		t.Fatal("standalone loopback identity must be restricted to standalone mode")
	}
}

type initDataSource struct {
	*connector.OceanBaseDataSource
	driverName string
}

func (d initDataSource) DriverName() string     { return d.driverName }
func (d initDataSource) DataSourceName() string { return d.Password }

type initDriver struct {
	password, member string
	opened, closed   int
}

func (d *initDriver) Open(password string) (driver.Conn, error) {
	if d.password != password {
		return nil, &mysql.MySQLError{Number: 1045}
	}
	d.opened++
	return &initConn{driver: d}, nil
}

type initConn struct{ driver *initDriver }

func (c *initConn) Close() error                        { c.driver.closed++; return nil }
func (c *initConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unsupported") }
func (c *initConn) Begin() (driver.Tx, error)           { return nil, errors.New("unsupported") }
func (c *initConn) ExecContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}
func (c *initConn) QueryContext(_ context.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
	if strings.Contains(q, "__all_virtual_sys_parameter_stat") {
		return &initRows{[]string{"count"}, []driver.Value{int64(2)}}, nil
	}
	if strings.Contains(q, "__all_server") {
		return &initRows{[]string{"zone", "svr_ip", "svr_port", "status", "start_service_time"}, []driver.Value{"zone", c.driver.member, int64(obconst.RpcPort), "active", int64(1)}}, nil
	}
	return nil, errors.New("unexpected query")
}

type initRows struct {
	columns []string
	values  []driver.Value
}

func (r *initRows) Columns() []string { return r.columns }
func (r *initRows) Close() error      { return nil }
func (r *initRows) Next(values []driver.Value) error {
	if r.values == nil {
		return io.EOF
	}
	copy(values, r.values)
	r.values = nil
	return nil
}

var initializationDriverID atomic.Uint64

func TestInitializationAlwaysAuthenticatesAndClosesPhysicalSession(t *testing.T) {
	ctx := context.Background()
	logger := logr.Discard()
	password := rand.Text()
	d := &initDriver{password: password, member: "192.0.2.1"}
	source := initDataSource{connector.NewOceanBaseDataSource(d.member, obconst.SqlPort, obconst.RootUser, obconst.SysTenant, password, obconst.DefaultDatabase), fmt.Sprintf("init-%d", initializationDriverID.Add(1))}
	sql.Register(source.driverName, d)
	m := &OBClusterManager{Ctx: ctx, OBCluster: &v1alpha1.OBCluster{}, Logger: &logger}
	var zone v1alpha1.OBZone
	zone.Spec.Topology.Zone = "zone"
	zone.Status.OBServerStatus = []apitypes.OBServerReplicaStatus{{Server: d.member}}
	zones := []v1alpha1.OBZone{zone}
	// A matching cached pool can still ping its old authenticated connection after a password change.
	old, err := database.GetConnector(source)
	if err != nil {
		t.Fatal("open first physical session")
	}
	defer old.Close()
	d.password = rand.Text()
	if !old.IsAlive() {
		t.Fatal("expected old authenticated session to remain usable")
	}
	if _, err := openInitializationDatabase(source); err == nil {
		t.Fatal("stale credential passed fresh login")
	}
	source.Password = d.password
	failure := errors.New("operation lost its response")
	for _, tc := range []struct {
		name       string
		member     string
		actionErr  error
		wantAction bool
	}{
		{"verified", "192.0.2.1", nil, true}, {"different database members", "203.0.113.1", nil, false}, {"ambiguous action", "192.0.2.1", failure, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d.member = tc.member
			beforeOpen, beforeClose := d.opened, d.closed
			db, err := openInitializationDatabase(source)
			if err != nil {
				t.Fatal("open independent connection")
			}
			called := false
			verified, err := m.runOnInitializationDatabase(db, zones, func(_ *operation.OceanbaseOperationManager) error { called = true; return tc.actionErr })
			if called != tc.wantAction || verified != tc.wantAction || (err == nil) != (tc.wantAction && tc.actionErr == nil) || (tc.actionErr != nil && !errors.Is(err, tc.actionErr)) || d.opened != beforeOpen+1 || d.closed != beforeClose+1 {
				t.Fatalf("verified=%t action=%t opened=%d closed=%d", verified, called, d.opened-beforeOpen, d.closed-beforeClose)
			}
		})
	}
}
