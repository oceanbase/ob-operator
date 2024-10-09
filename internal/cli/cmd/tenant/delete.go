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

// NewDeleteCmd delete ob tenant
func NewDeleteCmd() *cobra.Command {
	o := tenant.NewDeleteOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "delete <tenant_name>",
		Short:   "Delete an ob tenant",
		Aliases: []string{"d"},
		Args:    cobra.ExactArgs(1),
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			err := clients.DeleteOBTenant(cmd.Context(), types.NamespacedName{
				Namespace: o.Namespace,
				Name:      o.Name,
			})
			if err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Delete OBTenant %s successfully", o.Name)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
