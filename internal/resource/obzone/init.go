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

package obzone

import (
	"github.com/oceanbase/ob-operator/pkg/task"
)

func init() {
	// obzone
	task.GetRegistry().Register(fCreateOBZone, CreateOBZone)
	task.GetRegistry().Register(fAddOBServer, AddOBServer)
	task.GetRegistry().Register(fDeleteOBServer, DeleteOBServer)
	task.GetRegistry().Register(fPrepareOBZoneForBootstrap, PrepareOBZoneForBootstrap)
	task.GetRegistry().Register(fUpgradeOBZone, UpgradeOBZone)
	task.GetRegistry().Register(fForceUpgradeOBZone, ForceUpgradeOBZone)
	task.GetRegistry().Register(fMaintainOBZoneAfterBootstrap, MaintainOBZoneAfterBootstrap)
	task.GetRegistry().Register(fDeleteOBZoneFinalizer, DeleteOBZoneFinalizer)
	task.GetRegistry().Register(fScaleUpOBServers, ScaleUpOBServers)
	task.GetRegistry().Register(fExpandPVC, ResizePVC)
}
