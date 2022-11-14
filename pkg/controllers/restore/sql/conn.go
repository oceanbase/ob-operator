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

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"k8s.io/klog/v2"
)

type DBConnectProperties struct {
	IP       string
	Port     int
	User     string
	Password string
	Database string
	Timeout  int
}

func GetDBClient(p *DBConnectProperties) (*gorm.DB, error) {
	connInfo := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=%ds&charset=utf8&parseTime=True&loc=Local", p.User, p.Password, p.IP, p.Port, p.Database, p.Timeout)
	client, err := gorm.Open("mysql", connInfo)
	if err != nil {
		errNum, errMsg := covertErrToMySQLError(err)
		klog.Errorln(errNum, errMsg)
		return nil, err
	}
	return client, err
}

func covertErrToMySQLError(err error) (uint16, string) {
	mysqlErr, ok := err.(*mysql.MySQLError)
	if ok {
		return mysqlErr.Number, mysqlErr.Message
	}
	klog.Errorln(err)
	return 0, ""
}
