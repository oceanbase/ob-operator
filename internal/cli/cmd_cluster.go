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
	"fmt"
	"log"

	"github.com/oceanbase/ob-operator/api/v1alpha1"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	watch *bool
	ns    *string
)

func init() {
	rootCmd.AddCommand(clusterCmd)
	watch = clusterCmd.PersistentFlags().BoolP("watch", "w", false, "watch for changes")
	ns = clusterCmd.PersistentFlags().StringP("namespace", "n", v1.NamespaceDefault, "namespace to use")

	clusterCmd.AddCommand(clusterListCmd, clusterDeleteCmd)
}

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage OBCluster resources",
}

var clusterListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "l"},
	Short:   "List OBCluster resources",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.ParseFlags(args); err != nil {
			log.Println(err.Error())
			return
		}
		cluster := v1alpha1.OBCluster{
			TypeMeta: v1.TypeMeta{
				Kind:       "OBCluster",
				APIVersion: "oceanbase.oceanbase.com/v1alpha1",
			},
		}
		gvk := cluster.GetObjectKind().GroupVersionKind()
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			log.Println(err.Error())
			return
		}
		unsList, err := dynamicClient.Resource(mapping.Resource).List(cmd.Context(), v1.ListOptions{})
		if err != nil {
			log.Println(err.Error())
			return
		}

		if len(unsList.Items) == 0 {
			log.Println("No pods found")
			return
		}

		tbLog.Println("Namespace \t Pod \t Creation Time \t Status")
		for _, pod := range unsList.Items {
			obj := &v1alpha1.OBCluster{}
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(pod.UnstructuredContent(), obj)
			if err != nil {
				log.Println(err.Error())
				return
			}
			tbLog.Printf("%s \t %s \t %s \t %s\n", obj.Namespace, obj.Name, obj.CreationTimestamp, obj.Status.Status)
		}
		_ = tbw.Flush()
	},
}

var clusterDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an existing cluster",
	Run: func(cmd *cobra.Command, args []string) {
		cluster := v1alpha1.OBCluster{
			TypeMeta: v1.TypeMeta{
				Kind:       "OBCluster",
				APIVersion: "oceanbase.oceanbase.com/v1alpha1",
			},
		}
		gvk := cluster.GetObjectKind().GroupVersionKind()
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			log.Println(err.Error())
			return
		}

		if len(args) == 0 {
			unsList, err := dynamicClient.Resource(mapping.Resource).Namespace(v1.NamespaceAll).List(cmd.Context(), v1.ListOptions{})
			if err != nil {
				log.Println(err.Error())
				return
			}
			clusterList := v1alpha1.OBClusterList{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(unsList.UnstructuredContent(), &clusterList)
			if err != nil {
				log.Println(err.Error())
				return
			}
			items := []string{}
			for _, cluster := range clusterList.Items {
				items = append(items, fmt.Sprintf("%s (ns: %s)", cluster.Name, cluster.Namespace))
			}
			sl := promptui.Select{
				Label: "Select a cluster to delete",
				Items: items,
			}
			idx, value, err := sl.Run()
			if err != nil {
				log.Println(err.Error())
				return
			}
			log.Printf("You choose #%d: %s\n", idx, value)
			confirm := promptui.Prompt{
				Label:     "Are you sure you want to delete this cluster? > " + value,
				IsConfirm: true,
			}
			if _, err = confirm.Run(); err != nil {
				return
			}

			deletingCluster := clusterList.Items[idx]
			if err := deleteObj(cmd.Context(), &deletingCluster); err != nil {
				log.Println(err.Error())
				return
			}
			log.Println("cluster deleted successfully")
		} else {
			for _, name := range args {
				cluster.Name = name
				cluster.Namespace = *ns
				if _, err := getObj(cmd.Context(), &cluster); err != nil {
					log.Println(err.Error())
					return
				}
				conf := promptui.Prompt{
					Label:     "Are you sure you want to delete this cluster? > " + name,
					IsConfirm: true,
				}
				if _, err = conf.Run(); err != nil {
					continue
				}
				if err := deleteObj(cmd.Context(), &cluster); err != nil {
					log.Println(err.Error())
					return
				}
				log.Println("cluster deleted successfully")
			}
		}
	},
}
