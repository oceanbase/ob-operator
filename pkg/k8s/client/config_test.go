/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package client

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOutsideClusterLoadsMergedKubeconfig(t *testing.T) {
	dir := t.TempDir()
	contextFile := filepath.Join(dir, "contexts")
	clusterFile := filepath.Join(dir, "clusters")
	if err := os.WriteFile(contextFile, []byte(`apiVersion: v1
kind: Config
current-context: qa
contexts:
- name: qa
  context:
    cluster: target
`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(clusterFile, []byte(`apiVersion: v1
kind: Config
clusters:
- name: target
  cluster:
    server: https://pool.invalid:6443
`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KUBECONFIG", contextFile+string(os.PathListSeparator)+clusterFile)
	config, err := GetConfigOutsideCluster()
	if err != nil {
		t.Fatal(err)
	}
	if config.Host != "https://pool.invalid:6443" {
		t.Fatalf("loaded host %q", config.Host)
	}
}
