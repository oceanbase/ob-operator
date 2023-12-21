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

package cli

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(tenantCmd)
	tenantCmd.AddCommand(
		tenantListCmd,
		tenantCreateCmd,
		tenantDeleteCmd,
		tenantGetCmd,
		tenantUpgradeCmd,
		tenantActivateCmd,
		tenantSwitchoverCmd,
		tenantChangePwdCmd,
		tenantReplayCmd,
	)
}

var tenantCmd = &cobra.Command{
	Use:   "tenant",
	Short: "Manage OBTenant resources",
}

var tenantListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "l"},
	Short:   "List OBTenant resources",
}

var tenantCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new OBTenant resource",
}

var tenantDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"del", "d"},
	Short:   "Delete an existing OBTenant resource",
}

var tenantGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an existing OBTenant resource",
}

var tenantUpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade an existing primary tenant",
}

var tenantActivateCmd = &cobra.Command{
	Use:   "activate",
	Short: "Activate an standby tenant",
}

var tenantSwitchoverCmd = &cobra.Command{
	Use:   "switchover",
	Short: "Switchover two OBTenants",
}

var tenantChangePwdCmd = &cobra.Command{
	Use:   "changepwd",
	Short: "Change password of an existing OBTenant's root user",
}

var tenantReplayCmd = &cobra.Command{
	Use:   "replay",
	Short: "Replay log of a tenant",
}
