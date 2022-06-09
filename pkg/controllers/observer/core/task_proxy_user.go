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

package core

import (
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/sql"
)

func (ctrl *OBClusterCtrl) CreateUserForObproxy() error {
	clusterIP, err := ctrl.GetServiceClusterIPByName(ctrl.OBCluster.Namespace, ctrl.OBCluster.Name)
	if err != nil {
		return err
	}
	err = sql.CreateUser(clusterIP, "proxyro", "")
	if err != nil {
		return err
	}
	err = sql.GrantPrivilege(clusterIP, "select", "*.*", "proxyro")
	if err != nil {
		return err
	}
	return nil
}
