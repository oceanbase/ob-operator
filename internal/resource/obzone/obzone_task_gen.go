// Code generated by go generate; DO NOT EDIT.
package obzone

func init() {
	taskMap.Register(tAddZone, AddZone)
	taskMap.Register(tStartOBZone, StartOBZone)
	taskMap.Register(tCreateOBServer, CreateOBServer)
	taskMap.Register(tDeleteOBServer, DeleteOBServer)
	taskMap.Register(tDeleteAllOBServer, DeleteAllOBServer)
	taskMap.Register(tWaitReplicaMatch, WaitReplicaMatch)
	taskMap.Register(tWaitOBServerDeleted, WaitOBServerDeleted)
	taskMap.Register(tStopOBZone, StopOBZone)
	taskMap.Register(tOBClusterHealthCheck, OBClusterHealthCheck)
	taskMap.Register(tOBZoneHealthCheck, OBZoneHealthCheck)
	taskMap.Register(tUpgradeOBServer, UpgradeOBServer)
	taskMap.Register(tWaitOBServerUpgraded, WaitOBServerUpgraded)
	taskMap.Register(tDeleteOBZoneInCluster, DeleteOBZoneInCluster)
	taskMap.Register(tScaleUpOBServers, ScaleUpOBServers)
	taskMap.Register(tExpandPVC, ExpandPVC)
	taskMap.Register(tMountBackupVolume, MountBackupVolume)
	taskMap.Register(tDeleteLegacyOBServers, DeleteLegacyOBServers)
	taskMap.Register(tWaitOBServerBootstrapReady, WaitOBServerBootstrapReady)
	taskMap.Register(tWaitOBServerRunning, WaitOBServerRunning)
	taskMap.Register(tWaitForOBServerScalingUp, WaitForOBServerScalingUp)
	taskMap.Register(tWaitForOBServerExpandingPVC, WaitForOBServerExpandingPVC)
	taskMap.Register(tWaitForOBServerMounting, WaitForOBServerMounting)
	taskMap.Register(tRollingUpdateOBServers, RollingUpdateOBServers)
}
