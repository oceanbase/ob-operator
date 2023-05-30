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

package task

import (
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
)

// register all task flows at init
func init() {
	// obcluster
	GetRegistry().Register(flowname.BootstrapOBCluster, BootstrapOBCluster)
	GetRegistry().Register(flowname.MaintainOBClusterAfterBootstrap, MaintainOBClusterAfterBootstrap)
	GetRegistry().Register(flowname.AddOBZone, AddOBZone)

	// obzone
	GetRegistry().Register(flowname.CreateOBZone, CreateOBZone)
	GetRegistry().Register(flowname.PrepareOBZoneForBootstrap, PrepareOBZoneForBootstrap)
	GetRegistry().Register(flowname.MaintainOBZoneAfterBootstrap, MaintainOBZoneAfterBootstrap)

	// observer
	GetRegistry().Register(flowname.CreateOBServer, CreateOBServer)
	GetRegistry().Register(flowname.PrepareOBServerForBootstrap, PrepareOBServerForBootstrap)
	GetRegistry().Register(flowname.MaintainOBServerAfterBootstrap, MaintainOBServerAfterBootstrap)
}
