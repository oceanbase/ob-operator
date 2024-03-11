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
	rootCmd.AddCommand(newServiceCheckCmd())
}

func newServiceCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Support service mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newServiceValidateCmd())
	return cmd
}

func newServiceValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Check whether the observer supports service mode",
		Run: func(cmd *cobra.Command, args []string) {
			obVer, err := helper.GetCurrentVersion(DefaultHomePath)
			if err != nil {
				cmd.PrintErrf("Failed to get current version, %v \n", err)
				os.Exit(1)
			}
			obv, err := helper.ParseOceanBaseVersion(obVer)
			if err != nil {
				cmd.PrintErrf("Failed to parse current version, %v \n", err)
				os.Exit(1)
			}
			if obv.Cmp(MinServiceVersion) < 0 {
				cmd.PrintErrf("Current version %s is too low, please upgrade to %s first \n", obVer, MinServiceVersion.String())
				os.Exit(int(cmdconst.ExitCodeNotSupport))
			}
		},
	}

	return cmd
}
