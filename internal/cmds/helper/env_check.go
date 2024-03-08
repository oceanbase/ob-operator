/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package helper

import (
	"os"

	"github.com/spf13/cobra"

	cmdconst "github.com/oceanbase/ob-operator/internal/const/cmd"
	"github.com/oceanbase/ob-operator/pkg/helper"
)

func init() {
	rootCmd.AddCommand(newEnvCheckCmd())
}

func newEnvCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env-check",
		Short: "Check whether the environment is ready",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	cmd.AddCommand(newCheckStorageCmd())
	return cmd
}

func newCheckStorageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storage",
		Short: "Check whether the storage is ready",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				os.Exit(int(cmdconst.ExitCodeBadArgs))
			}
			err := helper.TryFallocate(args[0])
			if err != nil {
				cmd.PrintErrf("Failed to check storage, %v \n", err)
				os.Exit(int(cmdconst.ExitCodeNotSupport))
			}
		},
	}
	return cmd
}
