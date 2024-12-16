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
	"context"
	"log"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/cli/cluster"
	"github.com/oceanbase/ob-operator/internal/cli/demo"
	"github.com/oceanbase/ob-operator/internal/cli/tenant"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
	"github.com/oceanbase/ob-operator/internal/clients"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	"github.com/oceanbase/ob-operator/internal/const/status/tenantstatus"
)

var defaultTimeoutDuration = 30 * time.Minute

// NewCmd create demo command for cluster creation
func NewCmd(timeoutDuration time.Duration) *cobra.Command {
	clusterOptions := cluster.NewCreateOptions()
	tenantOptions := tenant.NewCreateOptions()
	logger := utils.GetDefaultLoggerInstance()
	pf := demo.NewPromptFactory()
	clusterTickerDuration := 2 * time.Second
	tenantTickerDuration := 1 * time.Second
	var clusterType string
	var wait bool
	var err error
	var prompt any
	if timeoutDuration == 0 {
		timeoutDuration = defaultTimeoutDuration
	}
	cmd := &cobra.Command{
		Use:   "demo <subcommand>",
		Short: "deploy demo ob cluster and tenant in easier way",
		Long:  `deploy demo ob cluster and tenant in easier way, currently support single node and three node cluster, with corresponding tenant`,
		Args:  cobra.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			// prompt for cluster create options
			for {
				prompt = pf.CreatePrompt(cluster.FLAG_NAME)
				if clusterOptions.Name, err = pf.RunPromptE(prompt); err != nil {
					logger.Fatalln(err)
				}
				prompt = pf.CreatePrompt(cluster.FLAG_NAMESPACE)
				if clusterOptions.Namespace, err = pf.RunPromptE(prompt); err != nil {
					logger.Fatalln(err)
				}
				if !utils.CheckIfClusterExists(cmd.Context(), clusterOptions.Name, clusterOptions.Namespace) {
					break
				}
				logger.Printf("Cluster %s already exists in namespace %s, please input another cluster name", clusterOptions.Name, clusterOptions.Namespace)
			}
			prompt = pf.CreatePrompt(cluster.CLUSTER_TYPE)
			if clusterType, err = pf.RunPromptE(prompt); err != nil {
				logger.Fatalln(err)
			}
			prompt = pf.CreatePrompt(cluster.FLAG_ROOT_PASSWORD)
			if clusterOptions.RootPassword, err = pf.RunPromptE(prompt); err != nil {
				logger.Fatalln(err)
			}

			// prompt for tenant create options
			for {
				prompt = pf.CreatePrompt(tenant.FLAG_TENANT_NAME_IN_K8S)
				if tenantOptions.Name, err = pf.RunPromptE(prompt); err != nil {
					logger.Fatalln(err)
				}
				if !utils.CheckIfTenantExists(cmd.Context(), tenantOptions.Name, clusterOptions.Namespace) {
					break
				}
				logger.Printf("Tenant %s already exists in namespace %s, please input another tenant name", tenantOptions.Name, clusterOptions.Namespace)
			}
			prompt = pf.CreatePrompt(tenant.FLAG_TENANT_NAME)
			if tenantOptions.TenantName, err = pf.RunPromptE(prompt); err != nil {
				logger.Fatalln(err)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
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
			obcluster, err := cluster.CreateOBCluster(cmd.Context(), clusterOptions)
			if err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Creating OBCluster instance: %s", clusterOptions.ClusterName)
			waitForClusterReady(cmd.Context(), obcluster, logger, timeoutDuration, clusterTickerDuration)
			logger.Printf("Run `echo $(kubectl get secret %s -n %s -o jsonpath='{.data.password}'|base64 --decode)` to get cluster secrets", obcluster.Spec.UserSecrets.Root, obcluster.Namespace)
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			// create tenant after cluster ready
			obtenant, err := tenant.CreateOBTenant(cmd.Context(), tenantOptions)
			if err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Creating OBTenant instance: %s", tenantOptions.TenantName)
			waitForTenantReady(cmd.Context(), obtenant, logger, timeoutDuration, tenantTickerDuration)
			logger.Printf("Run `echo $(kubectl get secret %s -n %s -o jsonpath='{.data.password}'|base64 --decode)` to get tenant secrets", obtenant.Spec.Credentials.Root, obtenant.Namespace)
		},
	}
	// TODO: if w is set, wait for cluster and tenant ready
	cmd.Flags().BoolVarP(&wait, cluster.FLAG_WAIT, "w", cluster.DEFAULT_WAIT, "wait for the cluster and tenant ready")
	return cmd
}

// waitForTenantReady wait for tenant ready, log the task status
func waitForClusterReady(ctx context.Context, obcluster *v1alpha1.OBCluster, logger *log.Logger, timeoutDuration time.Duration, tickerDuration time.Duration) {
	var err error
	lastTask := ""
	lastTaskStatus := ""
	timeout := time.NewTimer(timeoutDuration)
	ticker := time.NewTicker(tickerDuration)
	defer ticker.Stop()
	defer timeout.Stop()
	logger.Println("Waiting for cluster ready...")
	for obcluster.Status.Status != clusterstatus.Running {
		select {
		case <-ticker.C:
			obcluster, err = clients.GetOBCluster(ctx, obcluster.Namespace, obcluster.Name)
			if err != nil {
				logger.Fatalln(err)
			}
			if obcluster.Status.OperationContext != nil && (lastTask != string(obcluster.Status.OperationContext.Task) || lastTaskStatus != string(obcluster.Status.OperationContext.TaskStatus)) {
				logger.Printf("Task: %s, Status: %s", obcluster.Status.OperationContext.Task, obcluster.Status.OperationContext.TaskStatus)
				lastTask = string(obcluster.Status.OperationContext.Task)
				lastTaskStatus = string(obcluster.Status.OperationContext.TaskStatus)
				timeout.Reset(timeoutDuration)
			}
		case <-timeout.C:
			logger.Fatalf("Task: %s timeout", lastTask)
		}
	}
	logger.Println("Create Cluster successfully")
}

// waitForTenantReady wait for tenant ready, log the task status
func waitForTenantReady(ctx context.Context, obtenant *v1alpha1.OBTenant, logger *log.Logger, timeoutDuration time.Duration, tickerDuration time.Duration) {
	var err error
	lastTask := ""
	lastTaskStatus := ""
	timeout := time.NewTimer(timeoutDuration)
	ticker := time.NewTicker(tickerDuration)
	defer ticker.Stop()
	defer timeout.Stop()
	logger.Println("Waiting for tenant ready...")
	for obtenant.Status.Status != tenantstatus.Running {
		select {
		case <-ticker.C:
			obtenant, err = clients.GetOBTenant(ctx, types.NamespacedName{Namespace: obtenant.Namespace, Name: obtenant.Name})
			if err != nil {
				logger.Fatalln(err)
			}
			if obtenant.Status.OperationContext != nil && (lastTask != string(obtenant.Status.OperationContext.Task) || lastTaskStatus != string(obtenant.Status.OperationContext.TaskStatus)) {
				lastTask = string(obtenant.Status.OperationContext.Task)
				lastTaskStatus = string(obtenant.Status.OperationContext.TaskStatus)
				timeout.Reset(timeoutDuration)
			}
		case <-timeout.C:
			logger.Fatalf("Task: %s timeout", lastTask)
		}
	}
	logger.Println("Create Tenant instance successfully")
}
