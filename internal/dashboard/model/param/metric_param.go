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

package param

import "github.com/oceanbase/ob-operator/internal/dashboard/model/common"

type QueryRange struct {
	StartTimestamp float64 `json:"startTimestamp"`
	EndTimestamp   float64 `json:"endTimestamp"`
	Step           int64   `json:"step"`
}

type MetricQuery struct {
	Metrics     []string        `json:"metrics"`
	Labels      []common.KVPair `json:"labels"`
	GroupLabels []string        `json:"groupLabels"`
	QueryRange  QueryRange      `json:"queryRange"`
}
