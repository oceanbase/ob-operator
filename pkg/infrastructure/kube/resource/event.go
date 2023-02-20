/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package resource

// Container event reason list
const (
	CreatedContainer        = "Created"
	FailedToCreateContainer = "Failed"
	StartedContainer        = "Started"
	FailedToStartContainer  = "Failed"
	BackOffStartContainer   = "BackOff"
	KillingContainer        = "Killing"
	ExceededGracePeriod     = "ExceededGracePeriod"
)

// Pod event reason list
const (
	CreatedPod        = "CreatedPod"
	UpatedPod         = "UpatedPod"
	FailedToCreatePod = "FailedCreatePod"
	DeletedPod        = "DeletedPod"
	FailedToKillPod   = "FailedKillPod"
)

// Storage event reason list
const (
	CreatedPVC        = "CreatedPVC"
	FailedToCreatePVC = "FailedCreatePVC"
	DeletedPVC        = "DeletedPVC"
	FailedToDeletePVC = "FailedDeletePVC"
)

// Secret event reason list
const (
	CreatedSecret        = "CreatedSecret"
	FailedToCreateSecret = "FailedCreateSecret"
	DeletedSecret        = "DeletedSecret"
	FailedToDeleteSecret = "FailedDeleteSecret"
)

// Service event reason list
const (
	CreatedService        = "CreatedService"
	FailedToCreateService = "FailedCreateService"
	DeletedService        = "DeletedService"
	FailedToDeleteService = "FailedDeleteService"
)

// StatefulApp event reason list
const (
	CreatedStatefulApp        = "CreatedStatefulApp"
	FailedToCreateStatefulApp = "FailedCreateStatefulApp"
	DeletedStatefulApp        = "DeletedPVC"
	FailedToDeleteStatefulApp = "FailedDeleteStatefulApp"
)

// OBCluster event reason list
const (
	CreatedOBCluster        = "CreatedOBCluster"
	FailedToCreateOBCluster = "FailedCreateOBCluster"
	DeletedOBCluster        = "DeletedOBCluster"
	FailedToDeleteOBCluster = "FailedDeleteOBCluster"
)

// Storage event reason list
const (
	CreatedRootService        = "CreatedRootService"
	FailedToCreateRootService = "FailedCreateRootService"
	DeletedRootService        = "DeletedRootService"
	FailedToDeleteRootService = "FailedDeleteRootService"
)

// Backup event reason list
const (
	CreatedBackup        = "CreatedBackup"
	FailedToCreateBackup = "FailedCreateBackup"
	DeletedBackup        = "DeletedBackup"
	FailedToDeleteBackup = "FailedDeleteBackup"
)

// Restore event reason list
const (
	CreatedRestore        = "CreatedRestore"
	FailedToCreateRestore = "FailedCreateRestore"
	DeletedRestore        = "DeletedRestore"
	FailedToDeleteRestore = "FailedDeleteRestore"
)

// Storage event reason list
// OBZone event reason list
const (
	CreatedOBZone        = "CreatedOBZone"
	FailedToCreateOBZone = "FailedCreateOBZone"
	DeletedOBZone        = "DeletedOBZone"
	FailedToDeleteOBZone = "FailedDeleteOBZone"
)

// Tenant event reason list
const (
	CreatedTenant        = "CreatedTenant"
	FailedToCreateTenant = "FailedCreateTenant"
	DeletedTenant        = "DeletedTenant"
	FailedToDeleteTenant = "FailedDeleteTenant"
)

// Job event reason list
const (
	CreatedJob        = "CreatedJob"
	FailedToCreateJob = "FailedCreateJob"
	DeletedJob        = "DeletedJob"
	FailedToKillJob   = "FailedKillJob"
)

// Probe event reason list
const (
	ContainerUnhealthy = "Unhealthy"
)
