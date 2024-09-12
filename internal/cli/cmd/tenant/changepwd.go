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

// NewChangePwdCmd changes password of an obtenant
func NewChangePwdCmd() *cobra.Command {
	o := tenant.NewChangePwdOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "changepwd <tenant_name> --password=<password>",
		Short:   "Change password of ob tenant",
		PreRunE: o.Parse,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Complete(); err != nil {
				logger.Fatalln(err)
			}
			if err := o.Validate(); err != nil {
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
			if obtenant.Status.Status != tenantstatus.Running {
				logger.Fatalln(fmt.Errorf("Obtenant status invalid, Status:%s", obtenant.Status.Status))
			}
			if err := tenant.GenerateNewPwd(o, cmd.Context()); err != nil {
				logger.Fatalln(err)
			} else {
				logger.Println("New password generated success")
			}
			op := tenant.GetChangePwdOperation(o)
			_, err = clients.CreateOBTenantOperation(cmd.Context(), op)
			if err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Create changepwd operation for obtenant %s success", o.Name)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
