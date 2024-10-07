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
	"k8s.io/apimachinery/pkg/types"

	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/cli/tenant"
	"github.com/oceanbase/ob-operator/internal/clients"
)

// NewUpdateCmd update an obtenant
func NewUpdateCmd() *cobra.Command {
	o := tenant.NewUpdateOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "update <tenant_name>",
		Short:   "Update an ob tenant",
		Long:    "Update an ob tenant, support unitNumber/connectWhiteList/priority of zones",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"ud"},
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			obtenant, err := clients.GetOBTenant(cmd.Context(), types.NamespacedName{
				Name:      o.Name,
				Namespace: o.Namespace,
			})
			if err != nil {
				logger.Fatalln(err)
			}
			if err := cmdUtil.CheckTenantStatus(obtenant); err != nil {
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
			op := tenant.GetUpdateOperation(o)
			if _, err = clients.CreateOBTenantOperation(cmd.Context(), op); err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Create update operation for OBTenant %s successfully", o.Name)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
