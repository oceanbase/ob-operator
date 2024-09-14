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
package util

import (
	"fmt"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	"github.com/oceanbase/ob-operator/internal/const/status/tenantstatus"
)

// CheckTenantStatus check running status of obtenant
func CheckTenantStatus(tenant *v1alpha1.OBTenant) error {
	if tenant.Status.Status != tenantstatus.Running {
		return fmt.Errorf("Obtenant status invalid, Status:%s", tenant.Status.Status)
	}
	return nil
}

// CheckClusterStatus check running status of obcluster
func CheckClusterStatus(cluster *v1alpha1.OBCluster) error {
	if cluster.Status.Status != clusterstatus.Running {
		return fmt.Errorf("Obcluster status invalid, Status:%s", cluster.Status.Status)
	}
	return nil
}
