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

package judge

import (
	statefulappconst "github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/const"
)

func PodScale(expectedReplicas, availableReplicas int) string {
	var scaleState string
	if expectedReplicas > availableReplicas {
		scaleState = statefulappconst.ScaleUP
	} else if expectedReplicas < availableReplicas {
		scaleState = statefulappconst.ScaleDown
	} else {
		scaleState = statefulappconst.Maintain
	}
	return scaleState
}
