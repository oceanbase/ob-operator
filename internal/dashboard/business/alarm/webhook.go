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
	"encoding/json"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/payload"

	logger "github.com/sirupsen/logrus"
)

func LogPayload(pl *payload.WebhookPayload) error {
	for _, alert := range pl.Alerts {
		alertContent, err := json.Marshal(alert)
		if err != nil {
			logger.WithError(err).Error("Encode alarm info failed")
		}
		logger.Infof("Received alert: %s", alertContent)
	}
	return nil
}
