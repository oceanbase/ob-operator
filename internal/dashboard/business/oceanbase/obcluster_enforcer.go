/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package oceanbase

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	acbiz "github.com/oceanbase/ob-operator/internal/dashboard/business/ac"
	acmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/ac"
)

// filterTenants filters tenants by the user's permission
func filterClusters(username, action string, list *v1alpha1.OBClusterList) *v1alpha1.OBClusterList {
	newList := []v1alpha1.OBCluster{}
	for i, c := range list.Items {
		ok, err := acbiz.Enforce(context.TODO(), username, &acmodel.Policy{
			Domain: acbiz.DomainOBCluster,
			Object: acmodel.Object(c.Name),
			Action: acmodel.Action(action),
		})
		if err != nil {
			logrus.Error(err)
			continue
		}
		logrus.Debugf("enforce user %s for cluster %s is %t", username, c.Name, ok)
		if ok {
			newList = append(newList, list.Items[i])
		}
	}
	list.Items = newList
	return list
}
