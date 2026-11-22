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

package oceanseek_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
)

func TestDistributedMinioWaitsForOwnDNSBeforeStarting(t *testing.T) {
	docs, output, err := render(t, map[string]interface{}{"minio": map[string]interface{}{"mode": "distributed"}})
	if err != nil {
		t.Fatal(err, output)
	}
	var sts appsv1.StatefulSet
	decodeScheduledObject(t, find(t, docs, "StatefulSet", "qa-minio"), &sts)
	c := sts.Spec.Template.Spec.Containers[0]
	if len(c.Command) != 2 || len(c.Args) != 1 || c.StartupProbe == nil {
		t.Fatal("distributed MinIO must gate DNS before launching")
	}
	for _, resolve := range []bool{false, true} {
		t.Run(map[bool]string{true: "stale_then_current", false: "never_current"}[resolve], func(t *testing.T) {
			dir := t.TempDir()
			scripts := map[string]string{
				"timeout": "#!/bin/sh\nshift\nexec \"$@\"\n",
				"sleep":   "#!/bin/sh\nexit 0\n",
				"getent":  "#!/bin/sh\nif [ -f \"$DNS_STATE\" ] && [ \"$RESOLVE\" = yes ]; then echo '192.0.2.2 STREAM peer'; else touch \"$DNS_STATE\"; echo '192.0.2.1 STREAM stale'; fi\n",
				"minio":   "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$STARTED\"\n",
			}
			for name, script := range scripts {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0755); err != nil {
					t.Fatal(err)
				}
			}
			started := filepath.Join(dir, "started")
			cmd := exec.Command("/bin/sh", "-ec", c.Args[0])
			yes := "no"
			if resolve {
				yes = "yes"
			}
			cmd.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"), "DNS_STATE="+filepath.Join(dir, "dns"), "STARTED="+started, "RESOLVE="+yes, "OCEANSEEK_POD_IP=192.0.2.2", "OCEANSEEK_PEER_HOST=qa-minio-0.qa-minio-peer.qa.svc", "OCEANSEEK_ENDPOINTS=http://qa-minio-{0...2}.qa-minio-peer.qa.svc:9000/data")
			out, err := cmd.CombinedOutput()
			if resolve {
				data, readErr := os.ReadFile(started)
				if err != nil || readErr != nil || !strings.Contains(string(data), "http://qa-minio-{0...2}") {
					t.Fatalf("%v %v %s", err, readErr, out)
				}
			} else {
				if err == nil || !strings.Contains(string(out), "DNS did not converge") {
					t.Fatalf("%v %s", err, out)
				}
				if _, err := os.Stat(started); !os.IsNotExist(err) {
					t.Fatal("must not start with stale DNS")
				}
			}
		})
	}
}
