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

package core

import (
	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
)

func (r *StatefulAppReconciler) SubsetsCoordinator(statefulApp cloudv1.StatefulApp, subsetsSpec []cloudv1.Subset) error {
	var err error
	var status bool
	subsetCtrl := NewSubsetCtrl(r.CRClient, r.Recorder, statefulApp)
	for _, subset := range subsetsSpec {
		status, err = subsetCtrl.SubsetCoordinator(subset)
		// compare status
		// if true continue, don't need to update, but it is not ok
		// if false beark, do one thing at a time
		if !status || err != nil {
			break
		}
	}
	return err
}
