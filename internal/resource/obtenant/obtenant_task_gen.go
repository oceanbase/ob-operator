// Code generated by go generate; DO NOT EDIT.
package obtenant

func init() {
	taskMap.Register(tCheckTenant, CheckTenant)
	taskMap.Register(tCheckPoolAndConfig, CheckPoolAndConfig)
	taskMap.Register(tCreateTenantWithClear, CreateTenantWithClear)
	taskMap.Register(tCreateResourcePoolAndConfig, CreateResourcePoolAndConfig)
	taskMap.Register(tAddPool, AddPool)
	taskMap.Register(tDeletePool, DeletePool)
	taskMap.Register(tMaintainUnitConfig, MaintainUnitConfig)
	taskMap.Register(tDeleteTenant, DeleteTenant)
	taskMap.Register(tCheckAndApplyCharset, CheckAndApplyCharset)
	taskMap.Register(tCreateEmptyStandbyTenant, CreateEmptyStandbyTenant)
	taskMap.Register(tCheckPrimaryTenantLsIntegrity, CheckPrimaryTenantLsIntegrity)
	taskMap.Register(tCreateTenantRestoreJobCR, CreateTenantRestoreJobCR)
	taskMap.Register(tWatchRestoreJobToFinish, WatchRestoreJobToFinish)
	taskMap.Register(tCancelTenantRestoreJob, CancelTenantRestoreJob)
	taskMap.Register(tUpgradeTenantIfNeeded, UpgradeTenantIfNeeded)
	taskMap.Register(tCheckAndApplyUnitNum, CheckAndApplyUnitNum)
	taskMap.Register(tCheckAndApplyWhiteList, CheckAndApplyWhiteList)
	taskMap.Register(tCheckAndApplyPrimaryZone, CheckAndApplyPrimaryZone)
	taskMap.Register(tCheckAndApplyLocality, CheckAndApplyLocality)
	taskMap.Register(tOptimizeTenantByScenario, OptimizeTenantByScenario)
	taskMap.Register(tCreateUserWithCredentialSecrets, CreateUserWithCredentialSecrets)
}
