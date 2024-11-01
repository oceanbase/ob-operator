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
	"github.com/oceanbase/ob-operator/internal/cli/demo"
	"github.com/oceanbase/ob-operator/internal/cli/tenant"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
	"github.com/oceanbase/ob-operator/internal/clients"
)

// NewCmd create demo command for cluster creation
func NewCmd() *cobra.Command {
	clusterOptions := cluster.NewCreateOptions()
	tenantOptions := tenant.NewCreateOptions()
	logger := utils.GetDefaultLoggerInstance()
	pf := demo.NewPromptFactory()
	wait := false
	cmd := &cobra.Command{
		Use:   "demo <subcommand>",
		Short: "deploy demo ob cluster and tenant in easier way",
		Long:  `deploy demo ob cluster and tenant in easier way, currently support single node and three node cluster, with corresponding tenant`,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var clusterType string
			prompt := pf.CreatePrompt(cluster.FLAG_NAME)
			if clusterOptions.Name, err = pf.RunPromptE(prompt); err != nil {
				logger.Fatalln(err)
			}
			prompt = pf.CreatePrompt(cluster.FLAG_NAMESPACE)
			if clusterOptions.Namespace, err = pf.RunPromptE(prompt); err != nil {
				logger.Fatalln(err)
			}
			prompt = pf.CreatePrompt(cluster.CLUSTER_TYPE)
			if clusterType, err = pf.RunPromptE(prompt); err != nil {
				logger.Fatalln(err)
			}
			prompt = pf.CreatePrompt(cluster.FLAG_ROOT_PASSWORD)
			if clusterOptions.RootPassword, err = pf.RunPromptE(prompt); err != nil {
				logger.Fatalln(err)
			}
			if err := clusterOptions.Complete(); err != nil {
				logger.Fatalln(err)
			}
			if err := tenantOptions.Complete(); err != nil {
				logger.Fatalln(err)
			}
			if err := demo.SetDefaultClusterConf(clusterType, clusterOptions); err != nil {
				logger.Fatalln(err)
			}
			if err := demo.SetDefaultTenantConf(clusterType, clusterOptions.Namespace, clusterOptions.Name, tenantOptions); err != nil {
				logger.Fatalln(err)
			}
			obcluster := cluster.CreateOBClusterInstance(clusterOptions)
			if err := clients.CreateSecretsForOBCluster(cmd.Context(), obcluster, clusterOptions.RootPassword); err != nil {
				logger.Fatalf("failed to create secrets for ob cluster: %v", err)
			}
			if _, err := clients.CreateOBCluster(cmd.Context(), obcluster); err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Create OBCluster instance: %s", clusterOptions.ClusterName)
			logger.Printf("Run `echo $(kubectl get secret %s -clusterOptions jsonpath='{.data.password}'|base64 --decode)` to get the secrets", obcluster.Spec.UserSecrets.Root)
			obtenant, err := tenant.CreateOBTenant(cmd.Context(), tenantOptions)
			if err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Create OBTenant instance: %s", tenantOptions.TenantName)
			logger.Printf("Run `echo $(kubectl get secret %s -o jsonpath='{.data.password}'|base64 --decode)` to get the secrets", obtenant.Spec.Credentials.Root)
		},
	}
	cmd.Flags().BoolVarP(&wait, cluster.FLAG_WAIT, "w", cluster.DEFAULT_WAIT, "wait for the cluster and tenant ready")
	return cmd
}
