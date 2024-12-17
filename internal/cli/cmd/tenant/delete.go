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

	"github.com/oceanbase/ob-operator/internal/cli/tenant"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
	"github.com/oceanbase/ob-operator/internal/clients"
)

// NewDeleteCmd delete ob tenant
func NewDeleteCmd() *cobra.Command {
	o := tenant.NewDeleteOptions()
	logger := utils.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "delete <tenant_name>",
		Short:   "Delete an ob tenant",
		Args:    cobra.ExactArgs(1),
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			err := clients.DeleteOBTenant(cmd.Context(), types.NamespacedName{
				Namespace: o.Namespace,
				Name:      o.Name,
			})
			if err != nil {
				if kubeerrors.IsNotFound(err) {
					logger.Fatalf("OBTenant %s not found", o.Name)
				} else {
					logger.Fatalln(err)
				}
			}
			logger.Printf("Delete OBTenant %s successfully", o.Name)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
