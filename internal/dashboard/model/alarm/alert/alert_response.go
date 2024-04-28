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
	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/oceanbase"
)

type Event struct {
	Fingerprint string               `json:"fingerprint" binding:"required"`
	Rule        string               `json:"rule" binding:"required"`
	Serverity   alarm.Serverity      `json:"serverity" binding:"required"`
	Instance    oceanbase.OBInstance `json:"instance" binding:"required"`
	StartsAt    int64                `json:"startsAt" binding:"required"`
	UpdatedAt   int64                `json:"updatedAt" binding:"required"`
	EndsAt      int64                `json:"endsAt" binding:"required"`
	Status      Status               `json:"status" binding:"required"`
	Labels      []common.KVPair      `json:"labels,omitempty"`
	Summary     string               `json:"summary,omitempty"`
	Description string               `json:"description,omitempty"`
}
