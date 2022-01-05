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
	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	statefulappconst "github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/core/converter"
)

func SubsetScaleJundge(subsetsSpec []cloudv1.Subset, subsetsCurrentNameList []string) (string, string) {
	var scaleState string
	var subsetName string

	// new cr
	if len(subsetsCurrentNameList) == 0 {
		return statefulappconst.ScaleUP, subsetsSpec[0].Name
	}

	if len(subsetsSpec) > len(subsetsCurrentNameList) {
		scaleState = statefulappconst.ScaleUP
		// find which subsets need to add
		subsetName = converter.FindElementNotInSubsetsCurrentNameList(subsetsSpec, subsetsCurrentNameList)
	} else if len(subsetsSpec) < len(subsetsCurrentNameList) {
		scaleState = statefulappconst.ScaleDown
		// find which subsets need to delete
		subsetName = converter.FindElementNotInSubsetsSpec(subsetsSpec, subsetsCurrentNameList)
	} else {
		scaleState = statefulappconst.Maintain
	}

	return scaleState, subsetName
}
