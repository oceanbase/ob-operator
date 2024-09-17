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
package tenant

import (
	"fmt"
	"math"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/cli/generic"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	param "github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	oberr "github.com/oceanbase/ob-operator/pkg/errors"
)

type ScaleOptions struct {
	generic.ResourceOptions
	Pools        []param.ResourcePoolSpec `json:"pools" binding:"required"`
	ScaleType    string
	ZonePriority map[string]string
	Replica      int
	force        bool
	// Operation config
	OldResourcePools    []v1alpha1.ResourcePoolSpec `json:"oldResourcePools,omitempty"`
	ModifyResourcePools []v1alpha1.ResourcePoolSpec `json:"modifyResourcePools,omitempty"`
	AddResourcePools    []v1alpha1.ResourcePoolSpec `json:"addResourcePools,omitempty"`
	DeleteResourcePools []string                    `json:"deleteResourcePools,omitempty"`
	UnitConfig          *param.UnitConfig           `json:"unitConfig" binding:"required"`
}

func NewScaleOptions() *ScaleOptions {
	return &ScaleOptions{
		ZonePriority:        make(map[string]string),
		Pools:               make([]param.ResourcePoolSpec, 0),
		ModifyResourcePools: make([]v1alpha1.ResourcePoolSpec, 0),
		AddResourcePools:    make([]v1alpha1.ResourcePoolSpec, 0),
		DeleteResourcePools: make([]string, 0),
		UnitConfig:          &param.UnitConfig{},
	}
}

// GetScaleOperation creates scale opertaion
func GetScaleOperation(o *ScaleOptions) *v1alpha1.OBTenantOperation {
	scaleOp := &v1alpha1.OBTenantOperation{
		ObjectMeta: v1.ObjectMeta{
			Name:      o.Name + "-scale-" + rand.String(6),
			Namespace: o.Namespace,
			Labels:    map[string]string{oceanbaseconst.LabelRefOBTenantOp: o.Name},
		},
		Spec: v1alpha1.OBTenantOperationSpec{
			TargetTenant: &o.Name,
			Force:        o.force,
		},
	}
	switch o.ScaleType {
	case "deletePools":
		scaleOp.Spec.Type = apiconst.TenantOpDeleteResourcePools
		scaleOp.Spec.DeleteResourcePools = o.DeleteResourcePools
	case "addPools":
		scaleOp.Spec.Type = apiconst.TenantOpAddResourcePools
		scaleOp.Spec.AddResourcePools = o.AddResourcePools
	case "adjustPools":
		scaleOp.Spec.Type = apiconst.TenantOpModifyResourcePools
		scaleOp.Spec.ModifyResourcePools = o.ModifyResourcePools
	}

	return scaleOp
}

func (o *ScaleOptions) Parse(_ *cobra.Command, args []string) error {
	pools, err := utils.MapZonesToPools(o.ZonePriority)
	if err != nil {
		return err
	}
	o.Pools = pools
	o.Name = args[0]
	return nil
}

func (o *ScaleOptions) Complete() error {
	var err error
	var unitConfig *v1alpha1.UnitConfig
	unitConfig, err = o.parseUnitConfig()
	switch o.ScaleType {
	case "deletePools":
		for _, pool := range o.Pools {
			o.DeleteResourcePools = append(o.DeleteResourcePools, pool.Zone)
		}
	case "addPools":
		if err != nil {
			return err
		}
		for _, pool := range o.Pools {
			newPool := &v1alpha1.ResourcePoolSpec{
				Zone:     pool.Zone,
				Priority: pool.Priority,
				Type: &v1alpha1.LocalityType{
					Name:     o.Name,
					Replica:  o.Replica,
					IsActive: true,
				},
				UnitConfig: unitConfig,
			}
			o.AddResourcePools = append(o.AddResourcePools, *newPool)
		}
	case "adjustPools":
		for _, pool := range o.Pools {
			for i := 0; i < len(o.OldResourcePools); i++ {
				obpool := o.OldResourcePools[i]
				if obpool.Zone == pool.Zone {
					poolConfig := &v1alpha1.ResourcePoolSpec{
						Zone:     pool.Zone,
						Priority: pool.Priority,
						Type: &v1alpha1.LocalityType{
							Name:     o.Name,
							Replica:  o.Replica,
							IsActive: true,
						},
						UnitConfig: unitConfig,
					}
					o.ModifyResourcePools = append(o.ModifyResourcePools, *poolConfig)
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
	poolNum := len(o.OldResourcePools)
	maxDeleteNum := poolNum - poolNum/2
	typeDelete, typeAdjust, typeAdd, found := false, false, false, false
	for _, pool := range o.Pools {
		found = false
		for i := 0; i < poolNum; i++ {
			obpool := o.OldResourcePools[i]
			if obpool.Zone == pool.Zone {
				found = true
				if pool.Priority == 0 {
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
		if typeDelete && poolNum-deleteNum < maxDeleteNum {
			return fmt.Errorf("Obtenant has %d Pools, can only delete %d pools", poolNum, maxDeleteNum)
		}
	}
	trueCount := 0
	if typeDelete {
		trueCount++
		o.ScaleType = "deletePools"
	}
	if typeAdjust {
		trueCount++
		o.ScaleType = "adjustPools"
	}
	if typeAdd {
		trueCount++
		o.ScaleType = "addPools"
	}
	if trueCount > 1 {
		return fmt.Errorf("Only one type of scale is allowed at a time")
	}
	if trueCount == 0 {
		return fmt.Errorf("No scale type specified")
	}
	return nil
}

func (o *ScaleOptions) parseUnitConfig() (*v1alpha1.UnitConfig, error) {
	cpuCount, err := resource.ParseQuantity(o.UnitConfig.CPUCount)
	if err != nil {
		return nil, oberr.NewBadRequest("invalid cpu count: " + err.Error())
	}
	memorySize, err := resource.ParseQuantity(o.UnitConfig.MemorySize)
	if err != nil {
		return nil, oberr.NewBadRequest("invalid memory size: " + err.Error())
	}
	logDiskSize, err := resource.ParseQuantity(o.UnitConfig.LogDiskSize)
	if err != nil {
		return nil, oberr.NewBadRequest("invalid log disk size: " + err.Error())
	}
	var maxIops, minIops int
	if o.UnitConfig.MaxIops > math.MaxInt32 {
		maxIops = math.MaxInt32
	} else {
		maxIops = int(o.UnitConfig.MaxIops)
	}
	if o.UnitConfig.MinIops > math.MaxInt32 {
		minIops = math.MaxInt32
	} else {
		minIops = int(o.UnitConfig.MinIops)
	}
	return &v1alpha1.UnitConfig{
		MaxCPU:      cpuCount,
		MemorySize:  memorySize,
		MinCPU:      cpuCount,
		LogDiskSize: logDiskSize,
		MaxIops:     maxIops,
		MinIops:     minIops,
		IopsWeight:  o.UnitConfig.IopsWeight,
	}, nil
}

// AddFlags for scale options
func (o *ScaleOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Namespace, "namespace", "default", "namespace of OBTenant")
	cmd.Flags().IntVarP(&o.Replica, "replica", "r", 1, "replica of each local zone")
	cmd.Flags().StringToStringVar(&o.ZonePriority, "priority", nil, "zone priority config of OBTenant")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false, "force operation")
	o.AddUnitFlags(cmd)
}

// AddUnitFlags add unit-resource-related flags
func (o *ScaleOptions) AddUnitFlags(cmd *cobra.Command) {
	unitFlags := pflag.NewFlagSet("unit", pflag.ContinueOnError)
	unitFlags.Int64Var(&o.UnitConfig.MaxIops, "max-iops", 1024, "The max iops of unit")
	unitFlags.Int64Var(&o.UnitConfig.MinIops, "min-iops", 1024, "The min iops of unit")
	unitFlags.IntVar(&o.UnitConfig.IopsWeight, "iops-weight", 1, "The iops weight of unit")
	unitFlags.StringVar(&o.UnitConfig.CPUCount, "cpu-count", "1", "The cpu count of unit")
	unitFlags.StringVar(&o.UnitConfig.MemorySize, "memory-size", "2Gi", "The memory size of unit")
	unitFlags.StringVar(&o.UnitConfig.LogDiskSize, "log-disk-size", "4Gi", "The log disk size of unit")
	cmd.Flags().AddFlagSet(unitFlags)
}
