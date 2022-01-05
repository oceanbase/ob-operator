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

package converter

import (
	"github.com/pkg/errors"

	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/sql"
)

func IsRSJobSuccess(clusterIP, podIP string) (bool, error) {
	rsJobStatusList := sql.GetRSJobStatus(clusterIP, podIP)
	if len(rsJobStatusList) == 0 {
		return false, errors.New("get rs job status faild")
	}
	lastJob := rsJobStatusList[len(rsJobStatusList)-1]
	// job status is not SUCCESS
	if lastJob.JobStatus != observerconst.RSJobStatusSuccess {
		return false, nil
	}
	return true, nil
}
