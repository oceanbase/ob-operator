// Code generated by go generate; DO NOT EDIT.
package obcluster

func init() {
	flowMap[fMigrateOBClusterFromExisting] = MigrateOBClusterFromExisting
	flowMap[fBootstrapOBCluster] = BootstrapOBCluster
	flowMap[fMaintainOBClusterAfterBootstrap] = MaintainOBClusterAfterBootstrap
	flowMap[fAddOBZone] = AddOBZone
	flowMap[fDeleteOBZone] = FlowDeleteOBZone
	flowMap[fModifyOBZoneReplica] = FlowModifyOBZoneReplica
	flowMap[fMaintainOBParameter] = FlowMaintainOBParameter
	flowMap[fUpgradeOBCluster] = UpgradeOBCluster
	flowMap[fScaleUpOBZones] = FlowScaleUpOBZones
	flowMap[fExpandPVC] = FlowExpandPVC
	flowMap[fMountBackupVolume] = FlowMountBackupVolume
}
