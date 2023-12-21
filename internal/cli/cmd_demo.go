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

package cli

import (
	"errors"
	"log"
	"regexp"

	"github.com/oceanbase/ob-operator/internal/cli/helpers"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/tools/record/util"
)

func init() {
	rootCmd.AddCommand(demoCmd)

	demoCmd.AddCommand(demoSetupCmd)
	demoSetupCmd.AddCommand(singleNodeCmd)
}

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Manage demos for OceanBase Operator",
}

var demoSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up demos for OceanBase Operator",
}

var singleNodeCmd = &cobra.Command{
	Use:   "single-node",
	Short: "Set up a single-node OceanBase cluster",
	Run: func(cmd *cobra.Command, args []string) {
		pattern := regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")
		prompt := promptui.Prompt{
			Label:     "Enter cluster name: ",
			Templates: textTmpls,
			Validate: func(input string) error {
				if pattern.MatchString(input) {
					return nil
				}
				return errors.New("invalid cluster name, required pattern is [a-z0-9]([-a-z0-9]*[a-z0-9])?")
			},
		}

		name, err := prompt.Run()
		if err != nil {
			log.Println(err.Error())
			return
		}

		cluster := helpers.NewOBCluster(name, 1, 1)

		secrets := []string{cluster.Spec.UserSecrets.Root, cluster.Spec.UserSecrets.ProxyRO, cluster.Spec.UserSecrets.Monitor, cluster.Spec.UserSecrets.Operator}
		for _, secret := range secrets {
			if _, err := clientset.CoreV1().Secrets(cluster.GetNamespace()).Get(cmd.Context(), secret, metav1.GetOptions{}); err != nil {
				if util.IsKeyNotFoundError(err) {
					_, err := clientset.CoreV1().Secrets(cluster.GetNamespace()).Create(cmd.Context(), &v1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      secret,
							Namespace: cluster.GetNamespace(),
						},
						StringData: map[string]string{
							"password": rand.String(16),
						},
					}, metav1.CreateOptions{})
					if err != nil {
						log.Println(err.Error())
						return
					}
				}
			}
		}

		gvk := cluster.GroupVersionKind()
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			log.Println(err.Error())
			return
		}
		unsContent, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cluster)
		if err != nil {
			log.Println(err.Error())
			return
		}
		uns := &unstructured.Unstructured{Object: unsContent}
		_, err = dynamicClient.Resource(mapping.Resource).Namespace("default").Create(cmd.Context(), uns, metav1.CreateOptions{})
		if err != nil {
			log.Println(err.Error())
			return
		}
		log.Println("single node cluster created successfully")
	},
}
