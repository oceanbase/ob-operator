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

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/cli/tenant"
	"github.com/oceanbase/ob-operator/internal/clients"
)

// NewActivateCmd activates a standby obtenant
func NewActivateCmd() *cobra.Command {
	o := tenant.NewActivateOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "activate <standby_tenant_name>",
		Short:   "Activate a standby ob tenant",
		PreRunE: o.Parse,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Validate(); err != nil {
				logger.Fatalln(err)
			}
			obtenant, err := clients.GetOBTenant(cmd.Context(), types.NamespacedName{
				Name:      o.Name,
				Namespace: o.Namespace,
			})
			if err != nil {
				logger.Fatalln(err)
			}
			if err := cmdUtil.CheckTenantStatus(obtenant); err != nil {
				logger.Fatalln(err)
			}
			if obtenant.Status.TenantRole == apiconst.TenantRolePrimary {
				logger.Fatalf("OBTenant %s is already PRIMARY", o.Name)
			}
			op := tenant.GetActivateOperation(o)
			if _, err = clients.CreateOBTenantOperation(cmd.Context(), op); err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Create activate operation for OBTenant %s successfully", o.Name)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
