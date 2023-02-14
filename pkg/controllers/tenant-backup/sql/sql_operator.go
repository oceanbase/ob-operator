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
	"github.com/oceanbase/ob-operator/pkg/controllers/tenant-backup/model"

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

func (op *SqlOperator) ShowParameter(name string) []model.SysParameterStat {
	res := make([]model.SysParameterStat, 0)
	sql := ReplaceAll(ShowParameterTemplate, GetParameterSQLReplacer(name))
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.SysParameterStat{}).Raw(sql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.SysParameterStat
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

func (op *SqlOperator) GetArchiveLogDest() []model.TenantArchiveDest {
	res := make([]model.TenantArchiveDest, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.TenantArchiveDest{}).Raw(GetArchiveLogDestSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.TenantArchiveDest
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

func (op *SqlOperator) GetArchiveLog() []model.TenantArchiveLog {
	res := make([]model.TenantArchiveLog, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.TenantArchiveLog{}).Raw(GetArchiveLogSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.TenantArchiveLog
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

func (op *SqlOperator) GetBackupJob(name string) []model.BackupJob {
	getBackupFullJobSQL := ReplaceAll(GetBackupFullJobSQLTemplate, GetParameterSQLReplacer(name))
	getBackupInncJobSQL := ReplaceAll(GetBackupIncJobSQLTemplate, GetParameterSQLReplacer(name))
	res := make([]model.BackupJob, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.BackupJob{}).Raw(getBackupFullJobSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.BackupJob
			for rows.Next() {
				err = client.ScanRows(rows, &rowData)
				if err == nil {
					res = append(res, rowData)
				}
			}
		}
		rows, err = client.Model(&model.BackupJob{}).Raw(getBackupInncJobSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.BackupJob
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

func (op *SqlOperator) StartAchiveLog() error {
	return op.ExecSQL(StartArchiveLogSQL)
}

func (op *SqlOperator) SetParameter(name, value string) error {
	sql := ReplaceAll(SetParameterTemplate, SetParameterSQLReplacer(name, value))
	return op.ExecSQL(sql)
}
