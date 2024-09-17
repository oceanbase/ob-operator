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

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/cli/tenant"
	"github.com/oceanbase/ob-operator/internal/clients"
)

// NewReplayLogCmd replay log of an ob tenant
func NewReplayLogCmd() *cobra.Command {
	o := tenant.NewReplayLogOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "replaylog <tenant_name>",
		Short:   "replay log of an ob tenant",
		Aliases: []string{"r"},
		Args:    cobra.ExactArgs(1),
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			nn := types.NamespacedName{
				Namespace: o.Namespace,
				Name:      o.Name,
			}
			obtenant, err := clients.GetOBTenant(cmd.Context(), nn)
			if err != nil {
				logger.Fatalln(err)
			}
			if err := cmdUtil.CheckTenantStatus(obtenant); err != nil {
				logger.Fatalln(err)
			}
			if obtenant.Status.TenantRole != apiconst.TenantRoleStandby {
				logger.Fatalln(fmt.Errorf("The tenant is not standby tenant"))
			}
			op := tenant.GetReplayLogOperation(o)
			if _, err = clients.CreateOBTenantOperation(cmd.Context(), op); err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Create replay log operation of tenant %s success", o.Name)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
