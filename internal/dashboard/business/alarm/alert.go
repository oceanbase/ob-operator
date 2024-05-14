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
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	alarmconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/alarm/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/alert"
	"github.com/oceanbase/ob-operator/pkg/errors"

	apimodels "github.com/prometheus/alertmanager/api/v2/models"
	logger "github.com/sirupsen/logrus"
)

func ListAlerts(filter *alert.AlertFilter) ([]alert.Alert, error) {
	client := resty.New().SetTimeout(time.Duration(alarmconstant.DefaultAlarmQueryTimeout * time.Second))
	gettableAlerts := make(apimodels.GettableAlerts, 0)
	resp, err := client.R().SetQueryParams(map[string]string{
		"active":      "true",
		"silenced":    "true",
		"inhibited":   "true",
		"unprocessed": "true",
		"receiver":    "",
	}).SetHeader("content-type", "application/json").SetResult(&gettableAlerts).Get(fmt.Sprintf("%s%s", alarmconstant.AlertManagerAddress, alarmconstant.AlertUrl))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrExternal, "Query alerts from alertmanager")
	} else if resp.StatusCode() != http.StatusOK {
		return nil, errors.Newf(errors.ErrExternal, "Query alerts from alertmanager got unexpected status: %d", resp.StatusCode())
	}
	filteredAlerts := make([]alert.Alert, 0)
	for _, gettableAlert := range gettableAlerts {
		alert, err := alert.NewAlert(gettableAlert)
		if err != nil {
			logger.WithError(err).Error("Parse alert got error, just skip")
		}
		if filterAlert(alert, filter) {
			filteredAlerts = append(filteredAlerts, *alert)
		}
	}
	return filteredAlerts, nil
}

func filterAlert(alert *alert.Alert, filter *alert.AlertFilter) bool {
	matched := true
	if filter.Serverity != "" {
		matched = matched && (filter.Serverity == alert.Serverity)
	}
	if filter.StartTime != 0 {
		matched = matched && (filter.StartTime <= alert.StartsAt)
	}
	if filter.EndTime != 0 {
		matched = matched && (filter.EndTime >= alert.StartsAt)
	}
	if filter.Keyword != "" {
		matched = matched && strings.Contains(alert.Description, filter.Keyword)
	}
	if filter.Instance != nil {
		matched = matched && filter.Instance.Equals(alert.Instance)
	}
	return matched
}
