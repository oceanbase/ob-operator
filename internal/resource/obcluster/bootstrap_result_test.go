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
	"testing"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

func TestLostBootstrapRequiresActiveOwnedDatabaseMembers(t *testing.T) {
	wanted := map[string]map[string]bool{
		"zone1": {"192.0.2.1": true, "192.0.2.2": true},
		"zone2": {"192.0.2.3": true},
	}
	one := model.OBServer{Zone: "zone1", Ip: "192.0.2.2", Port: obconst.RpcPort, Status: "active", StartServiceTime: 1}
	two := model.OBServer{Zone: "zone2", Ip: "192.0.2.3", Port: obconst.RpcPort, Status: "active", StartServiceTime: 2}
	if !bootstrapMembersReady(wanted, []model.OBServer{two, one}) {
		t.Fatal("bootstrap may select any member of each zone")
	}
	for _, edit := range []func(*model.OBServer){
		func(m *model.OBServer) { m.Ip = "192.0.2.9" },
		func(m *model.OBServer) { *m = one },
		func(m *model.OBServer) { m.Status = "inactive" },
		func(m *model.OBServer) { m.Port = obconst.SqlPort },
		func(m *model.OBServer) { m.StartServiceTime = 0 },
	} {
		member := two
		edit(&member)
		if bootstrapMembersReady(wanted, []model.OBServer{one, member}) {
			t.Fatalf("unexpected bootstrap proof with %+v", member)
		}
	}
	if bootstrapMembersReady(wanted, []model.OBServer{one}) || bootstrapMembersReady(nil, nil) {
		t.Fatal("missing bootstrap zones cannot prove SQL completion")
	}
	m := &OBClusterManager{OBCluster: &v1alpha1.OBCluster{}}
	m.OBCluster.Status.Status = clusterstatus.New
	if m.CanReplayTask(tBootstrap) || m.ObserveLostTask(tBootstrap) == nil || m.ObserveLostTask(tCreateUsers) != nil {
		t.Fatal("lost bootstrap requires a read-only verifier")
	}
	m.OBCluster.Spec.DeploymentMode = obconst.DeploymentModeSharedStorage
	if m.ObserveLostTask(tBootstrap) != nil {
		t.Fatal("shared-storage bootstrap requires additional storage identity evidence")
	}
}
