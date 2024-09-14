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

	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/cli/tenant"
	"github.com/oceanbase/ob-operator/internal/clients"
	"github.com/oceanbase/ob-operator/internal/const/status/tenantstatus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
)

// NewSwitchOverCmd switchover two tenants
func NewSwitchOverCmd() *cobra.Command {
	o := tenant.NewSwitchOverOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "switchover <primary_tenant_name> <standby_tenant_name>",
		Short:   "Switchover of primary tenant and standby tenant",
		PreRunE: o.Parse,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Complete(); err != nil {
				logger.Fatalln(err)
			}
			if err := o.Validate(); err != nil {
				logger.Fatalln(err)
			}
			nn := types.NamespacedName{
				Name:      o.StandbyTenant,
				Namespace: o.Namespace,
			}
			obtenant, err := clients.GetOBTenant(cmd.Context(), nn)
			if err != nil {
				logger.Fatalln(err)
			}
			if obtenant.Status.Status != tenantstatus.Running {
				logger.Fatalln(fmt.Errorf("Obtenant status invalid, Status:%s", obtenant.Status.Status))
			}
			if obtenant.Spec.Source == nil || obtenant.Spec.Source.Tenant == nil {
				logger.Fatalf("Obtenant %s has no primary tenant", o.StandbyTenant)
			}
			op := tenant.GetSwitchOverOperation(o)
			_, err = clients.CreateOBTenantOperation(cmd.Context(), op)
			if err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Create switchover operation for primary tenant %s and standby tenant %s success", o.PrimaryTenant, o.StandbyTenant)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
