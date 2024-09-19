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
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/cli/generic"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	param "github.com/oceanbase/ob-operator/internal/dashboard/model/param"
)

type ScaleOptions struct {
	generic.ResourceOptions
	ScaleType  string
	UnitNumber int
	force      bool
	// Operation config
	UnitConfig          *param.UnitConfig           `json:"unitConfig" binding:"required"`
	OldResourcePools    []v1alpha1.ResourcePoolSpec `json:"oldResourcePools,omitempty"`
	ModifyResourcePools []v1alpha1.ResourcePoolSpec `json:"modifyResourcePools,omitempty"`
}

func NewScaleOptions() *ScaleOptions {
	return &ScaleOptions{
		UnitConfig:          &param.UnitConfig{},
		OldResourcePools:    make([]v1alpha1.ResourcePoolSpec, 0),
		ModifyResourcePools: make([]v1alpha1.ResourcePoolSpec, 0),
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
	case "unit-number":
		scaleOp.Spec.Type = apiconst.TenantOpSetUnitNumber
		scaleOp.Spec.UnitNumber = o.UnitNumber
	case "unit-config":
		scaleOp.Spec.Type = apiconst.TenantOpModifyResourcePools
		scaleOp.Spec.ModifyResourcePools = o.ModifyResourcePools
	}

	return scaleOp
}

func (o *ScaleOptions) Complete() error {
	unitConfig, err := utils.ParseUnitConfig(o.UnitConfig)
	if err != nil {
		return err
	}
	switch o.ScaleType {
	case "unit-config":
		for _, pool := range o.OldResourcePools {
			poolConfig := pool.DeepCopy()
			poolConfig.UnitConfig = unitConfig
			o.ModifyResourcePools = append(o.ModifyResourcePools, *poolConfig)
		}
	case "addPrimaryZones", "deletePrimaryZones":
		// TODO: add primaryZone and delete primaryZone
	default:
	}
	return nil
}

func (o *ScaleOptions) Validate() error {
	typeCount := 0
	unitFlags := []string{"max-iops", "min-iops", "iops-weight", "cpu-count", "memory-size", "log-disk-size"}
	if o.CheckFlagsChanged(unitFlags) {
		o.ScaleType = "unit-config"
		typeCount++
	}
	if o.CheckFlagChanged("unit-number") {
		o.ScaleType = "unit-number"
		typeCount++
	}
	if typeCount > 1 {
		return errors.New("Only one type of scale is allowed at a time")
	}
	if typeCount == 0 {
		return errors.New("No scale type specified")
	}
	if o.ScaleType == "unit-number" && o.UnitNumber < 1 {
		return errors.New("unit number must be greater than one")
	}
	return nil
}

// AddFlags for scale options
func (o *ScaleOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Namespace, "namespace", "default", "namespace of OBTenant")
	cmd.Flags().IntVar(&o.UnitNumber, "unit-number", 1, "unit-number of pools")
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
