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
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

func (ctrl *TenantCtrl) DeleteTenant() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Deleting Tenant ", ctrl.Tenant.Name))
	}

	tenantExist, _, err := ctrl.TenantExist(ctrl.Tenant.Name)
	if err != nil {
		klog.Errorln("Check Whether The Tenant Exists Error: ", err)
		return err
	}
	if tenantExist {
		return sqlOperator.DeleteTenant(ctrl.Tenant.Name)
	}
	return nil
}

func (ctrl *TenantCtrl) DeletePool() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Deleting Pool", ctrl.Tenant.Name))
	}
	for _, zone := range ctrl.Tenant.Spec.Topology {
		poolName := ctrl.GeneratePoolName(ctrl.Tenant.Name, zone.ZoneName)
		poolExist, _, err := ctrl.PoolExist(poolName)
		if err != nil {
			klog.Errorln("Check Whether The Resource Pool Exists Error: ", err)
			return err
		}
		if poolExist {
			err = sqlOperator.DeletePool(poolName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ctrl *TenantCtrl) DeleteUnit() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Deleting Unit", ctrl.Tenant.Name))
	}
	for _, zone := range ctrl.Tenant.Spec.Topology {
		unitName := ctrl.GenerateUnitName(ctrl.Tenant.Name, zone.ZoneName)
		err, unitExist := ctrl.UnitExist(unitName)
		if err != nil {
			klog.Errorln("Check Whether The Resource Unit Exists Error: ", err)
			return err
		}
		if unitExist {
			err = sqlOperator.DeleteUnit(unitName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
