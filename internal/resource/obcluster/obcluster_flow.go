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
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func BootstrapOBCluster() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fBootstrapOBCluster,
			Tasks:        []tasktypes.TaskName{tCheckAndCreateUserSecrets, tCreateOBZone, tWaitOBZoneBootstrapReady, tBootstrap},
			TargetStatus: clusterstatus.Bootstrapped,
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: clusterstatus.New,
			},
		},
	}
}

func MaintainOBClusterAfterBootstrap() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fMaintainOBClusterAfterBootstrap,
			Tasks:        []tasktypes.TaskName{tWaitOBZoneRunning, tCreateUsers, tMaintainOBParameter, tCreateServiceForMonitor},
			TargetStatus: clusterstatus.Running,
		},
	}
}

func AddOBZone() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fAddOBZone,
			Tasks:        []tasktypes.TaskName{tCreateOBZone, tWaitOBZoneRunning, tModifySysTenantReplica},
			TargetStatus: clusterstatus.Running,
		},
	}
}

func DeleteOBZone() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fDeleteOBZone,
			Tasks:        []tasktypes.TaskName{tModifySysTenantReplica, tDeleteOBZone, tWaitOBZoneDeleted},
			TargetStatus: clusterstatus.Running,
		},
	}
}

func ModifyOBZoneReplica() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fModifyOBZoneReplica,
			Tasks:        []tasktypes.TaskName{tModifyOBZoneReplica, tWaitOBZoneTopologyMatch, tWaitOBZoneRunning},
			TargetStatus: clusterstatus.Running,
		},
	}
}

func MaintainOBParameter() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fMaintainOBParameter,
			Tasks:        []tasktypes.TaskName{tMaintainOBParameter},
			TargetStatus: clusterstatus.Running,
		},
	}
}

func UpgradeOBCluster() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fUpgradeOBCluster,
			Tasks:        []tasktypes.TaskName{tValidateUpgradeInfo, tBackupEssentialParameters, tUpgradeCheck, tBeginUpgrade, tRollingUpgradeByZone, tFinishUpgrade, tRestoreEssentialParameters},
			TargetStatus: clusterstatus.Running,
			OnFailure: tasktypes.FailureRule{
				Strategy: strategy.Pause,
			},
		},
	}
}
