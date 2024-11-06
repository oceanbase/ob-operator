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
package utils

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// PrintFlagValues is used for debugging
func PrintFlagValues(cmd *cobra.Command) {
	_ = cmd.ParseFlags(nil)

	flags := cmd.NonInheritedFlags()
	flags.VisitAll(func(f *pflag.Flag) {
		fmt.Printf("%s : %v\n", f.Name, f.Value.String())
	})

	inheritedFlags := cmd.InheritedFlags()
	inheritedFlags.VisitAll(func(f *pflag.Flag) {
		fmt.Printf("%s : %v\n", f.Name, f.Value.String())
	})
}
