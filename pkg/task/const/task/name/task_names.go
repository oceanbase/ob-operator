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
)

// obzone tasks
const (
	CreateOBServer             = "create observer"
	DeleteOBServer             = "delete observer"
	DeleteAllOBServer          = "delete all observer"
	AddZone                    = "add zone"
	StartZone                  = "start zone"
	WaitOBServerBootstrapReady = "wait observer bootstrap ready"
	WaitOBServerRunning        = "wait observer running"
	WaitReplicaMatch           = "wait replica match"
	WaitOBServerDeleted        = "wait observer deleted"
	StopOBZone                 = "stop obzone"
	DeleteOBZoneInCluster      = "delete obzone in cluster"
)

// observer tasks
const (
	WaitOBClusterBootstrapped    = "wait obcluster bootstrapped"
	CreateOBPVC                  = "create observer pvc"
	CreateOBPod                  = "create observer pod"
	WaitOBPodReady               = "wait observer pod ready"
	StartOBServer                = "start observer"
	AddServer                    = "add observer"
	DeleteOBServerInCluster      = "delete observer in cluster"
	WaitOBServerDeletedInCluster = "wait observer deleted in cluster"
)

// obparameter tasks
const (
	SetOBParameter = "set ob parameter"
)
