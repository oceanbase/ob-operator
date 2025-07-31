/*
Copyright (c) 2025 OceanBase
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
	"context"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	obclusterclient "github.com/oceanbase/ob-operator/internal/clients"
	oceanbase "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	k8sclient "github.com/oceanbase/ob-operator/pkg/k8s/client"
)

var (
	namespace string
	cluster   string
	output    string
)

func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate configuration files",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	cmd.AddCommand(newGenerateOBDiagConfigCmd())
	return cmd
}

func newGenerateOBDiagConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "obdiag-config",
		Short: "Generate ob-diag configuration file",
		RunE:  generateOBDiagConfig,
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace where the obcluster object is in")
	cmd.Flags().StringVarP(&cluster, "cluster", "c", "", "obcluster object's name")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file path")
	cmd.MarkFlagRequired("namespace")
	cmd.MarkFlagRequired("cluster")
	cmd.MarkFlagRequired("output")
	return cmd
}

func generateOBDiagConfig(cmd *cobra.Command, args []string) error {
	clientset := k8sclient.GetClient().ClientSet

	obcluster, err := obclusterclient.GetOBCluster(context.Background(), namespace, cluster)
	if err != nil {
		return err
	}

	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), obcluster.Spec.UserSecrets.Root, metav1.GetOptions{})
	if err != nil {
		return err
	}
	password := string(secret.Data["password"])

	observers, err := obclusterclient.ListOBServersOfOBCluster(context.Background(), obcluster)
	if err != nil {
		return err
	}

	var nodes []NodeConfig
	var dbHost string
	var dbPort int

	k8sConfigDir := output + "/k8s"
	if err := os.MkdirAll(k8sConfigDir, 0755); err != nil {
		return err
	}

	for _, observer := range observers.Items {

		var ip string
		annotationMode, ok := obcluster.Annotations[oceanbase.AnnotationsMode]
		if ok && annotationMode == oceanbase.ModeService {
			ip = observer.Status.ServiceIp
			if dbHost == "" {
				dbHost = observer.Status.ServiceIp
				dbPort = oceanbase.SqlPort
			}
		} else {
			ip = observer.Status.PodIp
			if dbHost == "" {
				dbHost = observer.Status.PodIp
				dbPort = oceanbase.SqlPort
			}
		}

		nodeConfig := NodeConfig{
			PodName: observer.Name,
			IP:      ip,
		}

		zone := observer.Spec.Zone
		for _, obzone := range obcluster.Spec.Topology {
			if obzone.Zone != zone || obzone.K8sCluster == "" {
				continue
			}
			k8sCluster, err := obclusterclient.GetK8sCluster(context.Background(), obzone.K8sCluster)
			if err != nil {
				return err
			}
			config, err := k8sCluster.DecodeKubeConfig()
			if err != nil {
				return err
			}
			configPath := k8sConfigDir + "/" + obzone.K8sCluster + ".yaml"
			if err := os.WriteFile(configPath, config, 0644); err != nil {
				return err
			}
			nodeConfig.KubernetesConfigFile = configPath
		}
		nodes = append(nodes, nodeConfig)
	}

	diagConfig := &OBDiagConfig{
		OBCluster: OBClusterConfig{
			DBHost:        dbHost,
			DBPort:        dbPort,
			OBClusterName: obcluster.Spec.ClusterName,
			TenantSys: TenantSysConfig{
				User:     "root@sys",
				Password: password,
			},
			Servers: ServerConfig{
				Nodes: nodes,
				Global: GlobalConfig{
					Namespace:     namespace,
					SshType:       "kubernetes",
					ContainerName: "observer",
					HomePath:      "/home/admin/oceanbase",
					DataDir:       "/home/admin/data-file",
					RedoDir:       "/home/admin/data-log",
				},
			},
		},
	}

	yamlData, err := yaml.Marshal(diagConfig)
	if err != nil {
		return err
	}

	return os.WriteFile(output+"/config.yaml", yamlData, 0644)
}

func init() {
	rootCmd.AddCommand(newGenerateCmd())
}
