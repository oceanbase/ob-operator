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

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	rootCmd.AddCommand(clusterCmd)
	clusterCmd.AddCommand(clusterListCmd)
}

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage OBCluster resources",
	Run:   func(cmd *cobra.Command, args []string) {},
}

var clusterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List OBCluster resources",
	Run: func(cmd *cobra.Command, args []string) {
		podList, err := clientset.CoreV1().Pods("monitoring").List(cmd.Context(), v1.ListOptions{})
		if err != nil {
			panic(err)
		}
		for _, pod := range podList.Items {
			_, _ = fmt.Println(pod.Name, pod.CreationTimestamp, pod.Status.Phase)
		}
	},
}
