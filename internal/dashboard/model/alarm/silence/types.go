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

package silence

import (
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm"
)

type State string

const (
	StateActive State = "active"
)

type Status struct {
	State *State `json:"state" binding:"required"`
}

type Silencer struct {
	Comment   string          `json:"comment" binding:"required"`
	CreatedBy string          `json:"createdBy" binding:"required"`
	StartsAt  int64           `json:"startsAt" binding:"required"`
	EndsAt    int64           `json:"endsAt" binding:"required"`
	Matchers  []alarm.Matcher `json:"matchers" binding:"required"`
}
