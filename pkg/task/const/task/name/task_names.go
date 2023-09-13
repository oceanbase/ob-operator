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

package name

// obcluster tasks
const (
	CreateOBZone             = "create obzone"
	DeleteOBZone             = "delete obzone"
	WaitOBZoneBootstrapReady = "wait obzone bootstrap ready"
	Bootstrap                = "bootstrap"
	CreateUsers              = "create users"
	UpdateParameter          = "update parameter"
	ModifyOBZoneReplica      = "modify obzone replica"
	WaitOBZoneRunning        = "wait obzone running"
	WaitOBZoneTopologyMatch  = "wait obzone topology match"
	WaitOBZoneDeleted        = "wait obzone deleted"
	CreateOBClusterService   = "create obcluster service"
	MaintainOBParameter      = "maintain obparameter"
	// for upgrade
	ValidateUpgradeInfo        = "validate upgrade info"
	UpgradeCheck               = "upgrade check"
	BackupEssentialParameters  = "backup essential parameters"
	BeginUpgrade               = "execute upgrade pre script"
	RollingUpgradeByZone       = "rolling upgrade by zone"
	FinishUpgrade              = "execute upgrade post script"
	RestoreEssentialParameters = "restore essential parameters"
	CreateServiceForMonitor    = "create service for monitor"
)

// obzone tasks
const (
	CreateOBServer             = "create observer"
	UpgradeOBServer            = "upgrade observer"
	WaitOBServerUpgraded       = "wait observer upgraded"
	DeleteOBServer             = "delete observer"
	DeleteAllOBServer          = "delete all observer"
	AddZone                    = "add zone"
	StartOBZone                = "start obzone"
	WaitOBServerBootstrapReady = "wait observer bootstrap ready"
	WaitOBServerRunning        = "wait observer running"
	WaitReplicaMatch           = "wait replica match"
	WaitOBServerDeleted        = "wait observer deleted"
	StopOBZone                 = "stop obzone"
	DeleteOBZoneInCluster      = "delete obzone in cluster"
	OBClusterHealthCheck       = "obcluster health check"
	OBZoneHealthCheck          = "obzone health check"
)

// observer tasks
const (
	WaitOBClusterBootstrapped    = "wait obcluster bootstrapped"
	CreateOBPVC                  = "create observer pvc"
	CreateOBPod                  = "create observer pod"
	AnnotateOBServerPod          = "annotate observer pod"
	WaitOBServerReady            = "wait observer ready"
	StartOBServer                = "start observer"
	AddServer                    = "add observer"
	DeleteOBServerInCluster      = "delete observer in cluster"
	WaitOBServerDeletedInCluster = "wait observer deleted in cluster"
	WaitOBServerPodReady         = "wait observer pod ready"
	WaitOBServerActiveInCluster  = "wait observer active in cluster"
	UpgradeOBServerImage         = "upgrade observer image"
)

// obparameter tasks
const (
	SetOBParameter = "set ob parameter"
)

const (
	CheckTenant                     = "create tenant check"
	CheckPoolAndUnitConfig          = "create pool and unit config check"
	CreateTenant                    = "create tenant"
	CreateResourcePoolAndUnitConfig = "create resource pool and unit config"
	//AddFinalizer = "add finalizer"

	// maintain tenant
	MaintainWhiteList   = "maintain white list"
	MaintainCharset     = "maintain charset"
	MaintainUnitNum     = "maintain unit num"
	MaintainLocality    = "maintain locality"
	MaintainPrimaryZone = "maintain primary zone"

	// maintain resource pool
	AddResourcePool    = "add resource pool"
	DeleteResourcePool = "delete resource pool"

	// maintain unit config
	MaintainUnitConfig = "maintain unit config"

	DeleteTenant = "delete tenant"
)
