// Code generated by go generate; DO NOT EDIT.
package obcluster

func init() {
	taskMap.Register(tWaitOBZoneTopologyMatch, WaitOBZoneTopologyMatch)
	taskMap.Register(tWaitOBZoneDeleted, WaitOBZoneDeleted)
	taskMap.Register(tModifyOBZoneReplica, ModifyOBZoneReplica)
	taskMap.Register(tDeleteOBZone, DeleteOBZone)
	taskMap.Register(tCreateOBZone, CreateOBZone)
	taskMap.Register(tBootstrap, Bootstrap)
	taskMap.Register(tCreateUsers, CreateUsers)
	taskMap.Register(tMaintainOBParameter, MaintainOBParameter)
	taskMap.Register(tValidateUpgradeInfo, ValidateUpgradeInfo)
	taskMap.Register(tUpgradeCheck, UpgradeCheck)
	taskMap.Register(tBackupEssentialParameters, BackupEssentialParameters)
	taskMap.Register(tBeginUpgrade, BeginUpgrade)
	taskMap.Register(tRollingUpgradeByZone, RollingUpgradeByZone)
	taskMap.Register(tFinishUpgrade, FinishUpgrade)
	taskMap.Register(tModifySysTenantReplica, ModifySysTenantReplica)
	taskMap.Register(tCreateServiceForMonitor, CreateServiceForMonitor)
	taskMap.Register(tRestoreEssentialParameters, RestoreEssentialParameters)
	taskMap.Register(tCheckAndCreateUserSecrets, CheckAndCreateUserSecrets)
	taskMap.Register(tCreateOBClusterService, CreateOBClusterService)
	taskMap.Register(tCheckImageReady, CheckImageReady)
	taskMap.Register(tCheckClusterMode, CheckClusterMode)
	taskMap.Register(tCheckMigration, CheckMigration)
	taskMap.Register(tScaleUpOBZones, ScaleUpOBZones)
	taskMap.Register(tExpandPVC, ExpandPVC)
	taskMap.Register(tMountBackupVolume, MountBackupVolume)
	taskMap.Register(tWaitOBZoneBootstrapReady, WaitOBZoneBootstrapReady)
	taskMap.Register(tWaitOBZoneRunning, WaitOBZoneRunning)
	taskMap.Register(tCheckEnvironment, CheckEnvironment)
	taskMap.Register(tAnnotateOBCluster, AnnotateOBCluster)
}
