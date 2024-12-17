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
	"github.com/spf13/cobra"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/oceanbase/ob-operator/internal/cli/tenant"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
	"github.com/oceanbase/ob-operator/internal/clients"
)

// NewScaleCmd scale an obtenant
func NewScaleCmd() *cobra.Command {
	o := tenant.NewScaleOptions()
	logger := utils.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "scale <tenant_name>",
		Short:   "Scale an ob tenant",
		Long:    `Scale an ob tenant, support unit-number/unit config of zones.`,
		Args:    cobra.ExactArgs(1),
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			obtenant, err := clients.GetOBTenant(cmd.Context(), types.NamespacedName{
				Name:      o.Name,
				Namespace: o.Namespace,
			})
			if err != nil {
				if kubeerrors.IsNotFound(err) {
					logger.Fatalf("OBTenant %s not found", o.Name)
				} else {
					logger.Fatalln(err)
				}
			}
			if err := utils.CheckTenantStatus(obtenant); err != nil {
				logger.Fatalln(err)
			} else {
				o.OldResourcePools = obtenant.Spec.Pools
			}
			if err := o.Validate(); err != nil {
				logger.Fatalln(err)
			}
			if err := o.Complete(); err != nil {
				logger.Fatalln(err)
			}
			op := tenant.GetScaleOperation(o)
			if _, err = clients.CreateOBTenantOperation(cmd.Context(), op); err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Create scale operation for OBTenant %s successfully", o.Name)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
