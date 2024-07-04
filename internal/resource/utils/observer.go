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

package utils

import (
	"strconv"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

func CompOBServerDeletionPriority(obs1, obs2 *v1alpha1.OBServer) int {
	priority1 := 0
	priority2 := 0
	val1, exist := GetAnnotationField(obs1, oceanbaseconst.AnnotationsDeletionPriority)
	if exist {
		p1, err := strconv.Atoi(val1)
		if err == nil {
			priority1 = p1
		}
	}
	val2, exist := GetAnnotationField(obs2, oceanbaseconst.AnnotationsDeletionPriority)
	if exist {
		p2, err := strconv.Atoi(val2)
		if err == nil {
			priority2 = p2
		}
	}
	return priority1 - priority2
}

func ReverseCompOBServerDeletionPriority(obs1, obs2 *v1alpha1.OBServer) int {
	return -CompOBServerDeletionPriority(obs1, obs2)
}
