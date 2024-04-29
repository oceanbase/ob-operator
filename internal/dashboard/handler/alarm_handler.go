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
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/alert"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/receiver"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/route"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/rule"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/silence"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
)

// @ID ListAlerts
// @Tags Alarm
// @Summary List alerts
// @Description List alerts, filter with alarm objects, serverity, time and keywords.
// @Accept application/json
// @Produce application/json
// @Param body body alert.AlertFilter false "alert filter"
// @Success 200 object response.APIResponse{data=[]alert.Alert}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/alarm/alert/alerts [POST]
// @Security ApiKeyAuth
func ListAlerts(_ *gin.Context) ([]alert.Alert, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}

// @ID ListSilencers
// @Tags Alarm
// @Summary List alarm silencers
// @Description List alarm silencers, filter with alarm objects and keywords.
// @Accept application/json
// @Produce application/json
// @Param body body silence.SilencerFilter false "silencer filter"
// @Success 200 object response.APIResponse{data=[]silence.SilencerResponse}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/alarm/silence/silencers [POST]
// @Security ApiKeyAuth
func ListSilencers(_ *gin.Context) ([]silence.SilencerResponse, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}

// @ID GetSilencer
// @Tags Alarm
// @Summary Get alarm silencer
// @Description Get alarm silencer, query by silencer id.
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=silence.SilencerResponse}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param id path string true "silencer id"
// @Router /api/v1/alarm/silence/silencers/{id} [GET]
// @Security ApiKeyAuth
func GetSilencer(_ *gin.Context) (*silence.SilencerResponse, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}

// @ID CreateOrUpdateSilencer
// @Tags Alarm
// @Summary Create or update alarm silencer
// @Description Create or update alarm silencer.
// @Accept application/json
// @Produce application/json
// @Param body body silence.SilencerParam true "silencer"
// @Success 200 object response.APIResponse{data=silence.SilencerResponse}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/alarm/silence/silencers [PUT]
// @Security ApiKeyAuth
func CreateOrUpdateSilencer(_ *gin.Context) (*silence.SilencerResponse, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}

// @ID DeleteSilencer
// @Tags Alarm
// @Summary Delete alarm silencer
// @Description Delete alarm silencer by silencer id.
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param id path string true "silencer id"
// @Router /api/v1/alarm/silence/silencers/{id} [DELETE]
// @Security ApiKeyAuth
func DeleteSilencer(_ *gin.Context) (any, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}

// @ID ListRules
// @Tags Alarm
// @Summary List alarm rules
// @Description List alarm rules, filter with alarm objects type, serverity and keywords.
// @Accept application/json
// @Produce application/json
// @Param body body rule.RuleFilter false "rule filter"
// @Success 200 object response.APIResponse{data=[]rule.RuleResponse}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/alarm/rule/rules [POST]
// @Security ApiKeyAuth
func ListRules(_ *gin.Context) ([]rule.RuleResponse, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}

// @ID GetRule
// @Tags Alarm
// @Summary Get alarm rule
// @Description Get alarm rule, query by rule name.
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=rule.RuleResponse}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param name path string true "rule name"
// @Router /api/v1/alarm/rule/rules/{name} [GET]
// @Security ApiKeyAuth
func GetRule(_ *gin.Context) (*rule.RuleResponse, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}

// @ID CreateOrUpdateRule
// @Tags Alarm
// @Summary Create or update alarm rule
// @Description Create or update alarm rule.
// @Accept application/json
// @Produce application/json
// @Param body body rule.Rule true "rule"
// @Success 200 object response.APIResponse{data=rule.RuleResponse}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/alarm/rule/rules [PUT]
// @Security ApiKeyAuth
func CreateOrUpdateRule(_ *gin.Context) (*rule.RuleResponse, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}

// @ID DeleteRule
// @Tags Alarm
// @Summary Delete alarm rule
// @Description Delete alarm rule by rule name.
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param name path string true "rule name"
// @Router /api/v1/alarm/rule/rules/{name} [DELETE]
// @Security ApiKeyAuth
func DeleteRule(_ *gin.Context) (any, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}

// @ID ListReceivers
// @Tags Alarm
// @Summary List alarm receivers
// @Description List alarm receivers, do not support filter, list all receivers at once.
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]receiver.Receiver}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/alarm/receiver/receivers [POST]
// @Security ApiKeyAuth
func ListReceivers(_ *gin.Context) ([]receiver.Receiver, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}

// @ID GetReceiver
// @Tags Alarm
// @Summary Get alarm receiver
// @Description Get alarm receiver, query by receiver name.
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=receiver.Receiver}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param name path string true "rule name"
// @Router /api/v1/alarm/receiver/receivers/{name} [GET]
// @Security ApiKeyAuth
func GetReceiver(_ *gin.Context) (*receiver.Receiver, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}

// @ID CreateOrUpdateReceiver
// @Tags Alarm
// @Summary Create or update alarm receiver
// @Description Create or update alarm receiver.
// @Accept application/json
// @Produce application/json
// @Param body body receiver.Receiver true "receiver"
// @Success 200 object response.APIResponse{data=receiver.Receiver}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/alarm/receiver/receivers [PUT]
// @Security ApiKeyAuth
func CreateOrUpdateReceiver(_ *gin.Context) (*receiver.Receiver, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}

// @ID DeleteReceiver
// @Tags Alarm
// @Summary Delete alarm receiver
// @Description Delete alarm receiver by receiver name.
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param name path string true "receiver name"
// @Router /api/v1/alarm/receiver/receivers/{name} [DELETE]
// @Security ApiKeyAuth
func DeleteReceiver(_ *gin.Context) (any, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}

// @ID ListReceiverTemplates
// @Tags Alarm
// @Summary List alarm receiver templates
// @Description List alarm receiver templates.
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]receiver.Template}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/alarm/receiver/templates [POST]
// @Security ApiKeyAuth
func ListReceiverTemplates(_ *gin.Context) ([]receiver.Template, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}

// @ID GetReceiverTemplate
// @Tags Alarm
// @Summary Get alarm receiver template
// @Description Get alarm receiver template.
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=receiver.Template}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param type path string true "receiver type"
// @Router /api/v1/alarm/receiver/templates/{type} [GET]
// @Security ApiKeyAuth
func GetReceiverTemplate(_ *gin.Context) (*receiver.Template, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}

// @ID ListRoutes
// @Tags Alarm
// @Summary List alarm routes
// @Description List alarm routes, do not support filter, list all routes at once.
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]route.RouteResponse}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/alarm/route/routes [POST]
// @Security ApiKeyAuth
func ListRoutes(_ *gin.Context) ([]route.RouteResponse, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}

// @ID GetRoute
// @Tags Alarm
// @Summary Get alarm route
// @Description Get alarm route, query by route name.
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=route.RouteResponse}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param id path string true "route id"
// @Router /api/v1/alarm/route/routes/{id} [GET]
// @Security ApiKeyAuth
func GetRoute(_ *gin.Context) (*route.RouteResponse, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}

// @ID CreateOrUpdateRoute
// @Tags Alarm
// @Summary Create or update alarm route
// @Description Create or update alarm route.
// @Accept application/json
// @Produce application/json
// @Param body body route.Route true "route"
// @Success 200 object response.APIResponse{data=route.RouteResponse}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/alarm/route/routes [PUT]
// @Security ApiKeyAuth
func CreateOrUpdateRoute(_ *gin.Context) (*route.RouteResponse, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}

// @ID DeleteRoute
// @Tags Alarm
// @Summary Delete alarm channel
// @Description Delete alarm channel by channel name.
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param id path string true "route id"
// @Router /api/v1/alarm/route/routes/{id} [DELETE]
// @Security ApiKeyAuth
func DeleteRoute(_ *gin.Context) (any, error) {
	return nil, httpErr.NewNotImplemented("not implemented")
}
