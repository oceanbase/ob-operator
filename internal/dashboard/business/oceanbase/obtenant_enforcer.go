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

package oceanbase

import (
	"context"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/clients"
	acbiz "github.com/oceanbase/ob-operator/internal/dashboard/business/ac"
	acmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/ac"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
)

// TenantGuard checks if the user has permission to access cluster which the tenant belongs to
func TenantGuard(ns, name, action string) acbiz.EnforceFunc {
	return func(c *gin.Context) (bool, error) {
		session := sessions.Default(c)
		usernameIf := session.Get("username")
		if usernameIf == nil {
			return false, httpErr.New(httpErr.ErrUnauthorized, "Unauthorized")
		}
		username := usernameIf.(string)

		finalNs := ns
		finalName := name
		if strings.HasPrefix(ns, ":") {
			finalNs = c.Param(ns[1:])
		}
		if strings.HasPrefix(name, ":") {
			finalName = c.Param(name[1:])
		}
		t, err := clients.TenantClient.Get(c, finalNs, finalName, v1.GetOptions{})
		if err != nil {
			return false, err
		}

		return acbiz.Enforce(c, username, &acmodel.Policy{
			Domain: acbiz.DomainOBCluster,
			Object: acmodel.Object(t.Spec.ClusterName),
			Action: acmodel.Action(action),
		})
	}
}

// filterTenants filters tenants by the user's permission
func filterTenants(username, action string, list *v1alpha1.OBTenantList) *v1alpha1.OBTenantList {
	newList := []v1alpha1.OBTenant{}
	for i, t := range list.Items {
		ok, err := acbiz.Enforce(context.TODO(), username, &acmodel.Policy{
			Domain: acbiz.DomainOBCluster,
			Object: acmodel.Object(t.Spec.ClusterName),
			Action: acmodel.Action(action),
		})
		if err != nil {
			continue
		}
		if ok {
			newList = append(newList, list.Items[i])
		}
	}
	list.Items = newList
	return list
}
