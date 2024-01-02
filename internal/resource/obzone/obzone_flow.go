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
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func PrepareOBZoneForBootstrap() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fPrepareOBZoneForBootstrap,
			Tasks:        []tasktypes.TaskName{tCreateOBServer, tWaitOBServerBootstrapReady},
			TargetStatus: zonestatus.BootstrapReady,
		},
	}
}

func MaintainOBZoneAfterBootstrap() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fMaintainOBZoneAfterBootstrap,
			Tasks:        []tasktypes.TaskName{tWaitOBServerRunning},
			TargetStatus: zonestatus.Running,
		},
	}
}

func CreateOBZone() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fCreateOBZone,
			Tasks:        []tasktypes.TaskName{tAddZone, tStartOBZone, tCreateOBServer, tWaitOBServerRunning},
			TargetStatus: zonestatus.Running,
		},
	}
}

func AddOBServer() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fAddOBServer,
			Tasks:        []tasktypes.TaskName{tCreateOBServer, tWaitOBServerRunning},
			TargetStatus: zonestatus.Running,
		},
	}
}

func DeleteOBServer() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fDeleteOBServer,
			Tasks:        []tasktypes.TaskName{tDeleteOBServer, tWaitReplicaMatch},
			TargetStatus: zonestatus.Running,
		},
	}
}

func DeleteOBZoneFinalizer() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fDeleteOBZoneFinalizer,
			Tasks:        []tasktypes.TaskName{tStopOBZone, tDeleteAllOBServer, tWaitOBServerDeleted, tDeleteOBZoneInCluster},
			TargetStatus: zonestatus.FinalizerFinished,
		},
	}
}

func UpgradeOBZone() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fUpgradeOBZone,
			Tasks:        []tasktypes.TaskName{tOBClusterHealthCheck, tStopOBZone, tUpgradeOBServer, tWaitOBServerUpgraded, tOBZoneHealthCheck, tStartOBZone},
			TargetStatus: zonestatus.Running,
		},
	}
}

func ForceUpgradeOBZone() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fForceUpgradeOBZone,
			Tasks:        []tasktypes.TaskName{tOBClusterHealthCheck, tUpgradeOBServer, tWaitOBServerUpgraded, tOBZoneHealthCheck},
			TargetStatus: zonestatus.Running,
		},
	}
}

func ScaleUpOBServers() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fScaleUpOBServers,
			Tasks:        []tasktypes.TaskName{tScaleUpOBServers, tWaitForOBServerScalingUp, tWaitOBServerRunning},
			TargetStatus: zonestatus.Running,
		},
	}
}
