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
	"time"

	"github.com/go-resty/resty/v2"
	alarmconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/alarm/constant"
	"github.com/oceanbase/ob-operator/pkg/errors"
	"gopkg.in/yaml.v2"

	apimodels "github.com/prometheus/alertmanager/api/v2/models"
	amconfig "github.com/prometheus/alertmanager/config"
)

func GetAlertmanagerConfig() (*amconfig.Config, error) {
	statusResp := &apimodels.AlertmanagerStatus{}
	client := resty.New().SetTimeout(time.Duration(alarmconstant.DefaultAlarmQueryTimeout * time.Second))
	resp, err := client.R().SetHeader("content-type", "application/json").SetResult(statusResp).Get(fmt.Sprintf("%s%s", alarmconstant.AlertManagerAddress, alarmconstant.StatusUrl))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrExternal, "Query status from alertmanager")
	} else if resp.StatusCode() != http.StatusOK {
		return nil, errors.Newf(errors.ErrExternal, "Query status from alertmanager got unexpected status: %d", resp.StatusCode())
	}
	content := statusResp.Config.Original
	config := &amconfig.Config{}
	err = yaml.Unmarshal([]byte(*content), config)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInvalid, "Invalid config, parse failed")
	}
	return config, nil
}
