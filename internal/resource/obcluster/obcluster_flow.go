/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS,
WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED,
INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package obcluster

import (
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func genMigrateOBClusterFromExistingFlow(_ *OBClusterManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "migrate obcluster from existing",
			Tasks: []tasktypes.TaskName{
				tCheckMigration,
				tCheckImageReady,
				tCheckEnvironment,
				tCheckClusterMode,
				tCheckAndCreateNs,
				tCheckAndCreateUserSecrets,
				tCreateOBZone,
				tWaitOBZoneRunning,
				tCreateUsers,
				tMaintainOBParameter,
				tCreateOBClusterService,
				tAnnotateOBCluster,
			},
			TargetStatus: clusterstatus.Running,
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: clusterstatus.Failed,
			},
		},
	}
}

func genBootstrapOBClusterFlow(_ *OBClusterManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "bootstrap obcluster",
			Tasks: []tasktypes.TaskName{
				tCheckImageReady,
				tCheckEnvironment,
				tCheckClusterMode,
				tCheckAndCreateNs,
				tCheckAndCreateUserSecrets,
				tCreateOBZone,
				tWaitOBZoneBootstrapReady,
				tBootstrap,
			},
			TargetStatus: clusterstatus.Bootstrapped,
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: clusterstatus.Failed,
			},
		},
	}
}

func genMaintainOBClusterAfterBootstrapFlow(_ *OBClusterManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "maintain obcluster after bootstrap",
			Tasks: []tasktypes.TaskName{
				tWaitOBZoneRunning,
				tCreateUsers,
				tMaintainOBParameter,
				tCreateOBClusterService,
				tAnnotateOBCluster,
				tOptimizeClusterByScenario,
			},
			TargetStatus: clusterstatus.Running,
		},
	}
}

func genAddOBZoneFlow(_ *OBClusterManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "add obzone",
			Tasks: []tasktypes.TaskName{
				tCheckAndCreateNs,
				tCreateOBZone,
				tWaitOBZoneRunning,
				tModifySysTenantReplica,
			},
			TargetStatus: clusterstatus.Running,
		},
	}
}

func genDeleteOBZoneFlow(_ *OBClusterManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "delete obzone",
			Tasks: []tasktypes.TaskName{
				tModifySysTenantReplica,
				tDeleteOBZone,
				tWaitOBZoneDeleted,
			},
			TargetStatus: clusterstatus.Running,
		},
	}
}

func genModifyOBZoneReplicaFlow(_ *OBClusterManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "modify obzone replica",
			Tasks: []tasktypes.TaskName{
				tModifyOBZoneReplica,
				tWaitOBZoneTopologyMatch,
				tWaitOBZoneRunning,
			},
			TargetStatus: clusterstatus.Running,
		},
	}
}

func genMaintainOBParameterFlow(_ *OBClusterManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "maintain obparameter",
			Tasks: []tasktypes.TaskName{
				tMaintainOBParameter,
			},
			TargetStatus: clusterstatus.Running,
		},
	}
}

func genUpgradeOBClusterFlow(_ *OBClusterManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "upgrade obcluster",
			Tasks: []tasktypes.TaskName{
				tValidateUpgradeInfo,
				tBackupEssentialParameters,
				tUpgradeCheck,
				tBeginUpgrade,
				tRollingUpgradeByZone,
				tFinishUpgrade,
				tRestoreEssentialParameters,
			},
			TargetStatus: clusterstatus.Running,
			OnFailure: tasktypes.FailureRule{
				Strategy: strategy.Pause,
			},
		},
	}
}

func genScaleOBZonesVerticallyFlow(_ *OBClusterManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "scale obzones vertically",
			Tasks: []tasktypes.TaskName{
				tAdjustParameters,
				tScaleOBZonesVertically,
			},
			TargetStatus: clusterstatus.Running,
		},
	}
}

func genExpandPVCFlow(_ *OBClusterManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "expand pvc",
			Tasks: []tasktypes.TaskName{
				tExpandPVC,
				tWaitOBZoneRunning,
			},
			TargetStatus: clusterstatus.Running,
		},
	}
}

func genModifyServerTemplateFlow(_ *OBClusterManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "modify server template",
			Tasks: []tasktypes.TaskName{
				tModifyServerTemplate,
				tWaitOBZoneRunning,
			},
			TargetStatus: clusterstatus.Running,
		},
	}
}

func genRollingUpdateOBZonesFlow(_ *OBClusterManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "rolling update observers",
			Tasks: []tasktypes.TaskName{
				tRollingUpdateOBZones,
			},
			TargetStatus: clusterstatus.Running,
		},
	}
}
