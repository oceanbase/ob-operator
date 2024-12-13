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

	"github.com/oceanbase/ob-operator/internal/cli/tenant"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
)

// NewCreateCmd create an ob tenant
func NewCreateCmd() *cobra.Command {
	o := tenant.NewCreateOptions()
	logger := utils.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "create <tenant_name> --cluster=<cluster_name>",
		Short:   "Create an ob tenant",
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
			logger.Printf("Create OBTenant instance: %s", o.TenantName)
			logger.Printf("Run `echo $(kubectl get secret %s -n %s -o jsonpath='{.data.password}'|base64 --decode)` to get the secrets", obtenant.Spec.Credentials.Root, obtenant.Namespace)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
