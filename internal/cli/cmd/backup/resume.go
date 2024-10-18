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
package backup

import (
	"github.com/spf13/cobra"

	"github.com/oceanbase/ob-operator/internal/cli/backup"
	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
)

// NewResumeCmd resume backup policy of the ob tenant
func NewResumeCmd() *cobra.Command {
	o := backup.NewResumeOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "resume <tenant_name>",
		Short:   "resume backup policy of the specified ob tenant",
		Long:    `resume backup policy of the specified ob tenant`,
		Args:    cobra.ExactArgs(1),
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			if err := backup.UpdateTenantBackupPolicy(cmd.Context(), &o.UpdateOptions); err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Resume backup policy for OBTenant %s successfully", o.Name)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
