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
	"fmt"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	utils "github.com/oceanbase/ob-operator/internal/cli/utils"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	param "github.com/oceanbase/ob-operator/internal/dashboard/model/param"
)

type ScaleOptions struct {
	ResourceOptions
	Zones             map[string]string            `json:"zones"`
	Topology          []param.ZoneTopology         `json:"topology"`
	OldTopology       []apitypes.OBZoneTopology    `json:"oldTopology"`
	ScaleType         string                       `json:"scaleType"`
	DeleteZonesConfig []string                     `json:"deleteZones,omitempty"`
	AdjustZonesConfig []v1alpha1.AlterZoneReplicas `json:"adjustReplicas,omitempty"`
	AddZonesConfig    []apitypes.OBZoneTopology    `json:"addZones,omitempty"`
}

func NewScaleOptions() *ScaleOptions {
	return &ScaleOptions{
		AdjustZonesConfig: make([]v1alpha1.AlterZoneReplicas, 0),
		AddZonesConfig:    make([]apitypes.OBZoneTopology, 0),
	}
}

// GetScaleOperation creates scale opertaions
func GetScaleOperation(o *ScaleOptions) *v1alpha1.OBClusterOperation {
	scaleOp := &v1alpha1.OBClusterOperation{
		ObjectMeta: v1.ObjectMeta{
			Name:      o.Name + "-scale-" + rand.String(6),
			Namespace: o.Namespace,
			Labels:    map[string]string{oceanbaseconst.LabelRefOBClusterOp: o.Name},
		},
		Spec: v1alpha1.OBClusterOperationSpec{
			OBCluster: o.Name,
		},
	}
	switch o.ScaleType {
	case "deleteZones":
		scaleOp.Spec.DeleteZones = o.DeleteZonesConfig
		scaleOp.Spec.Type = apiconst.ClusterOpTypeDeleteZones
	case "addZones":
		scaleOp.Spec.AddZones = o.AddZonesConfig
		scaleOp.Spec.Type = apiconst.ClusterOpTypeAddZones
	case "adjustReplicas":
		scaleOp.Spec.AdjustReplicas = o.AdjustZonesConfig
		scaleOp.Spec.Type = apiconst.ClusterOpTypeAdjustReplicas
	}
	return scaleOp
}

func (o *ScaleOptions) Parse(_ *cobra.Command, args []string) error {
	topology, err := utils.MapZonesToTopology(o.Zones)
	if err != nil {
		return err
	}
	o.Topology = topology
	o.Name = args[0]
	return nil
}

func (o *ScaleOptions) Complete() error {
	switch o.ScaleType {
	case "deleteZones":
		for _, zone := range o.Topology {
			o.DeleteZonesConfig = append(o.DeleteZonesConfig, zone.Zone)
		}
	case "addZones":
		for _, zone := range o.Topology {
			newZone := &apitypes.OBZoneTopology{
				Zone:    zone.Zone,
				Replica: zone.Replicas,
			}
			o.AddZonesConfig = append(o.AddZonesConfig, *newZone)
		}
	case "adjustReplicas":
		for _, zone := range o.Topology {
			for i := 0; i < len(o.OldTopology); i++ {
				obzone := o.OldTopology[i]
				if obzone.Zone == zone.Zone {
					zoneConfig := &v1alpha1.AlterZoneReplicas{
						Zones: []string{zone.Zone},
						To:    zone.Replicas,
					}
					o.AdjustZonesConfig = append(o.AdjustZonesConfig, *zoneConfig)
					break
				}
			}
		}
	default:
	}
	return nil
}

func (o *ScaleOptions) Validate() error {
	deleteNum := 0
	zoneNum := len(o.OldTopology)
	maxDeleteNum := zoneNum - zoneNum/2
	typeDelete, typeAdjust, typeAdd, found := false, false, false, false
	for _, zone := range o.Topology {
		found = false
		for i := 0; i < zoneNum; i++ {
			obzone := o.OldTopology[i]
			if obzone.Zone == zone.Zone {
				found = true
				if zone.Replicas == 0 {
					typeDelete = true
					deleteNum++
				} else {
					typeAdjust = true
				}
				break
			}
		}
		if !found {
			typeAdd = true
		}
		if typeDelete && zoneNum-deleteNum < maxDeleteNum {
			return fmt.Errorf("Obcluster has %d Zones, can only delete %d zones", zoneNum, maxDeleteNum)
		}
	}
	trueCount := 0
	if typeDelete {
		trueCount++
		o.ScaleType = "deleteZones"
	}
	if typeAdjust {
		trueCount++
		o.ScaleType = "adjustReplicas"
	}
	if typeAdd {
		trueCount++
		o.ScaleType = "addZones"
	}
	if trueCount > 1 {
		return fmt.Errorf("Only one type of scale is allower at a time")
	}
	if trueCount == 0 {
		return fmt.Errorf("No scale type specified")
	}
	return nil
}

// Add Flags for scale options
func (o *ScaleOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Namespace, "namespace", "default", "namespace of ob cluster")
	cmd.Flags().StringToStringVar(&o.Zones, "zones", nil, "zone of ob cluster")
}
