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
	ActionRead  acmodel.Action = "READ"
	ActionWrite acmodel.Action = "WRITE"
	ActionAll   acmodel.Action = "ALL"
)

const (
	DomainOBCluster     acmodel.Domain = "OBCluster"
	DomainOBTenant      acmodel.Domain = "OBTenant"
	DomainAlarm         acmodel.Domain = "Alarm"
	DomainOBProxy       acmodel.Domain = "OBProxy"
	DomainAccessControl acmodel.Domain = "AccessControl"
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
		Domain: DomainOBTenant,
		Action: ActionRead,
	},
	{
		Domain: DomainOBTenant,
		Action: ActionWrite,
	},
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
		Domain: DomainAccessControl,
		Action: ActionRead,
	},
	{
		Domain: DomainAccessControl,
		Action: ActionWrite,
	},
}
