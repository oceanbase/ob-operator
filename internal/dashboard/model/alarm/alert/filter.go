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

package alert

import (
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/oceanbase"
)

type AlertFilter struct {
	Serverity alarm.Serverity      `json:"serverity,omitempty"`
	Instance  oceanbase.OBInstance `json:"instance,omitempty"`
	StartTime int64                `json:"startTime,omitempty"`
	EndTime   int64                `json:"endTime,omitempty"`
	Keyword   string               `json:"keyword,omitempty"`
}
