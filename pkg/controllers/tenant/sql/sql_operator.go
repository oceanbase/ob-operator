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

package sql

import (
	v1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"github.com/oceanbase/ob-operator/pkg/controllers/tenant/model"
	"github.com/pkg/errors"

	"k8s.io/klog"
)

type SqlOperator struct {
	ConnectProperties *DBConnectProperties
}

func NewSqlOperator(c *DBConnectProperties) *SqlOperator {
	return &SqlOperator{
		ConnectProperties: c,
	}
}

func (op *SqlOperator) TestOK() bool {
	err := op.ExecSQL("select 1")
	return err == nil
}

func (op *SqlOperator) ExecSQL(SQL string) error {
	if SQL != "select 1" {
		klog.Infoln(SQL)
	}
	client, err := GetDBClient(op.ConnectProperties)
	if err != nil {
		return errors.Wrap(err, "Get DB Connection")
	} else {
		defer client.Close()
		res := client.Exec(SQL)
		if res.Error != nil {
			errNum, errMsg := covertErrToMySQLError(res.Error)
			klog.Errorln(errNum, errMsg)
			return errors.New(errMsg)
		}
	}
	return nil
}

func (op *SqlOperator) GetTenantByName(name string) []model.Tenant {
	sql := ReplaceAll(GetTenantSQL, SetNameReplacer(name))
	res := make([]model.Tenant, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.Tenant{}).Raw(sql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.Tenant
			for rows.Next() {
				err = client.ScanRows(rows, &rowData)
				if err == nil {
					res = append(res, rowData)
				}
			}
		}
	}
	return res
}

func (op *SqlOperator) GetPoolByName(name string) []model.Pool {
	sql := ReplaceAll(GetPoolSQL, SetNameReplacer(name))
	res := make([]model.Pool, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.Pool{}).Raw(sql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.Pool
			for rows.Next() {
				err = client.ScanRows(rows, &rowData)
				if err == nil {
					res = append(res, rowData)
				}
			}
		}
	}
	return res
}

func (op *SqlOperator) GetPoolList() []model.Pool {
	res := make([]model.Pool, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.Pool{}).Raw(GetPoolListSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.Pool
			for rows.Next() {
				err = client.ScanRows(rows, &rowData)
				if err == nil {
					res = append(res, rowData)
				}
			}
		}
	}
	return res
}

func (op *SqlOperator) GetUnitList() []model.Unit {
	res := make([]model.Unit, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.Unit{}).Raw(GetUnitListSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.Unit
			for rows.Next() {
				err = client.ScanRows(rows, &rowData)
				if err == nil {
					res = append(res, rowData)
				}
			}
		}
	}
	return res
}

func (op *SqlOperator) GetUnitConfigV3List() []model.UnitConfigV3 {
	res := make([]model.UnitConfigV3, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.UnitConfigV3{}).Raw(GetUnitConfigV3ListSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.UnitConfigV3
			for rows.Next() {
				err = client.ScanRows(rows, &rowData)
				if err == nil {
					res = append(res, rowData)
				}
			}
		}
	}
	return res
}

func (op *SqlOperator) GetUnitConfigV4List() []model.UnitConfigV4 {
	res := make([]model.UnitConfigV4, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.UnitConfigV4{}).Raw(GetUnitConfigV4ListSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.UnitConfigV4
			for rows.Next() {
				err = client.ScanRows(rows, &rowData)
				if err == nil {
					res = append(res, rowData)
				}
			}
		}
	}
	return res
}

func (op *SqlOperator) GetUnitConfigByName(name string) []model.UnitConfig {
	sql := ReplaceAll(GetUnitConfigSQL, SetNameReplacer(name))
	res := make([]model.UnitConfig, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.UnitConfig{}).Raw(sql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.UnitConfig
			for rows.Next() {
				err = client.ScanRows(rows, &rowData)
				if err == nil {
					res = append(res, rowData)
				}
			}
		}
	}
	return res
}

func (op *SqlOperator) GetResource(zone v1.TenantReplica) []model.Resource {
	sql := ReplaceAll(GetResourceSQLTemplate, GetResourceSQLReplacer(zone.ZoneName))
	res := make([]model.Resource, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.Resource{}).Raw(sql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.Resource
			for rows.Next() {
				err = client.ScanRows(rows, &rowData)
				if err == nil {
					res = append(res, rowData)
				}
			}
		}
	}
	return res
}

func (op *SqlOperator) GetCharset() []model.Charset {
	res := make([]model.Charset, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.Charset{}).Raw(GetCharsetSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.Charset
			for rows.Next() {
				err = client.ScanRows(rows, &rowData)
				if err == nil {
					res = append(res, rowData)
				}
			}
		}
	}
	return res
}

func (op *SqlOperator) GetVariable(name string) []model.Variable {
	res := make([]model.Variable, 0)
	sql := ReplaceAll(GetVariableSQLTemplate, GetVariableSQLReplacer(name))
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.Variable{}).Raw(sql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.Variable
			for rows.Next() {
				err = client.ScanRows(rows, &rowData)
				if err == nil {
					res = append(res, rowData)
				}
			}
		}
	}
	return res
}

func (op *SqlOperator) GetInprogressJob(name string) []model.RsJob {
	sql := ReplaceAll(GetInprogressJobSQLTemplate, SetNameReplacer(name))
	res := make([]model.RsJob, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.RsJob{}).Raw(sql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.RsJob
			for rows.Next() {
				err = client.ScanRows(rows, &rowData)
				if err == nil {
					res = append(res, rowData)
				}
			}
		}
	}
	return res
}

func (op *SqlOperator) GetVersion() []model.OBVersion {
	res := make([]model.OBVersion, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.OBVersion{}).Raw(GetObVersionSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.OBVersion
			for rows.Next() {
				err = client.ScanRows(rows, &rowData)
				if err == nil {
					res = append(res, rowData)
				}
			}
		}
	}
	return res
}

func (op *SqlOperator) CreateUnitV3(name string, resourceUnit v1.ResourceUnit) error {
	sql := ReplaceAll(CreateUnitV3SQLTemplate, CreateUnitV3SQLReplacer(name, resourceUnit))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) CreateUnitV4(name string, resourceUnit v1.ResourceUnit, option string) error {
	sql := ReplaceAll(CreateUnitV4SQLTemplate, CreateUnitV4SQLReplacer(name, resourceUnit, option))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) CreatePool(poolName, unitName string, zone v1.TenantReplica) error {
	sql := ReplaceAll(CreatePoolSQLTemplate, CreatePoolSQLReplacer(poolName, unitName, zone))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) CreateTenant(tenantName, charset, zoneList, primaryZone, poolList, locality, collate, variableList string) error {
	sql := ReplaceAll(CreateTenantSQLTemplate, CreateTenantSQLReplacer(tenantName, charset, zoneList, primaryZone, poolList, locality, collate, variableList))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) SetTenantVariable(tenantName, variableList string) error {
	sql := ReplaceAll(SetTenantVariableSQLTemplate, SetTenantVariableSQLReplacer(tenantName, variableList))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) SetUnitConfigV3(name string, resourceUnit model.ResourceUnitV3) error {
	sql := ReplaceAll(SetUnitConfigV3SQLTemplate, SetUnitConfigV3SQLReplacer(name, resourceUnit))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) SetUnitConfigV4(name string, resourceUnit model.ResourceUnitV4, option string) error {
	sql := ReplaceAll(SetUnitConfigV4SQLTemplate, SetUnitConfigV4SQLReplacer(name, resourceUnit, option))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) SetPoolUnitNum(name string, unitNum int) error {
	sql := ReplaceAll(SetPoolUnitNumSQLTemplate, SetPoolUnitNumSQLReplacer(name, unitNum))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) SetTenantLocality(name, locality string) error {
	sql := ReplaceAll(SetTenantLocalitySQLTemplate, SetTenantLocalitySQLReplacer(name, locality))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) SetTenant(name, zoneList, primaryZone, poolList, charset, locality string) error {
	sql := ReplaceAll(SetTenantSQLTemplate, SetTenantSQLReplacer(name, zoneList, primaryZone, poolList, charset, locality))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) DeleteUnit(name string) error {
	sql := ReplaceAll(DeleteUnitSQLTemplate, SetNameReplacer(name))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) DeletePool(name string) error {
	sql := ReplaceAll(DeletePoolSQLTemplate, SetNameReplacer(name))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) DeleteTenant(name string) error {
	sql := ReplaceAll(DeleteTenantSQLTemplate, SetNameReplacer(name))
	return op.ExecSQL(sql)
}
