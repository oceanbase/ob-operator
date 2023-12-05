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
	ttypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

// obzone flows
const (
	fPrepareOBZoneForBootstrap    ttypes.FlowName = "prepare obzone for bootstrap"
	fMaintainOBZoneAfterBootstrap ttypes.FlowName = "maintain obzone after bootstrap"
	fAddOBServer                  ttypes.FlowName = "add observer"
	fDeleteOBServer               ttypes.FlowName = "delete observer"
	fUpgradeOBZone                ttypes.FlowName = "upgrade obzone"
	fForceUpgradeOBZone           ttypes.FlowName = "force upgrade obzone"
	fCreateOBZone                 ttypes.FlowName = "create obzone"
	fDeleteOBZoneFinalizer        ttypes.FlowName = "delete obzone finalizer"
)

// obzone tasks
const (
	tCreateOBServer             ttypes.TaskName = "create observer"
	tUpgradeOBServer            ttypes.TaskName = "upgrade observer"
	tWaitOBServerUpgraded       ttypes.TaskName = "wait observer upgraded"
	tDeleteOBServer             ttypes.TaskName = "delete observer"
	tDeleteAllOBServer          ttypes.TaskName = "delete all observer"
	tAddZone                    ttypes.TaskName = "add zone"
	tStartOBZone                ttypes.TaskName = "start obzone"
	tWaitOBServerBootstrapReady ttypes.TaskName = "wait observer bootstrap ready"
	tWaitOBServerRunning        ttypes.TaskName = "wait observer running"
	tWaitReplicaMatch           ttypes.TaskName = "wait replica match"
	tWaitOBServerDeleted        ttypes.TaskName = "wait observer deleted"
	tStopOBZone                 ttypes.TaskName = "stop obzone"
	tDeleteOBZoneInCluster      ttypes.TaskName = "delete obzone in cluster"
	tOBClusterHealthCheck       ttypes.TaskName = "obcluster health check"
	tOBZoneHealthCheck          ttypes.TaskName = "obzone health check"
)
