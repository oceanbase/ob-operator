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
	"github.com/oceanbase/ob-operator/pkg/task"
)

func init() {
	// observer
	task.GetRegistry().Register(fCreateOBServer, CreateOBServer)
	task.GetRegistry().Register(fPrepareOBServerForBootstrap, PrepareOBServerForBootstrap)
	task.GetRegistry().Register(fMaintainOBServerAfterBootstrap, MaintainOBServerAfterBootstrap)
	task.GetRegistry().Register(fDeleteOBServerFinalizer, DeleteOBServerFinalizer)
	task.GetRegistry().Register(fUpgradeOBServer, UpgradeOBServer)
	task.GetRegistry().Register(fRecoverOBServer, RecoverOBServer)
	task.GetRegistry().Register(fAnnotateOBServerPod, AnnotateOBServerPod)
	task.GetRegistry().Register(fAddServerInOB, AddServerInOB)
	task.GetRegistry().Register(fScaleUpOBServer, ScaleUpOBServer)
	task.GetRegistry().Register(fExpandPVC, ResizePVC)
}
