package cluster

import (
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/spf13/cobra"
)

type ShowOptions struct {
	BaseOptions
	Obcluster          *v1alpha1.OBCluster
	ObclusterOperation *v1alpha1.OBClusterOperationList
}

func NewShowOptions() *ShowOptions {
	return &ShowOptions{}
}
func (o *ShowOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Namespace, "namespace", "default", "namespace of ob cluster")
}
