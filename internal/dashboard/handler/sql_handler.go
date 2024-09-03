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
	"github.com/oceanbase/ob-operator/internal/dashboard/model/sql"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
)

// @ID ListSqlMetrics
// @Summary list sql metrics
// @Description list sqls metrics
// @Tags Sql
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]sql.SqlMetricMetaCategory}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/sql/metrics [GET]
// @Security ApiKeyAuth
func ListSqlMetrics(c *gin.Context) ([]sql.SqlMetricMetaCategory, error) {
	return nil, httpErr.NewNotImplemented("")
}

// @ID ListTopSqls
// @Summary list top sqls
// @Description list top sqls ordering by spcecific metrics
// @Tags Sql
// @Accept application/json
// @Produce application/json
// @Param body body sql.SqlFilter true "sql filter"
// @Success 200 object response.APIResponse{data=[]sql.SqlInfo}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/sql/topSqls [POST]
// @Security ApiKeyAuth
func ListTopSqls(c *gin.Context) ([]sql.SqlInfo, error) {
	return nil, httpErr.NewNotImplemented("")
}

// @ID ListSuspiciousSqls
// @Summary list suspicious sqls
// @Description list suspicious sqls
// @Tags Sql
// @Accept application/json
// @Produce application/json
// @Param body body sql.SqlFilter true "sql filter"
// @Success 200 object response.APIResponse{data=[]sql.SqlInfo}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/sql/suspiciousSqls [POST]
// @Security ApiKeyAuth
func ListSuspiciousSqls(c *gin.Context) ([]sql.SqlInfo, error) {
	return nil, httpErr.NewNotImplemented("")
}

// @ID RequestStatistics
// @Summary list request statistics
// @Description list request statistics
// @Tags Sql
// @Accept application/json
// @Produce application/json
// @Param body body sql.SqlRequestStatisticParam true "sql request statistic param"
// @Success 200 object response.APIResponse{data=[]sql.RequestStatisticInfo}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/sql/requestStatistics [POST]
// @Security ApiKeyAuth
func RequestStatistics(c *gin.Context) ([]sql.RequestStatisticInfo, error) {
	return nil, httpErr.NewNotImplemented("")
}

// @ID QuerySqlDetailInfo
// @Summary query SQL detail info
// @Description query detailed statistic info of a SQL
// @Tags Sql
// @Accept application/json
// @Produce application/json
// @Param body body sql.SqlDetailParam true "param for query detailed sql info"
// @Success 200 object response.APIResponse{data=sql.SqlDetailedInfo}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/sql/querySqlDetailInfo [POST]
// @Security ApiKeyAuth
func QuerySqlDetailInfo(c *gin.Context) (*sql.SqlDetailedInfo, error) {
	return nil, httpErr.NewNotImplemented("")
}

// @ID QueryPlanDetailInfo
// @Summary query plan detail info
// @Description query detailed statistic info of a plan
// @Tags Sql
// @Accept application/json
// @Produce application/json
// @Param body body sql.PlanDetailParam true "param for query detailed plan info"
// @Success 200 object response.APIResponse{data=sql.PlanDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/sql/queryPlanDetailInfo [POST]
// @Security ApiKeyAuth
func QueryPlanDetailInfo(c *gin.Context) (*sql.PlanDetail, error) {
	return nil, httpErr.NewNotImplemented("")
}
