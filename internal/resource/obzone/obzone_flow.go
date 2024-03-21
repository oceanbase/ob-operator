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
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/obzone"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func genMigrateOBZoneFromExistingFlow(_ *OBZoneManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fMigrateOBZoneFromExisting,
			Tasks:        []tasktypes.TaskName{tCreateOBServer, tWaitOBServerRunning, tDeleteLegacyOBServers},
			TargetStatus: zonestatus.Running,
		},
	}
}

func genPrepareOBZoneForBootstrapFlow(_ *OBZoneManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fPrepareOBZoneForBootstrap,
			Tasks:        []tasktypes.TaskName{tCreateOBServer, tWaitOBServerBootstrapReady},
			TargetStatus: zonestatus.BootstrapReady,
		},
	}
}

func genMaintainOBZoneAfterBootstrapFlow(_ *OBZoneManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fMaintainOBZoneAfterBootstrap,
			Tasks:        []tasktypes.TaskName{tWaitOBServerRunning},
			TargetStatus: zonestatus.Running,
		},
	}
}

func genCreateOBZoneFlow(_ *OBZoneManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fCreateOBZone,
			Tasks:        []tasktypes.TaskName{tAddZone, tStartOBZone, tCreateOBServer, tWaitOBServerRunning},
			TargetStatus: zonestatus.Running,
		},
	}
}

func genAddOBServerFlow(_ *OBZoneManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fAddOBServer,
			Tasks:        []tasktypes.TaskName{tCreateOBServer, tWaitOBServerRunning},
			TargetStatus: zonestatus.Running,
		},
	}
}

func genDeleteOBServerFlow(_ *OBZoneManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fDeleteOBServer,
			Tasks:        []tasktypes.TaskName{tDeleteOBServer, tWaitReplicaMatch},
			TargetStatus: zonestatus.Running,
			OnFailure: tasktypes.FailureRule{
				Strategy: strategy.RetryFromCurrent,
			},
		},
	}
}

func genDeleteOBZoneFinalizerFlow(_ *OBZoneManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fDeleteOBZoneFinalizer,
			Tasks:        []tasktypes.TaskName{tStopOBZone, tDeleteAllOBServer, tWaitOBServerDeleted, tDeleteOBZoneInCluster},
			TargetStatus: zonestatus.FinalizerFinished,
		},
	}
}

func genUpgradeOBZoneFlow(_ *OBZoneManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fUpgradeOBZone,
			Tasks:        []tasktypes.TaskName{tOBClusterHealthCheck, tStopOBZone, tUpgradeOBServer, tWaitOBServerUpgraded, tOBZoneHealthCheck, tStartOBZone},
			TargetStatus: zonestatus.Running,
		},
	}
}

func genForceUpgradeOBZoneFlow(_ *OBZoneManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fForceUpgradeOBZone,
			Tasks:        []tasktypes.TaskName{tOBClusterHealthCheck, tUpgradeOBServer, tWaitOBServerUpgraded, tOBZoneHealthCheck},
			TargetStatus: zonestatus.Running,
		},
	}
}

func genScaleUpOBServersFlow(_ *OBZoneManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fScaleUpOBServers,
			Tasks:        []tasktypes.TaskName{tScaleUpOBServers, tWaitForOBServerScalingUp, tWaitOBServerRunning},
			TargetStatus: zonestatus.Running,
		},
	}
}

func FlowExpandPVC(_ *OBZoneManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fExpandPVC,
			Tasks:        []tasktypes.TaskName{tExpandPVC, tWaitForOBServerExpandingPVC, tWaitOBServerRunning},
			TargetStatus: zonestatus.Running,
		},
	}
}

func genMountBackupVolumeFlow(_ *OBZoneManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fMountBackupVolume,
			Tasks:        []tasktypes.TaskName{tMountBackupVolume, tWaitForOBServerMounting, tWaitOBServerRunning},
			TargetStatus: zonestatus.Running,
		},
	}
}
