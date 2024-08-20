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
package version

import (
	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of oceanbase cli",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Println("Oceanbase cli Version:0.0.1")
		},
	}
	return cmd
}
