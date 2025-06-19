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

package ac

import (
	acmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/ac"
)

const (
	ActionRead  acmodel.Action = "read"
	ActionWrite acmodel.Action = "write"
	ActionAll   acmodel.Action = "*"
)

const (
	DomainAc         acmodel.Domain = "ac"
	DomainAlarm      acmodel.Domain = "alarm"
	DomainSystem     acmodel.Domain = "system"
	DomainK8sCluster acmodel.Domain = "k8s-cluster"
	DomainOBCluster  acmodel.Domain = "obcluster"
	DomainOBTenant   acmodel.Domain = "obtenant"
	DomainOBProxy    acmodel.Domain = "obproxy"
)

var AllPolicies = []acmodel.Policy{
	{
		Domain: DomainOBCluster,
		Action: ActionRead,
	},
	{
		Domain: DomainOBCluster,
		Action: ActionWrite,
	},
	{
		Domain: DomainSystem,
		Action: ActionRead,
	},
	{
		Domain: DomainSystem,
		Action: ActionWrite,
	},
	// {
	// 	Domain: DomainOBTenant,
	// 	Action: ActionRead,
	// },
	// {
	// 	Domain: DomainOBTenant,
	// 	Action: ActionWrite,
	// },
	{
		Domain: DomainAlarm,
		Action: ActionRead,
	},
	{
		Domain: DomainAlarm,
		Action: ActionWrite,
	},
	{
		Domain: DomainOBProxy,
		Action: ActionRead,
	},
	{
		Domain: DomainOBProxy,
		Action: ActionWrite,
	},
	{
		Domain: DomainAc,
		Action: ActionRead,
	},
	{
		Domain: DomainAc,
		Action: ActionWrite,
	},
	{
		Domain: DomainK8sCluster,
		Action: ActionRead,
	},
	{
		Domain: DomainK8sCluster,
		Action: ActionWrite,
	},
}
