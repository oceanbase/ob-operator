/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadInspectionRoleName(t *testing.T) {
	previous := dashboardConfig
	t.Cleanup(func() { dashboardConfig = previous })
	for _, roleName := range []string{"", "dashboard-qa-cluster-role"} {
		path := filepath.Join(t.TempDir(), "config.yaml")
		if err := os.WriteFile(path, []byte("inspection:\n  clusterRoleName: "+roleName+"\n  obdiag:\n    image: obdiag:test\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := loadConfig(path); err != nil {
			t.Fatal(err)
		}
		if dashboardConfig.Inspection.ClusterRoleName != roleName || dashboardConfig.Inspection.OBDiag.Image != "obdiag:test" {
			t.Fatalf("loaded config=%+v", dashboardConfig)
		}
	}
}
