package cli

import (
	"fmt"

	"github.com/oceanbase/ob-operator/internal/cli/cmd/cluster"
	"github.com/oceanbase/ob-operator/internal/cli/cmd/install"
	"github.com/oceanbase/ob-operator/internal/cli/cmd/tenant"
	"github.com/oceanbase/ob-operator/internal/cli/cmd/version"
	"github.com/spf13/cobra"
)

func NewCliCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "obocli",
		Short: "OceanBase Operator CLI",
		Long:  "OceanBase Operator CLI tool to manage OceanBase clusters, tenants, and backups.",
		Run: func(cmd *cobra.Command, args []string) {
			if cmd.Flags().Changed("version") {
				versionCmd := version.NewVersionCmd()
				versionCmd.Run(cmd, args)
			} else {
				cmd.Help()
			}
		},
	}
	cmd.AddCommand(version.NewVersionCmd())
	cmd.AddCommand(cluster.NewClusterCmd())
	cmd.AddCommand(tenant.NewTenantCmd())
	cmd.AddCommand(install.NewInstallCmd())
	cmd.Flags().BoolP("version", "v", false, "Print the version number of oceanbase cli")
	return cmd
}

func Execute() {
	rootCmd := NewCliCmd()
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
