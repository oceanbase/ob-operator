/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:

	http://license.coscl.org.cn/MulanPSL2

THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/
package cluster

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	modelcommon "github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	param "github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewCreateOptions() *CreateOptions {
	return &CreateOptions{
		OBServer: &param.OBServerSpec{
			Storage: &param.OBServerStorageSpec{},
		},
		Parameters: make([]modelcommon.KVPair, 0),
		Zones:      make(map[string]string),
		Topology:   make([]param.ZoneTopology, 0),
	}
}

type CreateOptions struct {
	Namespace    string               `json:"namespace"`
	Name         string               `json:"name"`
	ClusterName  string               `json:"clusterName"`
	ClusterId    int64                `json:"clusterId"`
	RootPassword string               `json:"rootPassword"`
	Topology     []param.ZoneTopology `json:"topology"`
	OBServer     *param.OBServerSpec  `json:"observer"`
	Monitor      *param.MonitorSpec   `json:"monitor"`
	Parameters   []modelcommon.KVPair `json:"parameters"`
	BackupVolume *param.NFSVolumeSpec `json:"backupVolume"`
	Zones        map[string]string    `json:"zones"`
	Mode         string               `json:"mode"`
}

// Parse Cli args and set options
func (o *CreateOptions) Parse() error {
	if !CheckResourceName(o.Name) {
		return fmt.Errorf("invalid resource name in k8s: %s", o.Name)
	}
	for zoneName, replicaStr := range o.Zones {
		replica, err := strconv.Atoi(replicaStr)
		if err != nil {
			return fmt.Errorf("invalid value for zone %s: %s", zoneName, replicaStr)
		}
		// 添加到ZoneTopology
		o.Topology = append(o.Topology, param.ZoneTopology{
			Zone:         zoneName,
			Replicas:     replica,
			NodeSelector: make([]common.KVPair, 0),
			Tolerations:  make([]common.KVPair, 0),
			Affinities:   make([]common.AffinitySpec, 0),
		})
	}
	return nil
}

// Validate cli args
func (o *CreateOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("cluster name is required")
	}
	return nil
}

// Complete command
func (o *CreateOptions) Complete() {
	if o.ClusterId == 0 {
		o.ClusterId = time.Now().Unix() % 4294901759
	}
	if o.RootPassword == "" {
		o.RootPassword = generateRandomPassword()
	}
	if o.Name == "" {
		o.Name = o.ClusterName
	}
	return
}

// AddZoneFlags adds the zone-related flags to the command.
func (o *CreateOptions) AddZoneFlags(cmd *cobra.Command) {
	zoneFlags := pflag.NewFlagSet("zone", pflag.ContinueOnError)
	zoneFlags.StringToStringVarP(&o.Zones, "zones", "z", map[string]string{"z1": "1"}, "The zones of the cluster in the format 'zone=value', multiple values can be provided separated by commas")
	cmd.Flags().AddFlagSet(zoneFlags)
}

// AddBaseFlags adds the base flags to the command.
func (o *CreateOptions) AddBaseFlags(cmd *cobra.Command) {
	baseFlags := cmd.Flags()
	baseFlags.StringVar(&o.Name, "name", "", "The name in k8s")
	baseFlags.StringVar(&o.Namespace, "namespace", "default", "The namespace of the cluster")
	baseFlags.Int64Var(&o.ClusterId, "id", 0, "The id of the cluster")
	baseFlags.StringVar(&o.RootPassword, "root-password", "", "The root password of the cluster")
	baseFlags.StringVar(&o.Mode, "mode", "", "The mode of the cluster")
}

// AddObserverFlags adds the observer-related flags to the command.
func (o *CreateOptions) AddObserverFlags(cmd *cobra.Command) {
	observerFlags := pflag.NewFlagSet("observer", pflag.ContinueOnError)
	observerFlags.StringVar(&o.OBServer.Image, "image", "oceanbase/oceanbase-cloud-native:4.2.1.6-106000012024042515", "The image of the observer")
	observerFlags.Int64Var(&o.OBServer.Resource.Cpu, "cpu", 2, "The cpu of the observer")
	observerFlags.Int64Var(&o.OBServer.Resource.MemoryGB, "memory", 10, "The memory of the observer")
	observerFlags.StringVar(&o.OBServer.Storage.Data.StorageClass, "data-storage-class", "local-path", "The storage class of the data storage")
	observerFlags.StringVar(&o.OBServer.Storage.RedoLog.StorageClass, "redo-log-storage-class", "local-path", "The storage class of the redo log storage")
	observerFlags.StringVar(&o.OBServer.Storage.Log.StorageClass, "log-storage-class", "local-path", "The storage class of the log storage")
	observerFlags.Int64Var(&o.OBServer.Storage.Data.SizeGB, "data-storage-size", 50, "The size of the data storage")
	observerFlags.Int64Var(&o.OBServer.Storage.RedoLog.SizeGB, "redo-log-storage-size", 50, "The size of the redo log storage")
	observerFlags.Int64Var(&o.OBServer.Storage.Log.SizeGB, "log-storage-size", 20, "The size of the log storage")
	cmd.Flags().AddFlagSet(observerFlags)
}

// AddMonitorFlags adds the monitor-related flags to the command.
func (o *CreateOptions) AddMonitorFlags(cmd *cobra.Command) {
	monitorFlags := pflag.NewFlagSet("monitor", pflag.ContinueOnError)
	monitorFlags.StringVar(&o.Monitor.Image, "monitor-image", "oceanbase/obagent:4.2.1-100000092023101717", "The image of the monitor")
	monitorFlags.Int64Var(&o.Monitor.Resource.Cpu, "monitor-cpu", 1, "The cpu of the monitor")
	monitorFlags.Int64Var(&o.Monitor.Resource.MemoryGB, "monitor-memory", 1, "The memory of the monitor")
	cmd.Flags().AddFlagSet(monitorFlags)
}

// AddBackupVolumeFlags adds the backup-volume-related flags to the command.
func (o *CreateOptions) AddBackupVolumeFlags(cmd *cobra.Command) {
	backupVolumeFlags := pflag.NewFlagSet("backup-volume", pflag.ContinueOnError)
	backupVolumeFlags.StringVar(&o.BackupVolume.Address, "backup-storage-class", "local-path", "The storage class of the backup storage")
	backupVolumeFlags.StringVar(&o.BackupVolume.Path, "backup-storage-size", "/opt/nfs", "The size of the backup storage")
	cmd.Flags().AddFlagSet(backupVolumeFlags)
}

// AddFlags adds the flags to the command.
func (o *CreateOptions) AddFlags(cmd *cobra.Command) {
	// Add base and specific feature flags, Only support observer and zone config
	o.AddBaseFlags(cmd)
	o.AddObserverFlags(cmd)
	o.AddZoneFlags(cmd)
}
