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

package task

import (
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
)

// register all task flows at init
func init() {
	// obcluster
	GetRegistry().Register(flowname.BootstrapOBCluster, BootstrapOBCluster)
	GetRegistry().Register(flowname.MaintainOBClusterAfterBootstrap, MaintainOBClusterAfterBootstrap)
	GetRegistry().Register(flowname.AddOBZone, AddOBZone)
	GetRegistry().Register(flowname.DeleteOBZone, DeleteOBZone)
	GetRegistry().Register(flowname.ModifyOBZoneReplica, ModifyOBZoneReplica)
	GetRegistry().Register(flowname.MaintainOBParameter, MaintainOBParameter)
	GetRegistry().Register(flowname.UpgradeOBCluster, UpgradeOBCluster)

	// obzone
	GetRegistry().Register(flowname.CreateOBZone, CreateOBZone)
	GetRegistry().Register(flowname.AddOBServer, AddOBServer)
	GetRegistry().Register(flowname.DeleteOBServer, DeleteOBServer)
	GetRegistry().Register(flowname.PrepareOBZoneForBootstrap, PrepareOBZoneForBootstrap)
	GetRegistry().Register(flowname.UpgradeOBZone, UpgradeOBZone)
	GetRegistry().Register(flowname.ForceUpgradeOBZone, ForceUpgradeOBZone)
	GetRegistry().Register(flowname.MaintainOBZoneAfterBootstrap, MaintainOBZoneAfterBootstrap)
	GetRegistry().Register(flowname.DeleteOBZoneFinalizer, DeleteOBZoneFinalizer)

	// observer
	GetRegistry().Register(flowname.CreateOBServer, CreateOBServer)
	GetRegistry().Register(flowname.PrepareOBServerForBootstrap, PrepareOBServerForBootstrap)
	GetRegistry().Register(flowname.MaintainOBServerAfterBootstrap, MaintainOBServerAfterBootstrap)
	GetRegistry().Register(flowname.DeleteOBServerFinalizer, DeleteOBServerFinalizer)
	GetRegistry().Register(flowname.UpgradeOBServer, UpgradeOBServer)
	GetRegistry().Register(flowname.RecoverOBServer, RecoverOBServer)
	GetRegistry().Register(flowname.AnnotateOBServerPod, AnnotateOBServerPod)

	// tenant-level backup
	GetRegistry().Register(flowname.PrepareBackupPolicy, PrepareBackupPolicy)
	GetRegistry().Register(flowname.StartBackupJob, StartBackupJob)
	GetRegistry().Register(flowname.StopBackupJob, StopBackupJob)
	GetRegistry().Register(flowname.MaintainRunningPolicy, MaintainRunningPolicy)
	GetRegistry().Register(flowname.PauseBackup, PauseBackup)
	GetRegistry().Register(flowname.ResumeBackup, ResumeBackup)

	// obparameter
	GetRegistry().Register(flowname.SetOBParameter, SetOBParameter)
}
