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
package cluster

import (
	"sort"

	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/clients"
	"github.com/spf13/cobra"
)

// NewListCmd list all ob clusters
func NewListCmd() *cobra.Command {
	tbw, tbLog := cmdUtil.GetTableLoggerInstance()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List ob clusters",
		Long:    `List ob clusters.`,
		Run: func(cmd *cobra.Command, args []string) {
			obclusterList, err := clients.ListAllOBClusters(cmd.Context())
			if err != nil {
				logger.Fatalln(err.Error())
			}
			sort.Slice(obclusterList.Items, func(i, j int) bool {
				return obclusterList.Items[i].Name < obclusterList.Items[j].Name
			})
			if len(obclusterList.Items) == 0 {
				logger.Println("No clusters found")
				return
			}
			tbLog.Println("Namespace \t Pod \t Create Time \t Status")
			for _, cluster := range obclusterList.Items {
				if err != nil {
					logger.Println(err.Error())
					return
				}
				tbLog.Printf("%s \t %s \t %s \t %s\n", cluster.Namespace, cluster.Name, cluster.CreationTimestamp, cluster.Status.Status)
			}
			_ = tbw.Flush()
		},
	}
	return cmd
}
