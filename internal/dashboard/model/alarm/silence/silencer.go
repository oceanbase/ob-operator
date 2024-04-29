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
	"github.com/oceanbase/ob-operator/internal/dashboard/model/oceanbase"
)

type SilencerResponse struct {
	Id        string                 `json:"id" binding:"required"`
	Instances []oceanbase.OBInstance `json:"instance" binding:"required"`
	Status    *Status                `json:"status" binding:"required"`
	UpdatedAt int64                  `json:"updatedAt" binding:"required"`
	Silencer
}

type SilencerParam struct {
	Id        string                 `json:"id,omitempty"`
	Instances []oceanbase.OBInstance `json:"instance" binding:"required"`
	Silencer
}

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
