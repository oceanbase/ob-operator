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
	"regexp"

	"github.com/oceanbase/ob-operator/pkg/controllers/backup/model"
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
		match, _ := regexp.MatchString("SET ENCRYPTION ON IDENTIFIED BY '(.*)' ONLY", SQL)
		if match {
			klog.Infoln("SET ENCRYPTION ON IDENTIFIED BY '******' ONLY")
		} else {
			klog.Infoln(SQL)
		}
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

func (op *SqlOperator) GetAllBackupSet() []model.AllBackupSet {
	res := make([]model.AllBackupSet, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.AllBackupSet{}).Raw(GetBackupSetSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.AllBackupSet
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

func (op *SqlOperator) SetParameter(name, value string) error {
	sql := ReplaceAll(SetParameterTemplate, SetParameterSQLReplacer(name, value))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) SetBackupPassword(pwd string) error {
	sql := ReplaceAll(SetBackupPasswordTemplate, SetBackupPasswordReplacer(pwd))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) GetArchieveLogStatus() []model.BackupArchiveLogStatus {
	res := make([]model.BackupArchiveLogStatus, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.BackupArchiveLogStatus{}).Raw(GetArchieveLogStatusSql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.BackupArchiveLogStatus
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

func (op *SqlOperator) GetBackupDest() []model.BackupDestValue {
	res := make([]model.BackupDestValue, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.BackupDestValue{}).Raw(GetBackupDestSql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.BackupDestValue
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

func (op *SqlOperator) GetAllTenant() []model.Tenant {
	res := make([]model.Tenant, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.Tenant{}).Raw(GetTenantSQL).Rows()
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

func (op *SqlOperator) GetBackupDatabaseJobHistory(name string) []model.AllBackupSet {
	getBackupFullJobHistorySQL := ReplaceAll(GetBackupFullJobHistorySQLTemplate, TenantIDReplacer(name))
	res := make([]model.AllBackupSet, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.AllBackupSet{}).Raw(getBackupFullJobHistorySQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.AllBackupSet
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

func (op *SqlOperator) GetBackupDatabaseJob(name string) []model.AllBackupSet {
	getBackupFullJobSQL := ReplaceAll(GetBackupFullJobSQLTemplate, TenantIDReplacer(name))
	res := make([]model.AllBackupSet, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.AllBackupSet{}).Raw(getBackupFullJobSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.AllBackupSet
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

func (op *SqlOperator) GetBackupIncrementalJobHistory(name string) []model.AllBackupSet {
	getBackupIncJobHistorySQL := ReplaceAll(GetBackupIncJobHistorySQLTemplate, TenantIDReplacer(name))
	res := make([]model.AllBackupSet, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.AllBackupSet{}).Raw(getBackupIncJobHistorySQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.AllBackupSet
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
func (op *SqlOperator) GetBackupIncrementalJob(name string) []model.AllBackupSet {
	getBackupIncJobSQL := ReplaceAll(GetBackupIncJobSQLTemplate, TenantIDReplacer(name))
	res := make([]model.AllBackupSet, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.AllBackupSet{}).Raw(getBackupIncJobSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.AllBackupSet
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

func (op *SqlOperator) StartArchieveLog() error {
	return op.ExecSQL(StartArchieveLogSql)
}

func (op *SqlOperator) StartBackupDatabase() error {
	return op.ExecSQL(StartBackupDatabaseSql)
}

func (op *SqlOperator) StartBackupIncremental() error {
	return op.ExecSQL(StartBackupIncrementalSql)
}

func (op *SqlOperator) StopArchiveLog() error {
	return op.ExecSQL(StopArchieveLogSql)
}
