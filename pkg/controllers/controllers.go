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

package controllers

import (
	"reflect"
	"runtime"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/oceanbase/ob-operator/pkg/controllers/backup"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer"
	"github.com/oceanbase/ob-operator/pkg/controllers/restore"
	"github.com/oceanbase/ob-operator/pkg/controllers/statefulapp"
)

var controllerAddFuncs []func(manager.Manager) error

func init() {
	controllerAddFuncs = append(controllerAddFuncs, statefulapp.Add)
	controllerAddFuncs = append(controllerAddFuncs, observer.Add)
	controllerAddFuncs = append(controllerAddFuncs, backup.Add)
	controllerAddFuncs = append(controllerAddFuncs, restore.Add)
}

// SetupWithManager load controller
func SetupWithManager(m manager.Manager) error {
	for _, f := range controllerAddFuncs {
		klog.Infoln("load", runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name())
		err := f(m)
		if err != nil {
			kindMatchErr, ok := err.(*meta.NoKindMatchError)
			if ok {
				klog.Errorf("CRD %v is not installed.", kindMatchErr.GroupKind)
				continue
			}
			return err
		}
	}
	return nil
}
