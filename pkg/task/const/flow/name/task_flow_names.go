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

package flow

// obcluster flows
const (
	BootstrapOBCluster              = "bootstrap obcluster"
	MaintainOBClusterAfterBootstrap = "maintain obcluster after bootstrap"
	AddOBZone                       = "add obzone"
	DeleteOBZone                    = "delete obzone"
	ModifyOBZoneReplica             = "modify obzone replica"
	UpgradeOBCluster                = "upgrade ob cluster"
	MaintainOBParameter             = "maintain ob parameter"
	DeleteOBClusterFinalizer        = "delete obcluster finalizer"
)

// obzone flows
const (
	PrepareOBZoneForBootstrap    = "prepare obzone for bootstrap"
	MaintainOBZoneAfterBootstrap = "maintain obzone after bootstrap"
	AddOBServer                  = "add observer"
	DeleteOBServer               = "delete observer"
	UpgradeOBZone                = "upgrade obzone"
	ForceUpgradeOBZone           = "force upgrade obzone"
	CreateOBZone                 = "create obzone"
	DeleteOBZoneFinalizer        = "delete obzone finalizer"
)

// observer flows
const (
	PrepareOBServerForBootstrap    = "prepare observer for bootstrap"
	MaintainOBServerAfterBootstrap = "maintain observer after bootstrap"
	CreateOBServer                 = "create observer"
	DeleteOBServerFinalizer        = "delete observer finalizer"
	UpgradeOBServer                = "upgrade observer"
	RecoverOBServer                = "recover observer"
	AnnotateOBServerPod            = "annotate observer pod"
)

// obparameter flows
const (
	SetOBParameter = "set ob parameter"
)

// tenant-level backup
const (
	PrepareBackupPolicy = "prepare backup policy"
	StartBackupJob      = "start backup job"
	StopBackupJob       = "stop backup job"
	MaintainCrontab     = "maintain crontab"
)
