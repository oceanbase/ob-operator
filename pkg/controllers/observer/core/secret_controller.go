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
	"context"

	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
)

func (ctrl *OBClusterCtrl) CreateDBUserSecret(tenantName, userName, password string) error {
	secret := converter.GenerateDBUserSecret(ctrl.OBCluster, tenantName, userName, password)
	secretExecutor := resource.NewSecretResource(ctrl.Resource)
	err := secretExecutor.Create(context.TODO(), secret)
	if err != nil {
		return err
	}
	return nil
}
