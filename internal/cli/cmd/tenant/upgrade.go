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
	"k8s.io/apimachinery/pkg/types"

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/cli/tenant"
	"github.com/oceanbase/ob-operator/internal/clients"
)

// NewUpgradeCmd upgrade obtenant
func NewUpgradeCmd() *cobra.Command {
	o := tenant.NewUpgradeOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "upgrade <tenant_name>",
		Short:   "Upgrade ob tenant to compatible version to the cluster",
		Long:    "Upgrade ob tenant to higher version, suitable for restoring low-version backup data to a high-version cluster",
		Args:    cobra.ExactArgs(1),
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Validate(); err != nil {
				logger.Fatalln(err)
			}
			if err := o.Complete(); err != nil {
				logger.Fatalln(err)
			}
			nn := types.NamespacedName{
				Name:      o.Name,
				Namespace: o.Namespace,
			}
			obtenant, err := clients.GetOBTenant(cmd.Context(), nn)
			if err != nil {
				logger.Fatalln(err)
			}
			if obtenant.Status.TenantRole != apiconst.TenantRolePrimary {
				logger.Fatalln(errors.New("The tenant is not primary tenant"))
			}
			op := tenant.GetUpgradeOperation(o)
			_, err = clients.CreateOBTenantOperation(cmd.Context(), op)
			if err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Create upgrade operation for obtenant %s success", o.Name)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
