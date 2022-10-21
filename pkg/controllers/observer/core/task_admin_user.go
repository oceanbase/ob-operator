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
	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-password/password"
	"k8s.io/klog/v2"
)

func (ctrl *OBClusterCtrl) CreateAdminUser(statefulApp cloudv1.StatefulApp) error {
	pwd, err := password.Generate(16, 4, 0, false, false)
	if err != nil {
		return err
	}

	klog.Info("generated password: %s", pwd)

	sqlOperator, err := ctrl.GetSqlOperatorFromStatefulApp(statefulApp)
	if err != nil {
		return errors.Wrap(err, "get sql operator when create user for operation")
	}
	err = sqlOperator.CreateUser("admin", pwd)
	if err != nil {
		return err
	}
	err = sqlOperator.GrantPrivilege("all", "*.*", "admin")
	if err != nil {
		return err
	}

	ctrl.CreateDBUserSecret("sys", "admin", pwd)
	if err != nil {
		return err
	}

	return nil
}
