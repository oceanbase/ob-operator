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

package obcluster

import (
	"github.com/oceanbase/ob-operator/pkg/task"
)

func init() {
	// obcluster
	task.GetRegistry().Register(fBootstrapOBCluster, BootstrapOBCluster)
	task.GetRegistry().Register(fMigrateOBClusterFromExisting, MigrateOBClusterFromExisting)
	task.GetRegistry().Register(fMaintainOBClusterAfterBootstrap, MaintainOBClusterAfterBootstrap)
	task.GetRegistry().Register(fAddOBZone, AddOBZone)
	task.GetRegistry().Register(fDeleteOBZone, DeleteOBZone)
	task.GetRegistry().Register(fModifyOBZoneReplica, ModifyOBZoneReplica)
	task.GetRegistry().Register(fMaintainOBParameter, MaintainOBParameter)
	task.GetRegistry().Register(fUpgradeOBCluster, UpgradeOBCluster)
	task.GetRegistry().Register(fScaleUpOBZones, ScaleUpOBZones)
	task.GetRegistry().Register(fExpandPVC, ResizePVC)
}
