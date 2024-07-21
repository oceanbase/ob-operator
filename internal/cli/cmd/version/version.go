package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of oceanbase cli",
		Long:  `All software has versions. This is oceanbase cli's`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("oceanbase cli")
		},
	}
	return cmd
}
