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

package observer

import (
	ttypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

// observer flows
const (
	fPrepareOBServerForBootstrap    ttypes.FlowName = "prepare observer for bootstrap"
	fMaintainOBServerAfterBootstrap ttypes.FlowName = "maintain observer after bootstrap"
	fCreateOBServer                 ttypes.FlowName = "create observer"
	fDeleteOBServerFinalizer        ttypes.FlowName = "delete observer finalizer"
	fUpgradeOBServer                ttypes.FlowName = "upgrade observer"
	fRecoverOBServer                ttypes.FlowName = "recover observer"
	fAnnotateOBServerPod            ttypes.FlowName = "annotate observer pod"
	fScaleUpOBServer                ttypes.FlowName = "scale up observer"
	fExpandPVC                      ttypes.FlowName = "expand pvc for observer"
	fMountBackupVolume              ttypes.FlowName = "mount backup volume for observer"
)

// observer tasks
const (
	tWaitOBClusterBootstrapped    ttypes.TaskName = "wait obcluster bootstrapped"
	tCreateOBServerSvc            ttypes.TaskName = "create observer svc"
	tCreateOBPVC                  ttypes.TaskName = "create observer pvc"
	tCreateOBPod                  ttypes.TaskName = "create observer pod"
	tAnnotateOBServerPod          ttypes.TaskName = "annotate observer pod"
	tWaitOBServerReady            ttypes.TaskName = "wait observer ready"
	tStartOBServer                ttypes.TaskName = "start observer"
	tAddServer                    ttypes.TaskName = "add observer"
	tDeleteOBServerInCluster      ttypes.TaskName = "delete observer in cluster"
	tWaitOBServerDeletedInCluster ttypes.TaskName = "wait observer deleted in cluster"
	tWaitOBServerPodReady         ttypes.TaskName = "wait observer pod ready"
	tWaitOBServerActiveInCluster  ttypes.TaskName = "wait observer active in cluster"
	tUpgradeOBServerImage         ttypes.TaskName = "upgrade observer image"
	tDeletePod                    ttypes.TaskName = "delete pod"
	tWaitForPodDeleted            ttypes.TaskName = "wait for pod being deleted"
	tExpandPVC                    ttypes.TaskName = "expand pvc"
	tWaitForPVCResized            ttypes.TaskName = "wait for pvc being resized"
	tMountBackupVolume            ttypes.TaskName = "mount backup volume"
	tWaitForBackupVolumeMounted   ttypes.TaskName = "wait for backup volume to be mounted"
)
