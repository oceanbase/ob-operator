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

package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/oceanbase/oceanbase-helper/pkg/oceanbase"
)

func init() {
	rootCmd.AddCommand(newServiceCheckCmd())
}

func newServiceCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "svc-check",
		Short: "Check whether the observer support service mode",
		Run: func(cmd *cobra.Command, args []string) {
			obVer, err := oceanbase.GetCurrentVersion(DefaultHomePath)
			if err != nil {
				cmd.PrintErrf("Failed to get current version, %v \n", err)
				os.Exit(1)
			}
			obv, err := oceanbase.ParseOceanBaseVersion(obVer)
			if err != nil {
				cmd.PrintErrf("Failed to parse current version, %v \n", err)
				os.Exit(1)
			}
			if obv.Cmp(MinServiceVersion) < 0 {
				cmd.PrintErrf("Current version %s is too low, please upgrade to %s first \n", obVer, MinServiceVersion.String())
				os.Exit(int(ExitCodeNotSupport))
			}
		},
	}
	return cmd
}
