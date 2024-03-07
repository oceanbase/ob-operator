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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/oceanbase/oceanbase-helper/pkg/oceanbase"
)

func init() {
	rootCmd.AddCommand(standaloneCmd)
	standaloneCmd.AddCommand(standaloneValidateCmd)
}

var standaloneCmd = &cobra.Command{
	Use:   "standalone",
	Short: "Check conditions for standalone mode",
	Run: func(cmd *cobra.Command, args []string) {
		standaloneValidateCmd.Run(cmd, args)
	},
}

var standaloneValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "validate current version supports standalone mode",
	Run: func(cmd *cobra.Command, args []string) {
		ver, err := oceanbase.GetCurrentVersion(DefaultHomePath)
		if err != nil {
			fmt.Printf("Failed to get current version, %v \n", err)
			os.Exit(1)
		}
		obv, err := oceanbase.ParseOceanBaseVersion(ver)
		if err != nil {
			fmt.Printf("Failed to parse current version, %v \n", err)
			os.Exit(1)
		}
		if obv.Cmp(MinStandaloneVersion) < 0 {
			fmt.Printf("Current version %s is too low, please upgrade to %s first \n", ver, MinStandaloneVersion.String())
			os.Exit(1)
		}
	},
}
