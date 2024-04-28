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
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/rule"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/silence"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
)

// @ID ListEvents
// @Tags Alarm
// @Summary List alarm events
// @Description List alarm events, filter with alarm objects, serverity, time and keywords.
// @Accept application/json
// @Produce application/json
// @Param body body alert.EventFilter false "event filter"
// @Success 200 object response.APIResponse{data=[]alert.Event}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/alarm/alert/events [POST]
// @Security ApiKeyAuth
func ListEvents(_ *gin.Context) ([]alert.Event, error) {
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
// @Router /api/v1/alarm/silence/silencers [GET]
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
// @Router /api/v1/alarm/rule/rules [GET]
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

// @ID CreateOrUpdateSilencer
// @Tags Alarm
// @Summary Create or update alarm silencer
// @Description Create or update alarm silencer.
// @Accept application/json
// @Produce application/json
// @Param body body rule.RuleParam true "rule"
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
