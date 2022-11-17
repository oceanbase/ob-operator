/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package converter

import (
	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/model"
)

func JudgeRootserviceStatusByObj(rsList []model.AllVirtualCoreMeta, rs cloudv1.RootService) bool {
	for _, cluster := range rs.Status.Topology {
		for _, zone := range cluster.Zone {
			nodeIsExists := false
			if zone.ServerIP == "" {
				continue
			}
			for _, node := range rsList {
				if node.SvrIP == zone.ServerIP && node.Role == zone.Role {
					nodeIsExists = true
					break
				}
			}
			if !nodeIsExists {
				return false
			}
		}
	}
	return true
}
