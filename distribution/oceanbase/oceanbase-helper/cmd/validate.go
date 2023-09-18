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
	"os/exec"
	"strings"

	"github.com/oceanbase/oceanbase-helper/pkg/oceanbase"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "validate a version can be upgrade to current version",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		err := validateUpgrade()
		if err != nil {
			fmt.Printf("Upgrade validate failed, %v \n", err)
			os.Exit(1)
		}
	},
}

func init() {
	upgradeCmd.AddCommand(validateCmd)
	validateCmd.PersistentFlags().StringP("start-version", "s", "", "upgrade start version")
	validateCmd.PersistentFlags().StringP("ob-installation-path", "p", "/home/admin/oceanbase", "oceanbase installation path")
	_ = viper.BindPFlag("start-version", validateCmd.PersistentFlags().Lookup("start-version"))
	_ = viper.BindPFlag("ob-installation-path", validateCmd.PersistentFlags().Lookup("ob-installation-path"))
}

func getCurrentVersion(oceanbaseInstallPath string) (string, error) {
	output, err := exec.Command("bash", "-c", fmt.Sprintf("export LD_LIBRARY_PATH=%s/lib; %s/bin/observer -V", oceanbaseInstallPath, oceanbaseInstallPath)).CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, "Failed to execute version command")
	}
	fmt.Println(string(output))
	lines := strings.Split(string(output), "\n")
	if len(lines) > 3 {
		versionStr := strings.Split(lines[1], " ")
		version := versionStr[len(versionStr)-1]
		releaseStr := strings.Split(strings.Split(lines[3], " ")[1], "-")[0]
		return fmt.Sprintf("%s-%s", version[0:len(version)-1], releaseStr), nil
	} else {
		return "", errors.New("OB Version Formattion is Wrong")
	}
}

func validateUpgrade() error {
	fromVersion := viper.GetString("start-version")
	oceanbaseInstallPath := viper.GetString("ob-installation-path")
	targetVersion, err := getCurrentVersion(oceanbaseInstallPath)
	fmt.Println(targetVersion)
	if err != nil {
		return errors.Wrap(err, "Failed to current oceanbase version")
	}
	route, err := oceanbase.GetOBUpgradeRoute(&oceanbase.OBUpgradeRouteParam{
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
