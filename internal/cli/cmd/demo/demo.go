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
	"github.com/spf13/cobra"

	"github.com/oceanbase/ob-operator/internal/cli/cluster"
	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/clients"
)

// NewCmd create demo command for cluster creation
func NewCmd() *cobra.Command {
	o := cluster.NewCreateOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	pf := NewPromptFactory()
	cmd := &cobra.Command{
		Use:   "demo <subcommand>",
		Short: "deploy demo ob cluster in easier way",
		Long:  `deploy demo ob cluster in easier way, currently support single node and three node cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var clusterType string
			prompt := pf.CreatePrompt(cluster.FLAG_NAME)
			if o.Name, err = pf.RunPromptE(prompt); err != nil {
				logger.Fatalln(err)
			}
			prompt = pf.CreatePrompt(cluster.FLAG_NAMESPACE)
			if o.Namespace, err = pf.RunPromptE(prompt); err != nil {
				logger.Fatalln(err)
			}
			prompt = pf.CreatePrompt(cluster.CLUSTER_TYPE)
			if clusterType, err = pf.RunPromptE(prompt); err != nil {
				logger.Fatalln(err)
			}
			prompt = pf.CreatePrompt(cluster.FLAG_ROOT_PASSWORD)
			if o.RootPassword, err = pf.RunPromptE(prompt); err != nil {
				logger.Fatalln(err)
			}
			prompt = pf.CreatePrompt(cluster.FLAG_BACKUP_ADDRESS)
			if o.BackupVolume.Address, err = pf.RunPromptE(prompt); err != nil {
				logger.Fatalln(err)
			}
			prompt = pf.CreatePrompt(cluster.FLAG_BACKUP_PATH)
			if o.BackupVolume.Path, err = pf.RunPromptE(prompt); err != nil {
				logger.Fatalln(err)
			}
			if err := o.Complete(); err != nil {
				logger.Fatalln(err)
			}
			if err := o.SetDefaultConfig(clusterType); err != nil {
				logger.Fatalln(err)
			}
			obcluster := cluster.CreateOBClusterInstance(o)
			if err := clients.CreateSecretsForOBCluster(cmd.Context(), obcluster, o.RootPassword); err != nil {
				logger.Fatalf("failed to create secrets for ob cluster: %v", err)
			}
			if _, err := clients.CreateOBCluster(cmd.Context(), obcluster); err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Create OBCluster instance: %s", o.ClusterName)
			logger.Printf("Run `echo $(kubectl get secret %s -o jsonpath='{.data.password}'|base64 --decode)` to get the secrets", obcluster.Spec.UserSecrets.Root)
		},
	}
	return cmd
}
