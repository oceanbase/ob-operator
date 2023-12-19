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
)

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.AddCommand(validateStandaloneCmd)
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "validate version and configs",
	Run: func(cmd *cobra.Command, args []string) {
		validateStandaloneCmd.Run(cmd, args)
	},
}

var validateStandaloneCmd = &cobra.Command{
	Use:   "standalone",
	Short: "validate current version supports standalone mode",
	Run: func(cmd *cobra.Command, args []string) {
		ver, err := getCurrentVersion(DefaultHomePath)
		if err != nil {
			fmt.Printf("Failed to get current version, %v \n", err)
			os.Exit(1)
		}
		if ver < MinStandaloneVersion {
			fmt.Printf("Current version %s is too low, please upgrade to %s first \n", ver, MinStandaloneVersion)
			os.Exit(1)
		}
	},
}
