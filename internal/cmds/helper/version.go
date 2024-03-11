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

	"github.com/oceanbase/ob-operator/pkg/helper"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print current installed version of OceanBase",
	Run: func(cmd *cobra.Command, args []string) {
		ver, err := helper.GetCurrentVersion(DefaultHomePath)
		if err != nil {
			cmd.PrintErrf("Version command failed, %v \n", err)
			os.Exit(1)
		}
		cmd.Println(ver)
	},
}
