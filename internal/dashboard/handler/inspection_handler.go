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

	insbiz "github.com/oceanbase/ob-operator/internal/dashboard/business/inspection"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/inspection"
	"github.com/oceanbase/ob-operator/pkg/errors"
)

// @ID ListInspectionPolicies
// @Summary list inspection policies
// @Description list inspection policies
// @Tags Inspection
// @Accept application/json
// @Produce application/json
// @Param namespace query string false "Namespace" string
// @Param name query string false "Object name" string
// @Param obclusterName query string false "obcluster name" string
// @Success 200 object response.APIResponse{data=[]inspection.Policy}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/inspection/policies [GET]
// @Security ApiKeyAuth
func ListInspectionPolicies(c *gin.Context) ([]inspection.Policy, error) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	obclusterName := c.Query("obclusterName")
	return insbiz.ListInspectionPolicies(c, namespace, name, obclusterName)
}

// @ID CreateOrUpdateInspectionPolicy
// @Summary create or update inspection policy
// @Description create or update inspection policy
// @Tags Inspection
// @Accept application/json
// @Produce application/json
// @Param body body inspection.Policy true "inspection policy"
// @Success 200 object response.APIResponse{data=inspection.Policy}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/inspection/policies [POST]
// @Security ApiKeyAuth
func CreateOrUpdateInspectionPolicy(_ *gin.Context) (*inspection.Policy, error) {
	return nil, errors.NewNotImplemented("")
}

// @ID GetInspectionPolicy
// @Summary get inspection policy
// @Description get inspection policy
// @Tags Inspection
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Success 200 object response.APIResponse{data=inspection.Policy}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/inspection/policies/{namespace}/{name} [GET]
// @Security ApiKeyAuth
func GetInspectionPolicy(c *gin.Context) (*inspection.Policy, error) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	return insbiz.GetInspectionPolicy(c, namespace, name)
}

// @ID DeleteInspectionPolicy
// @Summary delete inspection policy
// @Description delete inspection policy
// @Tags Inspection
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obcluster namespace"
// @Param name path string true "obcluster name"
// @Param scenario path string true "scenario"
// @Success 200 object response.APIResponse{data=bool}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/inspection/policies/{namespace}/{name}/{scenario} [DELETE]
// @Security ApiKeyAuth
func DeleteInspectionPolicy(c *gin.Context) (bool, error) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	scenario := c.Param("scenario")
	err := insbiz.DeleteInspectionPolicy(c, namespace, name, scenario)
	return err == nil, err
}

// @ID TriggerInspection
// @Summary trigger inspection
// @Description trigger inspection
// @Tags Inspection
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=inspection.Policy}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/inspection/policies/{namespace}/{name}/{scenario} [POST]
// @Security ApiKeyAuth
func TriggerInspection(_ *gin.Context) (*inspection.Policy, error) {
	return nil, errors.NewNotImplemented("")
}

// @ID ListInspectionReports
// @Summary list inspection reports
// @Description list inspection reports
// @Tags Inspection
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]inspection.ReportBriefInfo}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/inspection/reports [GET]
// @Security ApiKeyAuth
func ListInspectionReports(_ *gin.Context) ([]inspection.ReportBriefInfo, error) {
	return nil, errors.NewNotImplemented("")
}

// @ID GetInspectionReport
// @Summary get inspection report
// @Description get inspection report
// @Tags Inspection
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=inspection.Report}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/inspection/reports/{id} [GET]
// @Security ApiKeyAuth
func GetInspectionReport(_ *gin.Context) (*inspection.Report, error) {
	return nil, errors.NewNotImplemented("")
}
