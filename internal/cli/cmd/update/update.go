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
package update

import (
	"github.com/spf13/cobra"

	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/cli/update"
)

// NewCmd update the ob-operator and other components
func NewCmd() *cobra.Command {
	o := update.NewUpdateOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:   "update <components>",
		Short: "Command for ob-operator and components update",
		Long: `Command for ob-operator and components update.

Currently support:
- ob-operator: A Kubernetes operator that simplifies the deployment and management of OceanBase cluster and related resources on Kubernetes.
- ob-dashboard: A web application that provides resource management capabilities.
- local-path-provisioner: Provides a way for the Kubernetes users to utilize the local storage in each node, Storage of OceanBase cluster relies on it, which should be installed beforehand.
- cert-manager: Creates TLS certificates for workloads in Kubernetes and renews the certificates before they expire, ob-operator relies on it for certificate management, which should be installed beforehand.
		
if not specified, update ob-operator and ob-dashboard by default`,
		PreRunE: o.Parse,
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				if err := o.InstallAll(); err != nil {
					logger.Fatalln(err)
				} else {
					logger.Println("")
				}
			} else {
				if err := o.Install(args[0]); err != nil {
					logger.Fatalln(err)
				} else {
					logger.Printf("%s install successfully", args[0])
				}
			}
		},
	}
	return cmd
}
