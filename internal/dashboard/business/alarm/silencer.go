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
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	alarmconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/alarm/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/silence"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/errors"
	apimodels "github.com/prometheus/alertmanager/api/v2/models"
	logger "github.com/sirupsen/logrus"
)

func ListSilencers(filter *silence.SilencerFilter) ([]silence.SilencerResponse, error) {
	client := resty.New().SetTimeout(time.Duration(alarmconstant.DefaultAlarmQueryTimeout * time.Second))
	gettableSilencers := make(apimodels.GettableSilences, 0)
	queryFilter := make([]string, 0)
	if filter.Instance != nil {
		queryFilter = append(queryFilter, fmt.Sprintf("%s=\"%s\"", alarmconstant.LabelOBCluster, filter.Instance.OBCluster))
		switch filter.Instance.Type {
		case oceanbase.OBCluster:
			// already added
		case oceanbase.OBZone:
			queryFilter = append(queryFilter, fmt.Sprintf("%s=\"%s\"", alarmconstant.LabelOBZone, filter.Instance.OBZone))
		case oceanbase.OBServer:
			queryFilter = append(queryFilter, fmt.Sprintf("%s=\"%s\"", alarmconstant.LabelOBServer, filter.Instance.OBServer))
		case oceanbase.OBTenant:
			queryFilter = append(queryFilter, fmt.Sprintf("%s=\"%s\"", alarmconstant.LabelOBTenant, filter.Instance.OBTenant))
		default:
			return nil, errors.NewBadRequest("Unknown instance type")
		}
	}
	req := client.R().SetHeader("content-type", "application/json")
	if len(queryFilter) > 0 {
		req = req.SetQueryParamsFromValues(url.Values{
			"filter": queryFilter,
		})
	}
	resp, err := req.SetResult(&gettableSilencers).Get(fmt.Sprintf("%s%s", alarmconstant.AlertManagerAddress, alarmconstant.SilencerQueryUrl))
	if err != nil || resp.StatusCode() != http.StatusOK {
		return nil, errors.Wrap(err, errors.ErrExternal, "Query silencers from alertmanager")
	}
	logger.Infof("resp: %v", resp)
	logger.Infof("silencers: %v", gettableSilencers)
	filteredSilencers := make([]silence.SilencerResponse, 0)
	for _, gettableSilencer := range gettableSilencers {
		silencer := silence.NewSilencerResponse(gettableSilencer)
		if filterSilencer(silencer, filter) {
			filteredSilencers = append(filteredSilencers, *silencer)
		}
	}
	return filteredSilencers, nil
}

func filterSilencer(silencer *silence.SilencerResponse, filter *silence.SilencerFilter) bool {
	matched := true
	if filter.Keyword != "" {
		matched = matched && strings.Contains(silencer.Comment, filter.Keyword)
	}
	// require at least one instance matches
	// TODO: whether to consider a cluster in filter matches a tenant or observer if the cluster names are same
	if filter.Instance != nil {
		instanceMatched := false
		for _, instance := range silencer.Instances {
			if instance.Equals(filter.Instance) {
				instanceMatched = true
				break
			}
		}
		matched = matched && instanceMatched
	}
	return matched
}
