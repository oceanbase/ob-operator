/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package main

import (
	"github.com/oceanbase/ob-operator/pkg/oceanbase/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	"time"
)

func main() {
	p := connector.NewOceanbaseConnectProperties("10.42.0.220", 2881, "root", "sys", "root", "oceanbase")
	manager, err := operation.GetOceanbaseOperationManager(p)
	if err != nil {
		panic(err)
	}
	serverInfo := &model.ServerInfo{
		Ip:   "10.42.0.66",
		Port: 2882,
	}
	manager.DeleteServer(serverInfo)
	time.Sleep(time.Second * 60)
	manager.AddServer(serverInfo)
}
