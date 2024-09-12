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
	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/cli/tenant"
	"github.com/spf13/cobra"
)

// NewCreateCmd create an ob tenant
func NewCreateCmd() *cobra.Command {
	o := tenant.NewCreateOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "create <tenant_name> --cluster=<cluster_name> ",
		Short:   "Create ob tenant",
		Aliases: []string{"c"},
		PreRunE: o.Parse,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Complete(); err != nil {
				logger.Fatalln(err)
			}
			if err := o.Validate(); err != nil {
				logger.Fatalln(err)
			}
			obtenant, err := tenant.CreateOBTenant(cmd.Context(), o)
			if err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Create obtenant instance: %s", o.TenantName)
			logger.Printf("Run `echo $(kubectl get secret %s -o jsonpath='{.data.password}'|base64 --decode)` to get the secrets", obtenant.Spec.Credentials.Root)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
