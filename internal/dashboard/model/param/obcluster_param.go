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

package param

import (
	"fmt"
	"strings"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
)

type ZoneTopology struct {
	Zone         string                  `json:"zone" binding:"required"`
	Replicas     int                     `json:"replicas" binding:"required"`
	K8sCluster   string                  `json:"k8sCluster,omitempty"`
	NodeSelector []common.KVPair         `json:"nodeSelector,omitempty"`
	Tolerations  []common.TolerationSpec `json:"tolerations,omitempty"`
	Affinities   []common.AffinitySpec   `json:"affinities,omitempty"`
}

type OBServerStorageSpec struct {
	Data    common.StorageSpec  `json:"data" binding:"required"`
	RedoLog *common.StorageSpec `json:"redoLog,omitempty"`
	Log     common.StorageSpec  `json:"log" binding:"required"`
}

type MonitorStorageSpec struct {
	Config common.StorageSpec `json:"config"`
}

type OBServerSpec struct {
	Image    string               `json:"image" binding:"required"`
	Resource common.ResourceSpec  `json:"resource" binding:"required"`
	Storage  *OBServerStorageSpec `json:"storage" binding:"required"`
}

type MonitorSpec struct {
	Image    string              `json:"image" binding:"required"`
	Resource common.ResourceSpec `json:"resource" binding:"required"`
}

type NFSVolumeSpec struct {
	Address string `json:"address"`
	Path    string `json:"path"`
}

type CreateOBClusterParam struct {
	// DeploymentMode is independent of the networking mode (NORMAL/SERVICE/STANDALONE).
	DeploymentMode    string                    `json:"deploymentMode,omitempty" enums:"normal,shared_storage"`
	SharedStorageInfo *common.SharedStorageSpec `json:"sharedStorageInfo,omitempty"`
	LogServiceRef     *common.ObjectReference   `json:"logServiceRef,omitempty"`
	Namespace         string                    `json:"namespace" binding:"required"`
	Name              string                    `json:"name" binding:"required"`
	ClusterName       string                    `json:"clusterName" binding:"required"`
	ClusterId         int64                     `json:"clusterId" binding:"required"`
	RootPassword      string                    `json:"rootPassword" binding:"required"`
	ProxyroPassword   string                    `json:"proxyroPassword"`
	Topology          []ZoneTopology            `json:"topology" binding:"required"`
	OBServer          *OBServerSpec             `json:"observer" binding:"required"`
	Monitor           *MonitorSpec              `json:"monitor"`
	Parameters        []common.KVPair           `json:"parameters"`
	BackupVolume      *NFSVolumeSpec            `json:"backupVolume"`
	Mode              common.ClusterMode        `json:"mode" binding:"required"`

	// Enum: express_oltp, express_oltp, olap, kv, htap, express_oltp_perf
	Scenario           string `json:"scenario" binding:"required"`
	DeletionProtection bool   `json:"deletionProtection"`
	PvcIndependent     bool   `json:"pvcIndependent"`
}

// ValidateStorageMode runs before creating password Secrets. Older clients may
// omit deploymentMode; they retain the normal deployment behavior.
func (p *CreateOBClusterParam) ValidateStorageMode() error {
	if p.DeploymentMode != "" && p.DeploymentMode != "normal" && p.DeploymentMode != "shared_storage" {
		return fmt.Errorf("deploymentMode must be normal or shared_storage")
	}
	if p.OBServer == nil || p.OBServer.Storage == nil {
		return fmt.Errorf("observer.storage is required")
	}
	if p.DeploymentMode != "shared_storage" {
		if p.SharedStorageInfo != nil || p.LogServiceRef != nil {
			return fmt.Errorf("sharedStorageInfo and logServiceRef require shared_storage deploymentMode")
		}
		if p.OBServer.Storage.RedoLog == nil {
			return fmt.Errorf("observer.storage.redoLog is required for normal deploymentMode")
		}
		return nil
	}
	if p.LogServiceRef == nil || strings.TrimSpace(p.LogServiceRef.Name) == "" {
		return fmt.Errorf("logServiceRef.name is required for shared_storage deploymentMode")
	}
	if p.SharedStorageInfo == nil || strings.TrimSpace(p.SharedStorageInfo.BucketURL) == "" || strings.TrimSpace(p.SharedStorageInfo.SecretRef.Name) == "" {
		return fmt.Errorf("sharedStorageInfo.bucketURL and secretRef.name are required for shared_storage deploymentMode")
	}
	if p.OBServer.Storage.RedoLog != nil {
		return fmt.Errorf("observer.storage.redoLog must be omitted for shared_storage deploymentMode")
	}
	// The 30 GiB normal preset initializes only 6 GiB of cache, which fails
	// SS startup (system cache allocation plus reserved space). Use the
	// conservative minimum verified with oceanbase-ai 4.6.2.0.
	if p.OBServer.Storage.Data.SizeGB < 50 {
		return fmt.Errorf("shared_storage requires at least 50 GiB of local data cache in this Dashboard")
	}
	seen := make(map[string]bool)
	if len(p.Topology) == 0 {
		return fmt.Errorf("topology must contain at least one zone")
	}
	for _, zone := range p.Topology {
		if strings.TrimSpace(zone.Zone) == "" || seen[zone.Zone] || zone.Replicas < 1 {
			return fmt.Errorf("topology must contain unique non-empty zones with replicas >= 1")
		}
		seen[zone.Zone] = true
	}
	return nil
}

type UpgradeOBClusterParam struct {
	Image string `json:"image" binding:"required"`
}

type ScaleOBServerParam struct {
	Replicas int `json:"replicas" binding:"required"`
}

type K8sObjectIdentity struct {
	Namespace string `json:"namespace" uri:"namespace" binding:"required"`
	Name      string `json:"name" uri:"name" binding:"required"`
}

type OBZoneIdentity struct {
	Namespace  string `json:"namespace" uri:"namespace" binding:"required"`
	Name       string `json:"name" uri:"name" binding:"required"`
	OBZoneName string `json:"obzoneName" uri:"obzoneName" binding:"required"`
}

type PatchOBClusterParam struct {
	Resource           *common.ResourceSpec `json:"resource"`
	Storage            *OBServerStorageSpec `json:"storage"`
	Monitor            *MonitorSpec         `json:"monitor"`
	RemoveMonitor      bool                 `json:"removeMonitor"`
	BackupVolume       *NFSVolumeSpec       `json:"backupVolume"`
	RemoveBackupVolume bool                 `json:"removeBackupVolume"`

	// Replace all parameters
	Parameters []common.KVPair `json:"parameters,omitempty"`
	// Add or modify some parameters
	ModifiedParameters []common.KVPair `json:"modifiedParameters,omitempty"`
	// Delete some parameters
	DeletedParameters        []string `json:"deletedParameters,omitempty"`
	AddDeletionProtection    bool     `json:"addDeletionProtection"`
	RemoveDeletionProtection bool     `json:"removeDeletionProtection"`
}

type RestartOBServersParam v1alpha1.RestartOBServersConfig

type DeleteOBServersParam v1alpha1.DeleteOBServersConfig
