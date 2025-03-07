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

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/internal/cli/tenant"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
	"github.com/oceanbase/ob-operator/internal/clients"
)

// NewSwitchOverCmd switchover two tenants
func NewSwitchOverCmd() *cobra.Command {
	o := tenant.NewSwitchOverOptions()
	logger := utils.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "switchover <primary_tenant_name> <standby_tenant_name>",
		Short:   "Switchover of primary ob tenant and standby ob tenant",
		PreRunE: o.Parse,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Validate(); err != nil {
				logger.Fatalln(err)
			}
			standbyTenant, err := clients.GetOBTenant(cmd.Context(), types.NamespacedName{
				Name:      o.StandbyTenant,
				Namespace: o.Namespace,
			})
			if err != nil {
				if kubeerrors.IsNotFound(err) {
					logger.Fatalf("OBTenant %s not found", o.StandbyTenant)
				} else {
					logger.Fatalln(err)
				}
			}
			if err := utils.CheckTenantStatus(standbyTenant); err != nil {
				logger.Fatalln(err)
			}
			if err := utils.CheckPrimaryTenant(standbyTenant); err != nil {
				logger.Fatalln(err)
			}
			primaryTenant, err := clients.GetOBTenant(cmd.Context(), types.NamespacedName{
				Name:      o.PrimaryTenant,
				Namespace: o.Namespace,
			})
			if err != nil {
				if kubeerrors.IsNotFound(err) {
					logger.Fatalf("OBTenant %s not found", o.PrimaryTenant)
				} else {
					logger.Fatalln(err)
				}
			}
			if err := utils.CheckTenantStatus(primaryTenant); err != nil {
				logger.Fatalln(err)
			}
			if err := utils.CheckTenantRole(primaryTenant, apiconst.TenantRolePrimary); err != nil {
				logger.Fatalln(err)
			}
			op := tenant.GetSwitchOverOperation(o)
			if _, err := clients.CreateOBTenantOperation(cmd.Context(), op); err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Create switchover operation for primary tenant %s and standby tenant %s successfully", o.PrimaryTenant, o.StandbyTenant)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
