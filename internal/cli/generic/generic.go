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
package generic

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
)

type ResourceOptions struct {
	Name      string
	Namespace string
	Cmd       *cobra.Command
}

// Parse the args in obocli
func (o *ResourceOptions) Parse(cmd *cobra.Command, args []string) error {
	o.Name = args[0]
	o.Cmd = cmd
	return nil
}

// Complete the unset params in options
func (o *ResourceOptions) Complete() error {
	return nil
}

// Validate the params in options
func (o *ResourceOptions) Validate() error {
	if o.Namespace == "" {
		return errors.New("namespace not specified")
	}
	return nil
}

// CreateResourcePoolSpec Creates ResourcePoolSpec for tenant scale and update
func (o *ResourceOptions) CreateResourcePoolSpec(pool param.ResourcePoolSpec, unitConfig *v1alpha1.UnitConfig) *v1alpha1.ResourcePoolSpec {
	return &v1alpha1.ResourcePoolSpec{
		Zone:     pool.Zone,
		Priority: pool.Priority,
		Type: &v1alpha1.LocalityType{
			Name:     o.Name,
			Replica:  1,
			IsActive: true,
		},
		UnitConfig: unitConfig,
	}
}

// CheckIfFlagChanged checks if flags has changed
func (o *ResourceOptions) CheckIfFlagChanged(flags ...string) bool {
	for _, flagName := range flags {
		if flag := o.Cmd.Flags().Lookup(flagName); flag != nil && flag.Changed {
			return true
		}
	}
	return false
}
