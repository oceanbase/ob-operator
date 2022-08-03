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

    "k8s.io/klog/v2"
	
	"github.com/oceanbase/ob-operator/pkg/config/constant"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/model"
)

func SetServerOfflineTime(clusterIP string, offlineTime int) error {
	sql := ReplaceAll(SetServerOfflineTimeSQLTemplate, SetServerOfflineTimeSQLReplacer(offlineTime))
	return ExecSQL(clusterIP, constant.OBSERVER_MYSQL_PORT, DatabaseOb, sql, 5)
}

func CreateUser(clusterIP, user, password string) error {
	sql := ReplaceAll(CreateUserSQLTemplate, CreateUserSQLReplacer(user, password))
	return ExecSQL(clusterIP, constant.OBSERVER_MYSQL_PORT, DatabaseOb, sql, 5)
}

func GrantPrivilege(clusterIP, privilege, object, user string) error {
	sql := ReplaceAll(GrantPrivilegeSQLTemplate, GrantPrivilegeSQLReplacer(privilege, object, user))
	return ExecSQL(clusterIP, constant.OBSERVER_MYSQL_PORT, DatabaseOb, sql, 5)
}

func GetParameter(clusterIP, name string) []model.SysParameterStat {
	sql := ReplaceAll(GetParameterTemplate, GetParameterSQLReplacer(name))
	return GetSysParameterFromDB(clusterIP, constant.OBSERVER_MYSQL_PORT, DatabaseOb, sql)
}

// TODO: maybe need to set with scope
func SetParameter(clusterIP, name, value string) error {
	sql := ReplaceAll(SetParameterTemplate, SetParameterSQLReplacer(name, value))
	return ExecSQL(clusterIP, constant.OBSERVER_MYSQL_PORT, DatabaseOb, sql, 5)
}

func BootstrapForOB(IP, SQL string) {
	setTimeOutRes := ExecSQL(IP, constant.OBSERVER_MYSQL_PORT, "", SetTimeoutSQL, 5)
	if setTimeOutRes != nil {
		klog.Errorln("set ob_query_timeout error", setTimeOutRes)
	}
	bootstrapRes := ExecSQL(IP, constant.OBSERVER_MYSQL_PORT, "", SQL, 300)
	if bootstrapRes != nil {
		klog.Errorln("run bootstrap sql error", bootstrapRes)
	}
}

func GetOBServer(IP string) []model.AllServer {
	return GetOBServerFromDB(IP, constant.OBSERVER_MYSQL_PORT, DatabaseOb, GetOBServerSQL)
}

func GetOBZone(IP string) []model.AllZone {
	return GetOBZoneFromDB(IP, constant.OBSERVER_MYSQL_PORT, DatabaseOb, GetOBZoneSQL)
}

func AddServer(clusterIP, zoneName, podIP string) error {
	serverIP := fmt.Sprintf("%s:%d", podIP, constant.OBSERVER_RPC_PORT)
	sql := ReplaceAll(AddServerSQLTemplate, AddServerSQLReplacer(zoneName, serverIP))
	return ExecSQL(clusterIP, constant.OBSERVER_MYSQL_PORT, DatabaseOb, sql, 60)
}

func DelServer(clusterIP, podIP string) error {
	serverIP := fmt.Sprintf("%s:%d", podIP, constant.OBSERVER_RPC_PORT)
	sql := ReplaceAll(DelServerSQLTemplate, DelServerSQLReplacer(serverIP))
	return ExecSQL(clusterIP, constant.OBSERVER_MYSQL_PORT, DatabaseOb, sql, 60)
}

func GetRootService(IP string) []model.AllVirtualCoreMeta {
	return GetRootServiceFromDB(IP, constant.OBSERVER_MYSQL_PORT, DatabaseOb, GetRootServiceSQL)
}

func GetRSJobStatus(clusterIP, podIP string) []model.RSJobStatus {
	sql := ReplaceAll(GetRSJobStatusSQL, GetRSJobStatusSQLReplacer(podIP, constant.OBSERVER_RPC_PORT))
	return GetRSJobStatusFromDB(clusterIP, constant.OBSERVER_MYSQL_PORT, DatabaseOb, sql)
}

func AddZone(clusterIP, zoneName string) error {
	sql := ReplaceAll(AddZoneSQLTemplate, AddZoneSQLReplacer(zoneName))
	return ExecSQL(clusterIP, constant.OBSERVER_MYSQL_PORT, DatabaseOb, sql, 60)
}

func StartZone(clusterIP, zoneName string) error {
	sql := ReplaceAll(StartZoneSQLTemplate, StartZoneSQLReplacer(zoneName))
	return ExecSQL(clusterIP, constant.OBSERVER_MYSQL_PORT, DatabaseOb, sql, 60)
}

func GetAllUnit(clusterIP string) []model.AllUnit {
	return GetAllUnitFromDB(clusterIP, constant.OBSERVER_MYSQL_PORT, DatabaseOb, GetAllUnitSql)
}

func StopZone(clusterIP, zoneName string) error {
	sql := ReplaceAll(StopOBZoneTemplate, StopZoneSQLReplacer(zoneName))
	return ExecSQL(clusterIP, constant.OBSERVER_MYSQL_PORT, DatabaseOb, sql, 60)
}

func DeleteZone(clusterIP, zoneName string) error {
	sql := ReplaceAll(DeleteOBZoneTemplate, DeleteZoneSQLReplacer(zoneName))
	return ExecSQL(clusterIP, constant.OBSERVER_MYSQL_PORT, DatabaseOb, sql, 60)
}
