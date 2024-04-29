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
package route

import "github.com/oceanbase/ob-operator/internal/dashboard/model/alarm"

type Route struct {
	Receiver              string          `json:"receiver" binding:"required"`
	Matchers              []alarm.Matcher `json:"matchers" binding:"required"`
	AggregateLabels       []string        `json:"aggregateLabels" binding:"required"`
	GroupIntervalMinutes  int             `json:"groupInterval" binding:"required"`
	GroupWaitMinutes      int             `json:"groupWait" binding:"required"`
	RepeatIntervalMinutes int             `json:"repeatInterval" binding:"required"`
}

type RouteResponse struct {
	Id string `json:"id" binding:"required"`
	Route
}
