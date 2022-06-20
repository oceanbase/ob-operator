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
	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	"github.com/oceanbase/ob-operator/pkg/controllers/observer/model"
)

func ExecSQL(ip string, port int, db string, SQL string, timeout int) error {
	klog.Infoln(SQL)
	client := ConnOB(ip, port, db, timeout)
	if client != nil {
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

func GetOBServerFromDB(ip string, port int, db string, SQL string) []model.AllServer {
	client := ConnOB(ip, port, db, 5)
	res := make([]model.AllServer, 0)
	if client != nil {
		defer client.Close()
		rows, err := client.Model(&model.AllServer{}).Raw(SQL).Rows()
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

func GetRootServiceFromDB(ip string, port int, db string, SQL string) []model.AllVirtualCoreMeta {
	client := ConnOB(ip, port, db, 5)
	res := make([]model.AllVirtualCoreMeta, 0)
	if client != nil {
		defer client.Close()
		rows, err := client.Model(&model.AllVirtualCoreMeta{}).Raw(SQL).Rows()
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

func GetRSJobStatusFromDB(ip string, port int, db string, SQL string) []model.RSJobStatus {
	client := ConnOB(ip, port, db, 5)
	res := make([]model.RSJobStatus, 0)
	if client != nil {
		defer client.Close()
		rows, err := client.Model(&model.RSJobStatus{}).Raw(SQL).Rows()
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
