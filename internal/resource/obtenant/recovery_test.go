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

package obtenant

import (
	"testing"

	"github.com/oceanbase/ob-operator/pkg/coordinator"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func TestLostTaskReplay(t *testing.T) {
	var policy coordinator.TaskReplayPolicy = &OBTenantManager{}
	if !policy.CanReplayTask(tWatchRestoreJobToFinish) {
		t.Fatal("the persisted operation must resume waiting after a controller restart")
	}
	for _, name := range []tasktypes.TaskName{tCreateTenantRestoreJobCR, tCancelTenantRestoreJob, "unknown"} {
		if policy.CanReplayTask(name) {
			t.Errorf("unexpected replay of %q", name)
		}
	}
}
