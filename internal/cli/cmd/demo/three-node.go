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
package demo

import (
	"github.com/oceanbase/ob-operator/internal/cli/cluster"
	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/clients"
	"github.com/spf13/cobra"
)

// NewThreeNodeCmd create a 1-1-1 three node cluster in a easier way
func NewThreeNodeCmd() *cobra.Command {
	o := cluster.NewCreateOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	pf := NewPromptFactory()
	cmd := &cobra.Command{
		Use:   cluster.THREE_NODE,
		Short: "deploy a three node ob cluster",
		Long:  "deploy a 1-1-1 three node ob cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := RunPromptsForCluster(pf, o); err != nil {
				logger.Fatalln(err)
			}
			if err := o.Complete(); err != nil {
				logger.Fatalln(err)
			}
			if err := o.SetDefaultConfig(cluster.THREE_NODE); err != nil {
				logger.Fatalln(err)
			}
			obcluster := cluster.CreateOBClusterInstance(o)
			if err := clients.CreateSecretsForOBCluster(cmd.Context(), obcluster, o.RootPassword); err != nil {
				logger.Fatalf("failed to create secrets for ob cluster: %v", err)
			}
			if _, err := clients.CreateOBCluster(cmd.Context(), obcluster); err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Create three-node OBCluster instance: %s", o.ClusterName)
			logger.Printf("Run `echo $(kubectl get secret %s -o jsonpath='{.data.password}'|base64 --decode)` to get the secrets", obcluster.Spec.UserSecrets.Root)
		},
	}
	return cmd
}
