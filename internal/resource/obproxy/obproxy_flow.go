/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OR ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package obproxy

import (
	proxystatus "github.com/oceanbase/ob-operator/internal/const/status/obproxy"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func genCreateOBProxyFlow(_ *OBProxyManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "create obproxy",
			Tasks: []tasktypes.TaskName{
				tCopyProxyROSecret,
				tCreateOBProxyConfigMap,
				tCreateOBProxyService,
				tCreateOBProxyDeployment,
				tWaitOBProxyReady,
			},
			TargetStatus: proxystatus.Running,
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: proxystatus.Failed,
			},
		},
	}
}

func genUpdateOBProxyFlow(_ *OBProxyManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "update obproxy",
			Tasks: []tasktypes.TaskName{
				tUpdateOBProxyConfigMap,
				tUpdateOBProxyDeployment,
				tWaitOBProxyReady,
			},
			TargetStatus: proxystatus.Running,
		},
	}
}

func genScaleOBProxyFlow(_ *OBProxyManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "scale obproxy",
			Tasks: []tasktypes.TaskName{
				tScaleOBProxyDeployment,
				tWaitOBProxyReady,
			},
			TargetStatus: proxystatus.Running,
		},
	}
}

func genDeleteOBProxyFlow(_ *OBProxyManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name: "delete obproxy",
			Tasks: []tasktypes.TaskName{
				tDeleteOBProxyDeployment,
				tDeleteOBProxyService,
				tDeleteOBProxyConfigMap,
				tDeleteOBProxySecrets,
			},
			TargetStatus: proxystatus.FinalizerFinished,
			OnFailure: tasktypes.FailureRule{
				Strategy: strategy.StartOver,
			},
		},
	}
}
