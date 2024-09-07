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
	"fmt"

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

type UpdateOptions struct {
	generic.ResourceOptions
	force            bool
	Pools            []param.ResourcePoolSpec `json:"pools" binding:"required"`
	ConnectWhiteList string                   `json:"connectWhiteList,omitempty"`
	Charset          string                   `json:"charset,omitempty"`
	ZonePriority     map[string]string        `json:"zonePriority"`
	UpdateType       string                   `json:"updateType"`
	UnitConfig       *param.UnitConfig        `json:"unitConfig" binding:"required"`
	// Operation config
	OldResourcePools    []v1alpha1.ResourcePoolSpec `json:"oldResourcePools,omitempty"`
	ModifyResourcePools []v1alpha1.ResourcePoolSpec `json:"modifyResourcePools,omitempty"`
	AddResourcePools    []v1alpha1.ResourcePoolSpec `json:"addResourcePools,omitempty"`
	DeleteResourcePools []string                    `json:"deleteResourcePools,omitempty"`
}

func NewUpdateOptions() *UpdateOptions {
	return &UpdateOptions{
		ZonePriority:        make(map[string]string),
		Pools:               make([]param.ResourcePoolSpec, 0),
		ModifyResourcePools: make([]v1alpha1.ResourcePoolSpec, 0),
		AddResourcePools:    make([]v1alpha1.ResourcePoolSpec, 0),
		DeleteResourcePools: make([]string, 0),
		UnitConfig:          &param.UnitConfig{},
	}
}
func (o *UpdateOptions) Parse(cmd *cobra.Command, args []string) error {
	o.Name = args[0]
	o.Cmd = cmd
	if o.CheckIfFlagChanged("priority") {
		pools, err := utils.MapZonesToPools(o.ZonePriority)
		if err != nil {
			return err
		}
		o.Pools = pools
	}
	return nil
}

func (o *UpdateOptions) Complete() error {
	unitConfig, err := utils.ParseUnitConfig(o.UnitConfig)
	if err != nil {
		return err
	}
	switch o.UpdateType {
	case "addPools":
		for _, pool := range o.Pools {
			poolConfig := o.CreateResourcePoolSpec(pool, unitConfig)
			o.AddResourcePools = append(o.AddResourcePools, *poolConfig)
		}
	case "deletePools":
		for _, pool := range o.Pools {
			o.DeleteResourcePools = append(o.DeleteResourcePools, pool.Zone)
		}
	case "adjustPools":
		for _, pool := range o.Pools {
			for _, obpool := range o.OldResourcePools {
				if obpool.Zone == pool.Zone {
					poolConfig := o.CreateResourcePoolSpec(pool, obpool.UnitConfig)
					o.ModifyResourcePools = append(o.ModifyResourcePools, *poolConfig)
					break
				}
			}
		}
	}

	return nil
}

func GetUpdateOperation(o *UpdateOptions) *v1alpha1.OBTenantOperation {
	updateOp := &v1alpha1.OBTenantOperation{
		ObjectMeta: v1.ObjectMeta{
			Name:      o.Name + "-update-" + rand.String(6),
			Namespace: o.Namespace,
			Labels:    map[string]string{oceanbaseconst.LabelRefOBTenantOp: o.Name},
		},
		Spec: v1alpha1.OBTenantOperationSpec{
			TargetTenant: &o.Name,
			Force:        o.force,
		},
	}
	switch o.UpdateType {
	case "charset":
		updateOp.Spec.Charset = o.Charset
		updateOp.Spec.Type = apiconst.TenantOpSetCharset
	case "connect-white-list":
		updateOp.Spec.ConnectWhiteList = o.ConnectWhiteList
		updateOp.Spec.Type = apiconst.TenantOpSetConnectWhiteList
	case "addPools":
		updateOp.Spec.AddResourcePools = o.AddResourcePools
		updateOp.Spec.Type = apiconst.TenantOpAddResourcePools
	case "adjustPools":
		updateOp.Spec.ModifyResourcePools = o.ModifyResourcePools
		updateOp.Spec.Type = apiconst.TenantOpModifyResourcePools
	case "deletedPools":
		updateOp.Spec.DeleteResourcePools = o.DeleteResourcePools
		updateOp.Spec.Type = apiconst.TenantOpDeleteResourcePools
	}
	return updateOp
}

func (o *UpdateOptions) Validate() error {
	deleteNum := 0
	zoneNum := len(o.OldResourcePools)
	maxDeleteNum := zoneNum - 1
	typeMap := make(map[string]bool)
	updateTypeMap := func(name string) {
		if !typeMap[name] {
			typeMap[name] = true
			o.UpdateType = name
		}
	}
	if o.CheckIfFlagChanged("charset") {
		updateTypeMap("charset")
	}
	if o.CheckIfFlagChanged("connect-white-list") {
		updateTypeMap("connect-white-list")
	}
	if o.CheckIfFlagChanged("priority") && o.Pools != nil {
		found := false
		for _, pool := range o.Pools {
			for _, obpool := range o.OldResourcePools {
				if obpool.Zone == pool.Zone {
					found = true
					// priority set to 0 -> delete zone
					if pool.Priority == 0 {
						updateTypeMap("deletePools")
						deleteNum++
					} else {
						updateTypeMap("adjustPools")
					}
					break
				}
			}
			if !found {
				updateTypeMap("addPools")
			}
			if o.UpdateType == "deletedPools" && deleteNum > maxDeleteNum {
				return fmt.Errorf("OBTenant should have one zone at least")
			}
		}
		// Count the number of update types specified
		typeCount := len(typeMap)
		if typeCount > 1 {
			return errors.New("Only one type of update is allowed at a time")
		}
		if typeCount == 0 {
			return errors.New("No update type specified")
		}
	}
	return nil
}

// AddFlags add basic flags for tenant management
func (o *UpdateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Namespace, "namespace", "default", "The namespace of OBTenant")
	cmd.Flags().StringVar(&o.Charset, "charset", "", "The charset using in ob tenant")
	cmd.Flags().StringVar(&o.ConnectWhiteList, "connect-white-list", "", "The connect white list using in ob tenant")
	cmd.Flags().StringToStringVar(&o.ZonePriority, "priority", nil, "zone priority config of OBTenant")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false, "force operation")
	o.AddUnitFlags(cmd)
}

// AddUnitFlags add unit-resource-related flags
func (o *UpdateOptions) AddUnitFlags(cmd *cobra.Command) {
	unitFlags := pflag.NewFlagSet("unit", pflag.ContinueOnError)
	unitFlags.Int64Var(&o.UnitConfig.MaxIops, "max-iops", 1024, "The max iops of unit")
	unitFlags.Int64Var(&o.UnitConfig.MinIops, "min-iops", 1024, "The min iops of unit")
	unitFlags.IntVar(&o.UnitConfig.IopsWeight, "iops-weight", 1, "The iops weight of unit")
	unitFlags.StringVar(&o.UnitConfig.CPUCount, "cpu-count", "1", "The cpu count of unit")
	unitFlags.StringVar(&o.UnitConfig.MemorySize, "memory-size", "2Gi", "The memory size of unit")
	unitFlags.StringVar(&o.UnitConfig.LogDiskSize, "log-disk-size", "4Gi", "The log disk size of unit")
	cmd.Flags().AddFlagSet(unitFlags)
}
