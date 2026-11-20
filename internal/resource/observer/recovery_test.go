/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package observer

import (
	"testing"

	"github.com/oceanbase/ob-operator/pkg/coordinator"
	tt "github.com/oceanbase/ob-operator/pkg/task/types"
)

func TestObserverReplayPolicy(t *testing.T) {
	var policy coordinator.TaskReplayPolicy = &OBServerManager{}
	for _, name := range []tt.TaskName{tWaitOBServerReady, tCreateOBServerPVC} {
		if !policy.CanReplayTask(name) {
			t.Fatalf("bootstrap task %q must be recoverable", name)
		}
	}
	for _, name := range []tt.TaskName{"", "bootstrap", "create obzone", "add server", tCreateOBServerSvc, tCreateOBServerPod, "unknown"} {
		if policy.CanReplayTask(name) {
			t.Fatalf("unsafe task %q allowed", name)
		}
	}
}
