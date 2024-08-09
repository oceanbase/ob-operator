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
package install

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
)

// NewInstallCmd install the ob-operator and other components
func NewInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install <components>",
		Short: "Command for ob-operator and components installation",
		Run: func(cmd *cobra.Command, args []string) {
			// 如果没提供参数，默认安装所有组件，修改option参数即可
			log.Println("Installing components:", strings.Join(args, ", "))
		},
	}
	return cmd
}
