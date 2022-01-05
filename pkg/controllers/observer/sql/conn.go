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

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"k8s.io/klog/v2"
)

func ConnOB(IP, port, dbName string, timeout int) *gorm.DB {
	connInfo := fmt.Sprintf("root:@tcp(%s:%s)/%s?timeout=%ds&charset=utf8&parseTime=True&loc=Local", IP, port, dbName, timeout)
	client, err := gorm.Open("mysql", connInfo)
	if err != nil {
		errNum, errMsg := covertErrToMySQLError(err)
		klog.Errorln(errNum, errMsg)
		return nil
	}
	return client
}
