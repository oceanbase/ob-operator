package util

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func PrintFlagValues(cmd *cobra.Command) {
	// 确保命令的 flag 已经被解析
	_ = cmd.ParseFlags(nil)

	// 遍历所有 flag
	flags := cmd.NonInheritedFlags()
	flags.VisitAll(func(f *pflag.Flag) {
		fmt.Printf("%s : %v\n", f.Name, f.Value.String())
	})

	// 遍历继承的 flag（如果有）
	inheritedFlags := cmd.InheritedFlags()
	inheritedFlags.VisitAll(func(f *pflag.Flag) {
		fmt.Printf("%s : %v\n", f.Name, f.Value.String())
	})

}
