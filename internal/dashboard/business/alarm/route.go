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

package alarm

import (
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/route"
	"github.com/oceanbase/ob-operator/pkg/errors"
)

func GetRoute(id string) (*route.RouteResponse, error) {
	routes, err := ListRoutes()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrExternal, "Failed to get routes")
	}
	for _, r := range routes {
		if r.Id == id {
			return &r, nil
		}
	}
	return nil, errors.NewNotFound("Route not found")
}

func ListRoutes() ([]route.RouteResponse, error) {
	config, err := GetAlertmanagerConfig()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrExternal, "Failed to get config")
	}

	routes := make([]route.RouteResponse, 0, len(config.Route.Routes))
	for _, amroute := range config.Route.Routes {
		r := route.NewRoute(amroute)
		if r != nil {
			routes = append(routes, *route.NewRouteResponse(r))
		}
	}
	return routes, nil
}
