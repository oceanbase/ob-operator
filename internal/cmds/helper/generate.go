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

package helper

import (
	"context"
	"io/ioutil"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	sigclient "sigs.k8s.io/controller-runtime/pkg/client"
	runtime "k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	oceanbasev1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

var (scheme = runtime.NewScheme())

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = oceanbasev1alpha1.AddToScheme(scheme)
	rootCmd.AddCommand(newGenerateCmd())
}

type OBDiagConfig struct {
	Global struct {
		Namespace string `yaml:"namespace"`
	} `yaml:"global"`
	Servers []struct {
		IP   string `yaml:"ip"`
		Port int    `yaml:"port"`
	} `yaml:"servers"`
	TenantSys struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"tenant_sys"`
	DBHost        string `yaml:"db_host"`
	DBPort        int    `yaml:"db_port"`
	OBClusterName string `yaml:"ob_cluster_name"`
}

var (
	namespace string
	cluster   string
	output    string
)

func init() {
	rootCmd.AddCommand(newGenerateCmd())
}

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
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	k8sClient, err := sigclient.New(config, sigclient.Options{Scheme: scheme})
	if err != nil {
		return err
	}

	obCluster := &oceanbasev1alpha1.OBCluster{}
	if err := k8sClient.Get(context.Background(), sigclient.ObjectKey{Namespace: namespace, Name: cluster}, obCluster); err != nil {
		return err
	}

	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), obCluster.Spec.UserSecrets.Root, metav1.GetOptions{})
	if err != nil {
		return err
	}
	password := string(secret.Data["password"])

	observers := &oceanbasev1alpha1.OBServerList{}
	if err := k8sClient.List(context.Background(), observers, sigclient.InNamespace(namespace), sigclient.MatchingLabels{oceanbase.LabelRefOBCluster: cluster}); err != nil {
		return err
	}

	var servers []struct {
		IP   string `yaml:"ip"`
		Port int    `yaml:"port"`
	}
	var dbHost string
	var dbPort int

	for _, observer := range observers.Items {
		if obCluster.Annotations["oceanbase.oceanbase.com/mode"] == "service" {
			svc, err := clientset.CoreV1().Services(namespace).Get(context.Background(), observer.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			servers = append(servers, struct {
				IP   string `yaml:"ip"`
				Port int    `yaml:"port"`
			}{
				IP:   svc.Spec.ClusterIP,
				Port: 2881,
			})
			if dbHost == "" {
				dbHost = svc.Spec.ClusterIP
				dbPort = 2881
			}
		} else {
			pod, err := clientset.CoreV1().Pods(namespace).Get(context.Background(), observer.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			servers = append(servers, struct {
				IP   string `yaml:"ip"`
				Port int    `yaml:"port"`
			}{
				IP:   pod.Status.PodIP,
				Port: 2881,
			})
			if dbHost == "" {
				dbHost = pod.Status.PodIP
				dbPort = 2881
			}
		}
	}

	diagConfig := &OBDiagConfig{
		Global: struct {
			Namespace string `yaml:"namespace"`
		}{
			Namespace: namespace,
		},
		Servers: servers,
		TenantSys: struct {
			User     string `yaml:"user"`
			Password string `yaml:"password"`
		}{
			User:     "root@sys",
			Password: password,
		},
		DBHost:        dbHost,
		DBPort:        dbPort,
		OBClusterName: cluster,
	}

	yamlData, err := yaml.Marshal(diagConfig)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(output, yamlData, 0644)
}
