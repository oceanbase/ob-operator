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
	"fmt"

	"github.com/spf13/cobra"

	param "github.com/oceanbase/ob-operator/internal/dashboard/model/param"
)

type ScaleOptions struct {
	BaseOptions
	Zones    map[string]string    `json:"zones"`
	Topology []param.ZoneTopology `json:"topology"`
	ZoneNum  int                  `json:"zone_num"`
}

func NewScaleOptions() *ScaleOptions {
	return &ScaleOptions{}
}
func (o *ScaleOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Namespace, "namespace", "default", "namespace of ob cluster")
	cmd.Flags().StringToStringVar(&o.Zones, "zones", nil, "zone of ob cluster")
}
func (o *ScaleOptions) Parse() error {
	topology, err := mapZonesToTopology(o.Zones)
	if err != nil {
		return err
	}
	o.Topology = topology
	return nil
}
func (o *ScaleOptions) Validate() error {
	// Ensure obcluster has at least 2 zones
	deleteNum := 0
	for _, zone := range o.Topology {
		if zone.Replicas == 0 {
			deleteNum++
		}
		if o.ZoneNum-deleteNum < 2 {
			return fmt.Errorf("Obcluster has %d Zones, can only delete %d zones", o.ZoneNum, o.ZoneNum-2)
		}
	}

	return nil
}
