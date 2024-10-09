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

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/constant"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/internal/cli/generic"
	utils "github.com/oceanbase/ob-operator/internal/cli/utils"
	modelcommon "github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	param "github.com/oceanbase/ob-operator/internal/dashboard/model/param"
)

type CreateOptions struct {
	generic.ResourceOption
	ClusterName  string               `json:"clusterName"`
	ClusterId    int64                `json:"clusterId"`
	RootPassword string               `json:"rootPassword"`
	Topology     []param.ZoneTopology `json:"topology"`
	OBServer     *param.OBServerSpec  `json:"observer"`
	Monitor      *param.MonitorSpec   `json:"monitor"`
	Parameters   []modelcommon.KVPair `json:"parameters"`
	BackupVolume *param.NFSVolumeSpec `json:"backupVolume"`
	Zones        map[string]string    `json:"zones"`
	KvParameters map[string]string    `json:"kvParameters"`
	Mode         string               `json:"mode"`
}

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

func (o *CreateOptions) Validate() error {
	if o.Namespace == "" {
		return errors.New("namespace is not specified")
	}
	if !utils.CheckPassword(o.RootPassword) {
		return fmt.Errorf("Password is not secure, must contain at least 2 uppercase and lowercase letters, numbers and special characters")
	}
	if !utils.CheckResourceName(o.Name) {
		return fmt.Errorf("invalid resource name in k8s: %s", o.Name)
	}
	return nil
}

func (o *CreateOptions) Parse(_ *cobra.Command, args []string) error {
	topology, err := utils.MapZonesToTopology(o.Zones)
	if err != nil {
		return err
	}
	parameters, err := utils.MapParameters(o.KvParameters)
	if err != nil {
		return err
	}
	o.Parameters = parameters
	o.Topology = topology
	o.Name = args[0]
	return nil
}

func (o *CreateOptions) Complete() error {
	// if not specific id, using timestamp
	if o.ClusterId == 0 {
		o.ClusterId = utils.GenerateClusterID()
	}
	// if not specific password, using random password, range [8,32]
	if o.RootPassword == "" {
		o.RootPassword = utils.GenerateRandomPassword(8, 32)
	}
	if o.ClusterName == "" {
		o.ClusterName = o.Name
	}
	return nil
}

func buildOBServerTemplate(observerSpec *param.OBServerSpec) *apitypes.OBServerTemplate {
	if observerSpec == nil {
		return nil
	}
	observerTemplate := &apitypes.OBServerTemplate{
		Image: observerSpec.Image,
		Resource: &apitypes.ResourceSpec{
			Cpu:    *apiresource.NewQuantity(observerSpec.Resource.Cpu, apiresource.DecimalSI),
			Memory: *apiresource.NewQuantity(observerSpec.Resource.MemoryGB*constant.GB, apiresource.BinarySI),
		},
		Storage: &apitypes.OceanbaseStorageSpec{
			DataStorage: &apitypes.StorageSpec{
				StorageClass: observerSpec.Storage.Data.StorageClass,
				Size:         *apiresource.NewQuantity(observerSpec.Storage.Data.SizeGB*constant.GB, apiresource.BinarySI),
			},
			RedoLogStorage: &apitypes.StorageSpec{
				StorageClass: observerSpec.Storage.RedoLog.StorageClass,
				Size:         *apiresource.NewQuantity(observerSpec.Storage.RedoLog.SizeGB*constant.GB, apiresource.BinarySI),
			},
			LogStorage: &apitypes.StorageSpec{
				StorageClass: observerSpec.Storage.Log.StorageClass,
				Size:         *apiresource.NewQuantity(observerSpec.Storage.Log.SizeGB*constant.GB, apiresource.BinarySI),
			},
		},
	}
	return observerTemplate
}

func buildBackupVolume(nfsVolumeSpec *param.NFSVolumeSpec) *apitypes.BackupVolumeSpec {
	if nfsVolumeSpec == nil {
		return nil
	}
	backupVolume := &apitypes.BackupVolumeSpec{
		Volume: &corev1.Volume{
			Name: "ob-backup",
			VolumeSource: corev1.VolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server:   nfsVolumeSpec.Address,
					Path:     nfsVolumeSpec.Path,
					ReadOnly: false,
				},
			},
		},
	}
	return backupVolume
}

func buildOBClusterTopology(topology []param.ZoneTopology) []apitypes.OBZoneTopology {
	obzoneTopology := make([]apitypes.OBZoneTopology, 0)
	for _, zone := range topology {
		topo := apitypes.OBZoneTopology{
			Zone:         zone.Zone,
			NodeSelector: common.KVsToMap(zone.NodeSelector),
			Replica:      zone.Replicas,
		}
		if len(zone.Affinities) > 0 {
			topo.Affinity = &corev1.Affinity{}
			for _, kv := range zone.Affinities {
				switch kv.Type {
				case modelcommon.NodeAffinityType:
					if topo.Affinity.NodeAffinity == nil {
						topo.Affinity.NodeAffinity = &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{},
							},
						}
					}
					nodeSelectorTerm := corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      kv.Key,
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{kv.Value},
						}},
					}
					topo.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = append(topo.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, nodeSelectorTerm)
				case modelcommon.PodAffinityType:
					if topo.Affinity.PodAffinity == nil {
						topo.Affinity.PodAffinity = &corev1.PodAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{},
						}
					}
					podAffinityTerm := corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{{
								Key:      kv.Key,
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{kv.Value},
							}},
						},
					}
					topo.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(topo.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, podAffinityTerm)
				case modelcommon.PodAntiAffinityType:
					if topo.Affinity.PodAntiAffinity == nil {
						topo.Affinity.PodAntiAffinity = &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{},
						}
					}
					podAntiAffinityTerm := corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{{
								Key:      kv.Key,
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{kv.Value},
							}},
						},
					}
					topo.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(topo.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, podAntiAffinityTerm)
				}
			}
		}
		if len(zone.Tolerations) > 0 {
			topo.Tolerations = make([]corev1.Toleration, 0)
			for _, kv := range zone.Tolerations {
				toleration := corev1.Toleration{
					Key:      kv.Key,
					Operator: corev1.TolerationOpEqual,
					Value:    kv.Value,
					Effect:   corev1.TaintEffectNoSchedule,
				}
				topo.Tolerations = append(topo.Tolerations, toleration)
			}
		}
		obzoneTopology = append(obzoneTopology, topo)
	}
	return obzoneTopology
}

func buildOBClusterParameters(parameters []modelcommon.KVPair) []apitypes.Parameter {
	obparameters := make([]apitypes.Parameter, 0)
	for _, parameter := range parameters {
		obparameters = append(obparameters, apitypes.Parameter{
			Name:  parameter.Key,
			Value: parameter.Value,
		})
	}
	return obparameters
}

func buildMonitorTemplate(monitorSpec *param.MonitorSpec) *apitypes.MonitorTemplate {
	if monitorSpec == nil {
		return nil
	}
	monitorTemplate := &apitypes.MonitorTemplate{
		Image: monitorSpec.Image,
		Resource: &apitypes.ResourceSpec{
			Cpu:    *apiresource.NewQuantity(monitorSpec.Resource.Cpu, apiresource.DecimalSI),
			Memory: *apiresource.NewQuantity(monitorSpec.Resource.MemoryGB*constant.GB, apiresource.BinarySI),
		},
	}
	return monitorTemplate
}

// Create an OBClusterInstance
func CreateOBClusterInstance(param *CreateOptions) *v1alpha1.OBCluster {
	observerTemplate := buildOBServerTemplate(param.OBServer)
	monitorTemplate := buildMonitorTemplate(param.Monitor)
	backupVolume := buildBackupVolume(param.BackupVolume)
	parameters := buildOBClusterParameters(param.Parameters)
	topology := buildOBClusterTopology(param.Topology)
	obcluster := &v1alpha1.OBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   param.Namespace,
			Name:        param.Name,
			Annotations: map[string]string{},
		},
		Spec: v1alpha1.OBClusterSpec{
			ClusterName:      param.ClusterName,
			ClusterId:        param.ClusterId,
			OBServerTemplate: observerTemplate,
			MonitorTemplate:  monitorTemplate,
			BackupVolume:     backupVolume,
			Parameters:       parameters,
			Topology:         topology,
			UserSecrets:      utils.GenerateUserSecrets(param.Name, param.ClusterId),
		},
	}
	switch param.Mode {
	case string(modelcommon.ClusterModeStandalone):
		obcluster.Annotations[oceanbaseconst.AnnotationsMode] = oceanbaseconst.ModeStandalone
	case string(modelcommon.ClusterModeService):
		obcluster.Annotations[oceanbaseconst.AnnotationsMode] = oceanbaseconst.ModeService
	default:
	}
	return obcluster
}

// AddFlags adds base and specific feature flags, Only support observer and zone config
func (o *CreateOptions) AddFlags(cmd *cobra.Command) {
	o.AddBaseFlags(cmd)
	o.AddObserverFlags(cmd)
	o.AddZoneFlags(cmd)
	o.AddParameterFlags(cmd)
}

// AddZoneFlags adds the zone-related flags to the command.
func (o *CreateOptions) AddZoneFlags(cmd *cobra.Command) {
	zoneFlags := pflag.NewFlagSet(FLAGSET_ZONE, pflag.ContinueOnError)
	zoneFlags.StringToStringVarP(&o.Zones, FLAG_ZONES, "z", map[string]string{"z1": "1"}, "The zones of the cluster in the format 'Zone=Replica', multiple values can be provided separated by commas")
	cmd.Flags().AddFlagSet(zoneFlags)
}

// AddBaseFlags adds the base flags to the command.
func (o *CreateOptions) AddBaseFlags(cmd *cobra.Command) {
	baseFlags := cmd.Flags()
	baseFlags.StringVarP(&o.ClusterName, FLAG_CLUSTER_NAME, "n", "", "Cluster name, if not specified, use resource name in k8s instead")
	baseFlags.StringVar(&o.Namespace, FLAG_NAMESPACE, DEFAULT_NAMESPACE, "The namespace of the cluster")
	baseFlags.Int64Var(&o.ClusterId, FLAG_CLUSTER_ID, DEFAULT_ID, "The id of the cluster")
	baseFlags.StringVarP(&o.RootPassword, FLAG_ROOTPASSWD, "p", "", "The root password of the cluster")
	baseFlags.StringVar(&o.Mode, FLAG_MODE, "", "The mode of the cluster")
}

// AddObserverFlags adds the observer-related flags to the command.
func (o *CreateOptions) AddObserverFlags(cmd *cobra.Command) {
	observerFlags := pflag.NewFlagSet(FLAGSET_OBSERVER, pflag.ContinueOnError)
	observerFlags.StringVar(&o.OBServer.Image, FLAG_OBSERVER_IMAGE, DEFAULT_OBSERVER_IMAGE, "The image of the observer")
	observerFlags.Int64Var(&o.OBServer.Resource.Cpu, FLAG_OBSERVER_CPU, DEFAULT_CPU_NUM, "The cpu of the observer")
	observerFlags.Int64Var(&o.OBServer.Resource.MemoryGB, FLAG_MONITOR_MEMORY, DEFAULT_MONITOR_MEMORY, "The memory of the observer")
	observerFlags.StringVar(&o.OBServer.Storage.Data.StorageClass, FLAG_DATA_STORAGE_CLASS, DEFAULT_DATA_STORAGE_CLASS, "The storage class of the data storage")
	observerFlags.StringVar(&o.OBServer.Storage.RedoLog.StorageClass, FLAG_REDO_LOG_STORAGE_CLASS, DEFAULT_REDO_LOG_STORAGE_CLASS, "The storage class of the redo log storage")
	observerFlags.StringVar(&o.OBServer.Storage.Log.StorageClass, FLAG_LOG_STORAGE_CLASS, DEFAULT_LOG_STORAGE_CLASS, "The storage class of the log storage")
	observerFlags.Int64Var(&o.OBServer.Storage.Data.SizeGB, FLAG_DATA_STORAGE_SIZE, DEFAULT_DATA_STORAGE_SIZE, "The size of the data storage")
	observerFlags.Int64Var(&o.OBServer.Storage.RedoLog.SizeGB, FLAG_REDO_LOG_STORAGE_SIZE, DEFAULT_REDO_LOG_STORAGE_SIZE, "The size of the redo log storage")
	observerFlags.Int64Var(&o.OBServer.Storage.Log.SizeGB, FLAG_LOG_STORAGE_SIZE, DEFAULT_LOG_STORAGE_SIZE, "The size of the log storage")
	cmd.Flags().AddFlagSet(observerFlags)
}

// AddMonitorFlags adds the monitor-related flags to the command.
func (o *CreateOptions) AddMonitorFlags(cmd *cobra.Command) {
	monitorFlags := pflag.NewFlagSet(FLAGSET_MONITOR, pflag.ContinueOnError)
	monitorFlags.StringVar(&o.Monitor.Image, FLAG_MONITOR_IMAGE, DEFAULT_MONITOR_IMAGE, "The image of the monitor")
	monitorFlags.Int64Var(&o.Monitor.Resource.Cpu, FLAG_MONITOR_CPU, DEFAULT_MONITOR_CPU, "The cpu of the monitor")
	monitorFlags.Int64Var(&o.Monitor.Resource.MemoryGB, FLAG_MONITOR_MEMORY, DEFAULT_MONITOR_MEMORY, "The memory of the monitor")
	cmd.Flags().AddFlagSet(monitorFlags)
}

// AddBackupVolumeFlags adds the backup-volume-related flags to the command.
func (o *CreateOptions) AddBackupVolumeFlags(cmd *cobra.Command) {
	backupVolumeFlags := pflag.NewFlagSet(FLAGSET_BACKUP_VOLUME, pflag.ContinueOnError)
	backupVolumeFlags.StringVar(&o.BackupVolume.Address, FLAG_BACKUP_ADDRESS, DEFAULT_BACKUP_ADDRESS, "The storage class of the backup storage")
	backupVolumeFlags.StringVar(&o.BackupVolume.Path, FLAG_BACKUP_PATH, DEFAULT_BACKUP_PATH, "The size of the backup storage")
	cmd.Flags().AddFlagSet(backupVolumeFlags)
}

// AddParameterFlags adds the parameter-related flags, e.g. __min_full_resource_pool_memory, to the command
func (o *CreateOptions) AddParameterFlags(cmd *cobra.Command) {
	parameterFlags := pflag.NewFlagSet(FLAGSET_PARAMETERS, pflag.ContinueOnError)
	parameterFlags.StringToStringVar(&o.KvParameters, FLAG_PARAMETERS, map[string]string{"__min_full_resource_pool_memory": DEFAULT_MIN_FULL_RESOURCE_POOL_MEMORY, "system_memory": DEFAULT_SYSTEM_MEMORY}, "Other parameter settings in OBCluster, e.g., __min_full_resource_pool_memory")
	cmd.Flags().AddFlagSet(parameterFlags)
}
