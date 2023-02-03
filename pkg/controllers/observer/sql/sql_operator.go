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
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	"github.com/oceanbase/ob-operator/pkg/config/constant"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/model"
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
	client, err := GetDBClient(op.ConnectProperties)
	if err != nil {
		return errors.Wrap(err, "Get DB Connection")
	} else {
		defer client.Close()
		res := client.Exec(SQL)
		if res.Error != nil {
			errNum, errMsg := covertErrToMySQLError(res.Error)
			klog.Errorln(fmt.Sprintf("execute sql: %s failed", SQL))
			klog.Errorln(errNum, errMsg)
			return errors.New(errMsg)
		}
	}
	return nil
}

func (op *SqlOperator) SetServerOfflineTime(offlineTime int) error {
	sql := ReplaceAll(SetServerOfflineTimeSQLTemplate, SetServerOfflineTimeSQLReplacer(offlineTime))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) CreateUser(user, password string) error {
	sql := ReplaceAll(CreateUserSQLTemplate, CreateUserSQLReplacer(user, password))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) GrantPrivilege(privilege, object, user string) error {
	sql := ReplaceAll(GrantPrivilegeSQLTemplate, GrantPrivilegeSQLReplacer(privilege, object, user))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) SetParameter(name, value string) error {
	sql := ReplaceAll(SetParameterTemplate, SetParameterSQLReplacer(name, value))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) BootstrapForOB(SQL string) error {
	// TODO: set timeout with variables
	setTimeOutRes := op.ExecSQL(SetTimeoutSQL)
	if setTimeOutRes != nil {
		klog.Errorln("set ob_query_timeout error", setTimeOutRes)
	}
	bootstrapRes := op.ExecSQL(SQL)
	if bootstrapRes != nil {
		return errors.New(fmt.Sprintf("run bootstrap sql got error %v", bootstrapRes))
	}
	return nil
}

func (op *SqlOperator) AddServer(zoneName, podIP string) error {
	serverIP := fmt.Sprintf("%s:%d", podIP, constant.OBSERVER_RPC_PORT)
	sql := ReplaceAll(AddServerSQLTemplate, AddServerSQLReplacer(zoneName, serverIP))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) DelServer(podIP string) error {
	serverIP := fmt.Sprintf("%s:%d", podIP, constant.OBSERVER_RPC_PORT)
	sql := ReplaceAll(DelServerSQLTemplate, DelServerSQLReplacer(serverIP))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) AddZone(zoneName string) error {
	sql := ReplaceAll(AddZoneSQLTemplate, ZoneNameReplacer(zoneName))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) StartZone(zoneName string) error {
	sql := ReplaceAll(StartZoneSQLTemplate, ZoneNameReplacer(zoneName))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) StopZone(zoneName string) error {
	sql := ReplaceAll(StopOBZoneTemplate, ZoneNameReplacer(zoneName))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) DeleteZone(zoneName string) error {
	sql := ReplaceAll(DeleteOBZoneTemplate, ZoneNameReplacer(zoneName))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) BeginUpgrade() error {
	return op.ExecSQL(BeginUpgradeSQL)
}

func (op *SqlOperator) EndUpgrade() error {
	return op.ExecSQL(EndUpgradeSQL)
}

func (op *SqlOperator) UpgradeSchema() error {
	return op.ExecSQL(UpgradeSchemaSQL)
}

func (op *SqlOperator) RunRootInspection() error {
	return op.ExecSQL(RunRootInspectionJobSQL)
}

func (op *SqlOperator) MajorFreeze() error {
	return op.ExecSQL(MajorFreezeSQL)
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

func (op *SqlOperator) GetLeaderCount() []model.ZoneLeaderCount {
	res := make([]model.ZoneLeaderCount, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.ZoneLeaderCount{}).Raw(GetLeaderCountSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.ZoneLeaderCount
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

func (op *SqlOperator) GetParameter(name string) []model.SysParameterStat {
	res := make([]model.SysParameterStat, 0)
	sql := ReplaceAll(GetParameterTemplate, GetParameterSQLReplacer(name))
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

func (op *SqlOperator) GetOBServer() []model.AllServer {
	res := make([]model.AllServer, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.AllServer{}).Raw(GetOBServerSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.AllServer
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

func (op *SqlOperator) GetClogStat() []model.ClogStat {
	res := make([]model.ClogStat, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.ClogStat{}).Raw(GetClogStatSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.ClogStat
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

func (op *SqlOperator) GetOBZone() []model.AllZone {
	res := make([]model.AllZone, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.AllZone{}).Raw(GetOBZoneSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.AllZone
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

func (op *SqlOperator) GetFrozenVersion() []model.AllZone {
	res := make([]model.AllZone, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.AllZone{}).Raw(GetFrozeVersionSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.AllZone
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

func (op *SqlOperator) GetLastMergedVersion() []model.AllZone {
	res := make([]model.AllZone, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.AllZone{}).Raw(GetLastMergedVersionSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.AllZone
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

func (op *SqlOperator) GetRootService() []model.AllVirtualCoreMeta {
	res := make([]model.AllVirtualCoreMeta, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.AllVirtualCoreMeta{}).Raw(GetRootServiceSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.AllVirtualCoreMeta
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

func (op *SqlOperator) GetRSJobStatus(podIP string) []model.RSJobStatus {
	res := make([]model.RSJobStatus, 0)
	sql := ReplaceAll(GetRSJobStatusSQL, GetRSJobStatusSQLReplacer(podIP, constant.OBSERVER_RPC_PORT))
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.RSJobStatus{}).Raw(sql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.RSJobStatus
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

func (op *SqlOperator) GetAllUnit() []model.AllUnit {
	res := make([]model.AllUnit, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.AllUnit{}).Raw(GetAllUnitSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.AllUnit
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
