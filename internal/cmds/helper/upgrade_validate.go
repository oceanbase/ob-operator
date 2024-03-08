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
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/oceanbase/ob-operator/pkg/helper"
)

// upgradeValidateCmd represents the validate command
var upgradeValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "validate a version can be upgrade to current version",
	Run: func(cmd *cobra.Command, args []string) {
		err := validateUpgrade()
		if err != nil {
			cmd.PrintErrf("Upgrade validate failed, %v \n", err)
			os.Exit(1)
		}
	},
}

func init() {
	upgradeCmd.AddCommand(upgradeValidateCmd)
	upgradeValidateCmd.PersistentFlags().StringP("start-version", "s", "", "upgrade start version")
	upgradeValidateCmd.PersistentFlags().StringP("ob-installation-path", "p", "/home/admin/oceanbase", "oceanbase installation path")
	_ = viper.BindPFlag("start-version", upgradeValidateCmd.PersistentFlags().Lookup("start-version"))
	_ = viper.BindPFlag("ob-installation-path", upgradeValidateCmd.PersistentFlags().Lookup("ob-installation-path"))
}

func validateUpgrade() error {
	fromVersion := viper.GetString("start-version")
	oceanbaseInstallPath := viper.GetString("ob-installation-path")
	targetVersion, err := helper.GetCurrentVersion(oceanbaseInstallPath)
	log.Println(targetVersion)
	if err != nil {
		return errors.Wrap(err, "Failed to current oceanbase version")
	}
	route, err := helper.GetOBUpgradeRoute(&helper.OBUpgradeRouteParam{
		StartVersion:  fromVersion,
		TargetVersion: targetVersion,
		DepFilePath:   fmt.Sprintf("%s/etc/oceanbase_upgrade_dep.yml", oceanbaseInstallPath),
	})
	if err != nil {
		return errors.Wrapf(err, "Failed to get upgrade route from %s to %s", fromVersion, targetVersion)
	}
	for idx, n := range route {
		if n.RequireFromBinary.Value && idx != 0 && idx != len(route)-1 {
			return errors.Errorf("Found version %s require binary", n.Version)
		}
	}
	return nil
}
