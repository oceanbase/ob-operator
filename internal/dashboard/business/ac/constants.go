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

var AllPolicies = []acmodel.Policy{
	{
		Subject: acmodel.Subject("OBCluster"),
		Action:  acmodel.ActionRead,
	},
	{
		Subject: acmodel.Subject("OBCluster"),
		Action:  acmodel.ActionWrite,
	},
	{
		Subject: acmodel.Subject("OBTenant"),
		Action:  acmodel.ActionRead,
	},
	{
		Subject: acmodel.Subject("OBTenant"),
		Action:  acmodel.ActionWrite,
	},
	{
		Subject: acmodel.Subject("Alarm"),
		Action:  acmodel.ActionRead,
	},
	{
		Subject: acmodel.Subject("Alarm"),
		Action:  acmodel.ActionWrite,
	},
	{
		Subject: acmodel.Subject("OBProxy"),
		Action:  acmodel.ActionRead,
	},
	{
		Subject: acmodel.Subject("OBProxy"),
		Action:  acmodel.ActionWrite,
	},
	{
		Subject: acmodel.Subject("AccessControl"),
		Action:  acmodel.ActionRead,
	},
	{
		Subject: acmodel.Subject("AccessControl"),
		Action:  acmodel.ActionWrite,
	},
}
