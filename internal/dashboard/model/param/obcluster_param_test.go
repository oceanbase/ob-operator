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

package param

import (
	"testing"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
)

func TestValidateStorageMode(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*CreateOBClusterParam)
		valid  bool
	}{
		{"SS without redo", func(p *CreateOBClusterParam) {}, true},
		{"SS normal preset cache too small", func(p *CreateOBClusterParam) { p.OBServer.Storage.Data.SizeGB = 30 }, false},
		{"SS below verified cache minimum", func(p *CreateOBClusterParam) { p.OBServer.Storage.Data.SizeGB = 49 }, false},
		{"SS larger cache", func(p *CreateOBClusterParam) { p.OBServer.Storage.Data.SizeGB = 100 }, true},
		{"missing logservice", func(p *CreateOBClusterParam) { p.LogServiceRef = nil }, false},
		{"empty logservice name", func(p *CreateOBClusterParam) { p.LogServiceRef.Name = " " }, false},
		{"missing storage", func(p *CreateOBClusterParam) { p.SharedStorageInfo = nil }, false},
		{"missing bucket", func(p *CreateOBClusterParam) { p.SharedStorageInfo.BucketURL = "" }, false},
		{"inline credential", func(p *CreateOBClusterParam) { p.SharedStorageInfo.BucketURL = "s3://bucket?access_key=example" }, false},
		{"missing secret", func(p *CreateOBClusterParam) { p.SharedStorageInfo.SecretRef.Name = "" }, false},
		{"SS with redo", func(p *CreateOBClusterParam) { p.OBServer.Storage.RedoLog = &common.StorageSpec{} }, false},
		{"duplicate zones", func(p *CreateOBClusterParam) { p.Topology = append(p.Topology, p.Topology[0]) }, false},
		{"zero replica", func(p *CreateOBClusterParam) { p.Topology[0].Replicas = 0 }, false},
		{"empty topology", func(p *CreateOBClusterParam) { p.Topology = nil }, false},
		{"unknown mode", func(p *CreateOBClusterParam) { p.DeploymentMode = "typo" }, false},
		{"normal with SS fields", func(p *CreateOBClusterParam) { p.DeploymentMode = "normal" }, false},
		{"legacy normal", func(p *CreateOBClusterParam) {
			p.DeploymentMode = ""
			p.SharedStorageInfo = nil
			p.LogServiceRef = nil
			p.OBServer.Storage.RedoLog = &common.StorageSpec{}
		}, true},
		{"normal missing redo", func(p *CreateOBClusterParam) {
			p.DeploymentMode = "normal"
			p.SharedStorageInfo = nil
			p.LogServiceRef = nil
		}, false},
		{"nil observer", func(p *CreateOBClusterParam) { p.OBServer = nil }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &CreateOBClusterParam{
				DeploymentMode:    "shared_storage",
				SharedStorageInfo: &common.SharedStorageSpec{BucketURL: "s3://test", SecretRef: common.ObjectReference{Name: "object-store"}},
				LogServiceRef:     &common.ObjectReference{Name: "ls"},
				OBServer:          &OBServerSpec{Storage: &OBServerStorageSpec{Data: common.StorageSpec{SizeGB: 50}}},
				Topology:          []ZoneTopology{{Zone: "zone1", Replicas: 1}},
			}
			tc.mutate(p)
			if err := p.ValidateStorageMode(); (err == nil) != tc.valid {
				t.Fatalf("valid=%v, error=%v", tc.valid, err)
			}
		})
	}
}
