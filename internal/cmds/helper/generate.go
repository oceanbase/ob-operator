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

	for _, observer := range observers.Items {
		var ip string
		annotationMode, ok := obcluster.Annotations[oceanbase.AnnotationsMode]
		if ok && annotationMode == oceanbase.ModeService {
			svc, err := clientset.CoreV1().Services(namespace).Get(context.Background(), observer.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			ip = svc.Spec.ClusterIP
			if dbHost == "" {
				dbHost = svc.Spec.ClusterIP
				dbPort = 2881
			}
		} else {
			pod, err := clientset.CoreV1().Pods(namespace).Get(context.Background(), observer.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			ip = pod.Status.PodIP
			if dbHost == "" {
				dbHost = pod.Status.PodIP
				dbPort = 2881
			}
		}
		nodes = append(nodes, NodeConfig{
			PodName: observer.Name,
			IP:      ip,
		})
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
					DataDir:       "/home/admin/oceanbase/store",
					RedoDir:       "/home/admin/oceanbase/store",
				},
			},
		},
	}

	yamlData, err := yaml.Marshal(diagConfig)
	if err != nil {
		return err
	}

	return os.WriteFile(output, yamlData, 0644)
}

func init() {
	rootCmd.AddCommand(newGenerateCmd())
}