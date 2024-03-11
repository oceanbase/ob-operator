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

package models

type OBZoneStatus struct {
	ZoneName string `json:"zoneName"`
	Replica  int    `json:"replica"`
	Status   string `json:"status"`
}

type CommonFields struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	UID         string `json:"uid"`
	Status      string `json:"status"`
	RunningFlow string `json:"runningFlow,omitempty"`
	RunningTask string `json:"runningTask,omitempty"`
	TaskStatus  string `json:"taskStatus,omitempty"`
}

type StorageSpec struct {
	StorageClass string `json:"storageClass"`
	StorageSize  int64  `json:"storageSize"`
}

type OBTenantResourcePool struct {
	Zone        string `json:"zone"`
	Priority    int    `json:"priority"`
	Type        string `json:"type"`
	MaxCPU      int64  `json:"maxCPU"`
	MinCPU      int64  `json:"minCPU"`
	MemorySize  int64  `json:"memorySize"`
	MaxIOPS     int    `json:"maxIOPS"`
	MinIOPS     int    `json:"minIOPS"`
	IOPSWeight  int    `json:"IOPSWeight"`
	LogDiskSize int64  `json:"logDiskSize"`
	UnitNumber  int    `json:"unitNumber"`
}

type OBCluster struct {
	ClusterName    string `json:"clusterName"`
	ClusterId      int64  `json:"clusterId"`
	ClusterMode    string `json:"clusterMode"`
	SinglePVC      bool   `json:"singlePVC"`
	IndependentPVC bool   `json:"independentPVC"`
	Image          string `json:"image"`

	CPU                    int64        `json:"cpu"`
	Memory                 int64        `json:"memory"`
	SysLogStorage          *StorageSpec `json:"sysLogStorage"`
	DataStorage            *StorageSpec `json:"dataStorage"`
	RedoLogStorage         *StorageSpec `json:"redoLogStorage"`
	ConfiguredBackupVolume bool         `json:"configuredBackupVolume"`
	ConfiguredMonitor      bool         `json:"configuredMonitor"`

	Zones []OBZoneStatus `json:"zones"`

	CommonFields `json:",inline"`
}

type OBZone struct {
	ClusterName string `json:"clusterName"`
	ClusterId   int64  `json:"clusterId"`
	ClusterCR   string `json:"clusterCR"`

	Image string `json:"image"`

	CommonFields `json:",inline"`
}

type OBServer struct {
	ClusterName string `json:"clusterName"`
	ClusterId   int64  `json:"clusterId"`
	ClusterCR   string `json:"clusterCR"`
	ZoneName    string `json:"zoneName"`

	Image         string `json:"image"`
	CNI           string `json:"cni"`
	PodPhase      string `json:"podPhase"`
	PodIPHash     string `json:"podIPHash"`
	ServiceIPHash string `json:"serviceIPHash"`

	CommonFields `json:",inline"`
}

type OBTenant struct {
	TenantName  string `json:"tenantName"`
	ClusterName string `json:"clusterName"`
	TenantRole  string `json:"tenantRole"`
	UnitNumber  int    `json:"unitNumber"`

	Topology []OBTenantResourcePool `json:"topology"`

	PrimaryTenant          string `json:"primaryTenant"`
	RestoreArchiveDestType string `json:"archiveDestType"`
	RestoreBakDataDestType string `json:"bakDataDestType"`

	CommonFields `json:",inline"`
}

type OBBackupPolicy struct {
	TenantCR   string `json:"tenantCR"`
	TenantName string `json:"tenantName"`

	ArchiveDestType            string `json:"archiveDestType"`
	ArchiveSwitchPieceInterval string `json:"archiveSwitchPieceInterval"`
	BakDataDestType            string `json:"bakDataDestType"`
	BakDataFullCrontab         string `json:"bakDataFullCrontab"`
	BakDataIncrCrontab         string `json:"bakDataIncrCrontab"`
	EncryptBakData             bool   `json:"encryptBakData"`

	CommonFields `json:",inline"`
}
