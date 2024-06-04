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

package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/alarm"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/payload"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
)

// @ID LogAlerts
// @Tags Webhook
// @Summary Log alerts
// @Description Log alerts sent by alertmanager.
// @Accept application/json
// @Produce application/json
// @Param body body payload.WebhookPayload true "payload"
// @Success 200 object response.APIResponse{data=response.DashboardInfo}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/webhook/alert/log [POST]
func LogPayload(ctx *gin.Context) (any, error) {
	payload := &payload.WebhookPayload{}
	err := ctx.Bind(payload)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	return nil, alarm.LogPayload(payload)
}
