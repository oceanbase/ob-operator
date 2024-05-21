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

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm"
	amconfig "github.com/prometheus/alertmanager/config"
	amlabels "github.com/prometheus/alertmanager/pkg/labels"
	logger "github.com/sirupsen/logrus"
)

type Route struct {
	Receiver              string          `json:"receiver" binding:"required"`
	Matchers              []alarm.Matcher `json:"matchers" binding:"required"`
	AggregateLabels       []string        `json:"aggregateLabels" binding:"required"`
	GroupIntervalMinutes  int             `json:"groupInterval" binding:"required"`
	GroupWaitMinutes      int             `json:"groupWait" binding:"required"`
	RepeatIntervalMinutes int             `json:"repeatInterval" binding:"required"`
}

type RouteIdentity struct {
	Id string `json:"id" binding:"required"`
}

type RouteResponse struct {
	Id string `json:"id" binding:"required"`
	Route
}

func (r *Route) Hash() string {
	routeBytes, err := json.Marshal(r)
	if err != nil {
		logger.WithError(err).Errorf("Encode route object failed")
		return ""
	}
	hash := md5.Sum(routeBytes)
	return hex.EncodeToString(hash[:])
}

func NewRouteResponse(route *Route) *RouteResponse {
	id := route.Hash()
	routeResponse := &RouteResponse{
		Id:    id,
		Route: *route,
	}
	return routeResponse
}

func NewRoute(amroute *amconfig.Route) *Route {
	matchers := make([]alarm.Matcher, 0, len(amroute.Matchers))
	for _, ammatcher := range amroute.Matchers {
		matcher := alarm.Matcher{
			IsRegex: ammatcher.Type == amlabels.MatchRegexp,
			Name:    ammatcher.Name,
			Value:   ammatcher.Value,
		}
		matchers = append(matchers, matcher)
	}
	route := &Route{
		Receiver:              amroute.Receiver,
		Matchers:              matchers,
		AggregateLabels:       amroute.GroupByStr,
		GroupIntervalMinutes:  int(time.Duration(*amroute.GroupInterval).Minutes()),
		GroupWaitMinutes:      int(time.Duration(*amroute.GroupWait).Minutes()),
		RepeatIntervalMinutes: int(time.Duration(*amroute.RepeatInterval).Minutes()),
	}
	return route
}
