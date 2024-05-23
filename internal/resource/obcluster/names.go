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
	ttypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

// obcluster flows
const (
	fMigrateOBClusterFromExisting    ttypes.FlowName = "migrate obcluster from existing"
	fBootstrapOBCluster              ttypes.FlowName = "bootstrap obcluster"
	fMaintainOBClusterAfterBootstrap ttypes.FlowName = "maintain obcluster after bootstrap"
	fAddOBZone                       ttypes.FlowName = "add obzone"
	fDeleteOBZone                    ttypes.FlowName = "delete obzone"
	fModifyOBZoneReplica             ttypes.FlowName = "modify obzone replica"
	fUpgradeOBCluster                ttypes.FlowName = "upgrade ob cluster"
	fMaintainOBParameter             ttypes.FlowName = "maintain ob parameter"
	fDeleteOBClusterFinalizer        ttypes.FlowName = "delete obcluster finalizer"
	fScaleUpOBZones                  ttypes.FlowName = "scale up obzones"
	fExpandPVC                       ttypes.FlowName = "expand pvc for obcluster"
	fMountBackupVolume               ttypes.FlowName = "mount backup volume for obcluster"
)

// obcluster tasks
const (
	tCheckMigration            ttypes.TaskName = "check before migration"
	tCheckImageReady           ttypes.TaskName = "check image ready"
	tCheckClusterMode          ttypes.TaskName = "check cluster mode"
	tCheckAndCreateUserSecrets ttypes.TaskName = "check and create user secrets"
	tCreateOBZone              ttypes.TaskName = "create obzone"
	tDeleteOBZone              ttypes.TaskName = "delete obzone"
	tWaitOBZoneBootstrapReady  ttypes.TaskName = "wait obzone bootstrap ready"
	tBootstrap                 ttypes.TaskName = "bootstrap"
	tCreateUsers               ttypes.TaskName = "create users"
	tUpdateParameter           ttypes.TaskName = "update parameter"
	tModifyOBZoneReplica       ttypes.TaskName = "modify obzone replica"
	tModifySysTenantReplica    ttypes.TaskName = "modify sys tenant replica"
	tWaitOBZoneRunning         ttypes.TaskName = "wait obzone running"
	tWaitOBZoneTopologyMatch   ttypes.TaskName = "wait obzone topology match"
	tWaitOBZoneDeleted         ttypes.TaskName = "wait obzone deleted"
	tCreateOBClusterService    ttypes.TaskName = "create obcluster service"
	tMaintainOBParameter       ttypes.TaskName = "maintain obparameter"
	// for upgrade
	tValidateUpgradeInfo        ttypes.TaskName = "validate upgrade info"
	tUpgradeCheck               ttypes.TaskName = "upgrade check"
	tBackupEssentialParameters  ttypes.TaskName = "backup essential parameters"
	tBeginUpgrade               ttypes.TaskName = "execute upgrade pre script"
	tRollingUpgradeByZone       ttypes.TaskName = "rolling upgrade by zone"
	tFinishUpgrade              ttypes.TaskName = "execute upgrade post script"
	tRestoreEssentialParameters ttypes.TaskName = "restore essential parameters"
	tCreateServiceForMonitor    ttypes.TaskName = "create service for monitor"
	tScaleUpOBZones             ttypes.TaskName = "scale up obzones"
	tExpandPVC                  ttypes.TaskName = "expand pvc"
	tMountBackupVolume          ttypes.TaskName = "mount backup volume"
	tCheckEnvironment           ttypes.TaskName = "check environment"
	tAnnotateOBCluster          ttypes.TaskName = "annotate obcluster"
)
