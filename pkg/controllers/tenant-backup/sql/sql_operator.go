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

func (op *SqlOperator) GetBackupDatabaseJobHistory(name string) []model.BackupJob {
	getBackupFullJobHistorySQL := ReplaceAll(GetBackupFullJobHistorySQLTemplate, GetParameterSQLReplacer(name))
	res := make([]model.BackupJob, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.BackupJob{}).Raw(getBackupFullJobHistorySQL).Rows()
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

func (op *SqlOperator) GetBackupDatabaseJob(name string) []model.BackupJob {
	getBackupFullJobSQL := ReplaceAll(GetBackupFullJobSQLTemplate, GetParameterSQLReplacer(name))
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
	}
	return res
}

func (op *SqlOperator) GetBackupIncrementalJobHistory(name string) []model.BackupJob {
	getBackupIncJobHistorySQL := ReplaceAll(GetBackupIncJobHistorySQLTemplate, GetParameterSQLReplacer(name))
	res := make([]model.BackupJob, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.BackupJob{}).Raw(getBackupIncJobHistorySQL).Rows()
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

func (op *SqlOperator) GetBackupIncrementalJob(name string) []model.BackupJob {
	getBackupIncJobSQL := ReplaceAll(GetBackupIncJobSQLTemplate, GetParameterSQLReplacer(name))
	res := make([]model.BackupJob, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.BackupJob{}).Raw(getBackupIncJobSQL).Rows()
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

func (op *SqlOperator) GetBackupDest() []model.TenantBackupDest {
	res := make([]model.TenantBackupDest, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.TenantBackupDest{}).Raw(GetBackupDestSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.TenantBackupDest
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

func (op *SqlOperator) GetDeletePolicy() []model.DeletePolicy {
	res := make([]model.DeletePolicy, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.DeletePolicy{}).Raw(GetDeletePolicySQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.DeletePolicy
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

func (op *SqlOperator) SetBackupPassword(pwd string) error {
	sql := ReplaceAll(SetBackupPasswordTemplate, SetBackupPasswordReplacer(pwd))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) StartBackupDatabase() error {
	return op.ExecSQL(StartBackupDatabaseSql)
}

func (op *SqlOperator) StartBackupIncremental() error {
	return op.ExecSQL(StartBackupIncrementalSql)
}

func (op *SqlOperator) CancelArchiveLog(name string) error {
	sql := ReplaceAll(CancelArchiveLogSQLTemplate, GetParameterSQLReplacer(name))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) CancelBackup() error {
	return op.ExecSQL(CancelBackupSQL)
}

func (op *SqlOperator) DropDeletePolicy(policyName string) error {
	sql := ReplaceAll(DropDeleteBackupSQLTemplate, GetParameterSQLReplacer(policyName))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) SetDeletePolicy(policy model.DeletePolicy) error {
	sql := ReplaceAll(SetDeletePolicySQLTemplate, SetDeletePolicySQLReplacer(policy.PolicyName, policy.RecoveryWindow))
	return op.ExecSQL(sql)
}
