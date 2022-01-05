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

package provider

import (
	"context"

	"github.com/oceanbase/ob-operator/pkg/cable/observer"
	"github.com/oceanbase/ob-operator/pkg/util"
)

func InitForK8s() {
	DirInit()
	// init http server
	Tiny.Init()
	// run http server
	go Tiny.Run()
	observer.Readiness = false
	observer.OBStarted = false
	util.FuncList = append(util.FuncList, StopForK8s)
}

func StopForK8s() {
	Tiny.Stop(context.TODO())
}
