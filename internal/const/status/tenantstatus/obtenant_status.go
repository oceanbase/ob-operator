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

package tenantstatus

const (
	CreatingTenant         = "creating"
	Running                = "running"
	MaintainingWhiteList   = "maintaining whitelist"
	MaintainingCharset     = "maintaining charset"
	MaintainingUnitNum     = "maintaining unit num"
	MaintainingPrimaryZone = "maintaining primary zone"
	MaintainingLocality    = "maintaining locality"
	AddingResourcePool     = "adding resource pool"
	DeletingResourcePool   = "deleting resource pool"
	MaintainingUnitConfig  = "maintaining unit config"
	MaintainingParameters  = "maintaining parameters"
	MaintainingVariables   = "maintaining variables"
	DeletingTenant         = "deleting"
	FinalizerFinished      = "finalizer finished"
	PausingReconcile       = "pausing reconcile"

	Restoring            = "restoring"
	SwitchingRole        = "switching role"
	RestoreCanceled      = "restore canceled"
	CancelingRestore     = "canceling restore"
	RestoreFailed        = "restore failed"
	CreatingEmptyStandby = "creating empty standby"
	Failed               = "failed"
)
