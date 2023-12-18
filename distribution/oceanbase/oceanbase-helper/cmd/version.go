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

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
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
		return "", errors.New("OB Version Format is Wrong")
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print current installed version of OceanBase",
	Run: func(cmd *cobra.Command, args []string) {
		ver, err := getCurrentVersion(DefaultHomePath)
		if err != nil {
			fmt.Printf("Version command failed, %v \n", err)
			os.Exit(1)
		}
		fmt.Println(ver)
	},
}
