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
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	"github.com/spf13/cobra"
)

type CreateOptions struct {
	Namespace    string `json:"namespace"`
	Name         string `json:"name"`
	ClusterName  string `json:"clusterName"`
	ClusterId    int64  `json:"clusterId"`
	Image        string `json:"image"`
	CPU          int64  `json:"cpu"`
	Memory       int64  `json:"memory"`
	RootPassword string `json:"rootPassword"`
	Zones        []string
	Parameters   []common.KVPair `json:"parameters"`
	Mode         string          `json:"mode"`
}

func NewCreateOptions() *CreateOptions {
	return &CreateOptions{}
}

// Create an OBClusterInstance
func CreateOBClusterInstance(o *CreateOptions) *v1alpha1.OBCluster {

	// obcluster := &v1alpha1.OBCluster{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Namespace:   o.Namespace,
	// 		Name:        o.Name,
	// 		Annotations: map[string]string{},
	// 	},
	// 	Spec: v1alpha1.OBClusterSpec{
	// 		ClusterName:      o.ClusterName,
	// 		ClusterId:        o.ClusterId,
	// 		OBServerTemplate: observerTemplate,
	// 		MonitorTemplate:  monitorTemplate,
	// 		BackupVolume:     backupVolume,
	// 		Parameters:       parameters,
	// 		Topology:         topology,
	// 		UserSecrets:      generateUserSecrets(o.Name, o.ClusterId),
	// 	},
	// }
	// return obcluster
	return nil
}
func (o *CreateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Name, "name", "test", "The name in k8s")
	cmd.Flags().StringArrayVar(&o.Zones, "zones", []string{}, "List of zones in the format 'key=value'")
	cmd.Flags().StringVar(&o.RootPassword, "root-password", "root-password", "The root password of the cluster")
	cmd.Flags().StringVar(&o.Namespace, "namespace", "default", "The namespace of the cluster")
	cmd.Flags().StringVar(&o.Mode, "mode", "normal", "The mode of the cluster")
}
