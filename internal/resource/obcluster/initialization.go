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
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/go-sql-driver/mysql"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	secretconst "github.com/oceanbase/ob-operator/internal/const/secret"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/database"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func (m *OBClusterManager) initializationReader() client.Reader {
	if m.APIReader != nil {
		return m.APIReader
	}
	return m.Client
}

// Initialization accepts an explicitly empty password but never treats a
// malformed Secret as permission to use the bootstrap credential.
func (m *OBClusterManager) initializationPassword(name string) (string, error) {
	var secret corev1.Secret
	if err := m.initializationReader().Get(m.Ctx, client.ObjectKey{Namespace: m.OBCluster.Namespace, Name: name}, &secret); err != nil {
		return "", err
	}
	value, ok := secret.Data[secretconst.PasswordKeyName]
	if !ok {
		return "", fmt.Errorf("initialization Secret %s is missing its password field", name)
	}
	return string(value), nil
}

func (m *OBClusterManager) currentInitializationTask(task tasktypes.TaskName) (*v1alpha1.OBCluster, error) {
	reader := m.initializationReader()
	live := &v1alpha1.OBCluster{}
	if err := reader.Get(m.Ctx, client.ObjectKeyFromObject(m.OBCluster), live); err != nil {
		return nil, err
	}
	op := live.Status.OperationContext
	if live.UID == "" || live.UID != m.OBCluster.UID || live.Generation != m.OBCluster.Generation || live.DeletionTimestamp != nil ||
		live.Annotations[obconst.AnnotationsPauseReconciling] == "true" || op == nil || op.Task != task ||
		(op.TaskStatus != taskstatus.Pending && op.TaskStatus != taskstatus.Running) {
		return nil, errors.New("cluster identity or initialization task changed")
	}
	migrating := live.Status.Status == clusterstatus.MigrateFromExisting && op.Name == "migrate obcluster from existing" && task == tCreateUsers
	if !migrating && (live.Status.Status != clusterstatus.Bootstrapped || op.Name != postBootstrapFlow || op.TargetStatus != clusterstatus.Running) {
		return nil, errors.New("cluster is not initializing after bootstrap")
	}
	return live, nil
}

// Only the live task that owns the cluster may initialize its accounts.
// Verify the database and run the action on the same fresh physical session.
func (m *OBClusterManager) withInitializationClient(task tasktypes.TaskName, run func(*operation.OceanbaseOperationManager, string) error) error {
	live, err := m.currentInitializationTask(task)
	if err != nil {
		return err
	}
	rootPassword, err := m.initializationPassword(live.Spec.UserSecrets.Root)
	if err != nil {
		return err
	}
	perform := func(session *operation.OceanbaseOperationManager) error {
		// The API state can change while waiting for SQL or a member to recover.
		current, err := m.currentInitializationTask(task)
		if err != nil {
			return err
		}
		if current.Status.Status != live.Status.Status {
			return errors.New("initialization phase changed while connecting")
		}
		return run(session, rootPassword)
	}
	if live.Status.Status == clusterstatus.MigrateFromExisting {
		db, err := resourceutils.GetMigrationOperationClient(m.initializationReader(), m.Logger, live, rootPassword)
		if err != nil {
			return err
		}
		return perform(db)
	}
	current := *m
	current.OBCluster = live
	zones, err := current.readyBootstrapZones(m.Ctx, serverstatus.Running)
	if err != nil {
		return err
	}
	for _, zone := range zones {
		for _, server := range zone.Status.OBServerStatus {
			if err := m.Ctx.Err(); err != nil {
				return err
			}
			addr := server.GetConnectAddr()
			db, err := connectInitializationRoot(rootPassword, task == tCreateUsers, func(password string) (*operation.OceanbaseOperationManager, error) {
				source := connector.NewOceanBaseDataSource(addr, obconst.SqlPort, obconst.RootUser, obconst.SysTenant, password, obconst.DefaultDatabase)
				return openInitializationDatabase(source)
			})
			if err != nil {
				continue
			}
			verified, err := current.runOnInitializationDatabase(db, zones, perform)
			if verified {
				return err
			}
		}
	}
	return errors.New("cannot verify a database connection to the initializing cluster")
}

// Once the action starts, propagate its result even if the response was lost;
// trying a second server here could replay an already applied change.
func (m *OBClusterManager) runOnInitializationDatabase(db *operation.OceanbaseOperationManager, zones []v1alpha1.OBZone, run func(*operation.OceanbaseOperationManager) error) (bool, error) {
	db.Logger = m.Logger
	defer db.Close()
	verified := false
	err := db.WithDedicatedSession(m.Ctx, func(session *operation.OceanbaseOperationManager) error {
		if !bootstrapDatabaseMatches(m.Ctx, session, m.OBCluster) {
			return errors.New("initialization database identity mismatch")
		}
		var members []model.OBServer
		if err := session.QueryList(m.Ctx, &members, "SELECT zone, svr_ip, svr_port, LOWER(status) AS status, start_service_time FROM oceanbase.__all_server"); err != nil {
			return err
		}
		standalone := m.OBCluster.Annotations[obconst.AnnotationsMode] == obconst.ModeStandalone
		if !initializationMembersMatch(zones, standalone, members) {
			return errors.New("initialization database membership mismatch")
		}
		verified = true
		return run(session)
	})
	return verified, err
}

func openInitializationDatabase(source database.DataSource) (*operation.OceanbaseOperationManager, error) {
	conn := database.NewConnector(source)
	if err := conn.Init(); err != nil {
		return nil, err
	}
	return operation.NewOceanbaseOperationManager(conn), nil
}

func initializationMembersMatch(zones []v1alpha1.OBZone, standalone bool, members []model.OBServer) bool {
	type address struct{ zone, ip string }
	expected := map[address]bool{}
	for _, zone := range zones {
		for _, replica := range zone.Status.OBServerStatus {
			ip := replica.GetConnectAddr()
			if standalone {
				ip = "127.0.0.1"
			}
			key := address{zone.Spec.Topology.Zone, ip}
			if expected[key] || ip == "" {
				return false
			}
			expected[key] = true
		}
	}
	if len(expected) == 0 || len(members) != len(expected) {
		return false
	}
	for _, member := range members {
		key := address{member.Zone, member.Ip}
		if !expected[key] || member.Port != obconst.RpcPort || member.Status != "active" || member.StartServiceTime <= 0 {
			return false
		}
		delete(expected, key)
	}
	return true
}

func connectInitializationRoot(password string, creatingUsers bool, connect func(string) (*operation.OceanbaseOperationManager, error)) (*operation.OceanbaseOperationManager, error) {
	manager, err := connect(password)
	// Only an explicit login rejection before root has been initialized permits
	// the bootstrap credential. Timeouts and identity failures are inconclusive.
	var rejected *mysql.MySQLError
	if creatingUsers && password != "" && errors.As(err, &rejected) && rejected.Number == 1045 {
		return connect("")
	}
	return manager, err
}
