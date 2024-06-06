// Code generated by go generate; DO NOT EDIT.
package obzone

import ttypes "github.com/oceanbase/ob-operator/pkg/task/types"

const (
	tAddZone                          ttypes.TaskName = "add zone"
	tStartOBZone                      ttypes.TaskName = "start obzone"
	tCreateOBServer                   ttypes.TaskName = "create observer"
	tDeleteOBServer                   ttypes.TaskName = "delete observer"
	tDeleteAllOBServer                ttypes.TaskName = "delete all observer"
	tWaitReplicaMatch                 ttypes.TaskName = "wait replica match"
	tWaitOBServerDeleted              ttypes.TaskName = "wait observer deleted"
	tStopOBZone                       ttypes.TaskName = "stop obzone"
	tOBClusterHealthCheck             ttypes.TaskName = "obcluster health check"
	tOBZoneHealthCheck                ttypes.TaskName = "obzone health check"
	tUpgradeOBServer                  ttypes.TaskName = "upgrade observer"
	tWaitOBServerUpgraded             ttypes.TaskName = "wait observer upgraded"
	tDeleteOBZoneInCluster            ttypes.TaskName = "delete obzone in cluster"
	tScaleOBServersVertically         ttypes.TaskName = "scale observers vertically"
	tExpandPVC                        ttypes.TaskName = "expand pvc"
	tModifyPodTemplate                ttypes.TaskName = "modify pod template"
	tDeleteLegacyOBServers            ttypes.TaskName = "delete legacy observers"
	tWaitOBServerBootstrapReady       ttypes.TaskName = "wait observer bootstrap ready"
	tWaitOBServerRunning              ttypes.TaskName = "wait observer running"
	tWaitForOBServerScalingUp         ttypes.TaskName = "wait for observer scaling up"
	tWaitForOBServerExpandingPVC      ttypes.TaskName = "wait for observer expanding pvc"
	tWaitForOBServerTemplateModifying ttypes.TaskName = "wait for observer template modifying"
	tRollingReplaceOBServers          ttypes.TaskName = "rolling replace observers"
)
