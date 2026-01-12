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
	logger "github.com/sirupsen/logrus"

	sqlbiz "github.com/oceanbase/ob-operator/internal/dashboard/business/sql"
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
	lang := c.Query("language")
	return sqlbiz.ListSqlMetrics(lang)
}

// @ID ListSqlStats
// @Summary list top sqls
// @Description list top sqls ordering by spcecific metrics
// @Tags Sql
// @Accept application/json
// @Produce application/json
// @Param body body sql.SqlFilter true "sql filter"
// @Success 200 object response.APIResponse{data=sql.SqlStatsList}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/sql/stats [POST]
// @Security ApiKeyAuth
func ListSqlStats(c *gin.Context) (*sql.SqlStatsList, error) {
	filter := &sql.SqlFilter{}
	err := c.ShouldBindJSON(filter)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	logger.Infof("ListSqlStats filter: %+v", filter)
	res, err := sqlbiz.ListSqlStats(c, filter)
	if err != nil {
		logger.Errorf("ListSqlStats error: %v", err)
		return nil, err
	}
	logger.Infof("ListSqlStats returned %d records", len(res.Items))
	return res, nil
}

// @ID ListRequestStatistics
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
func ListRequestStatistics(c *gin.Context) ([]sql.RequestStatisticInfo, error) {
	param := &sql.SqlRequestStatisticParam{}
	err := c.ShouldBindJSON(param)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	return sqlbiz.ListRequestStatistics(c, param)
}

// @ID QuerySqlHistoryInfo
// @Summary query SQL history info
// @Description query history statistic info of a SQL
// @Tags Sql
// @Accept application/json
// @Produce application/json
// @Param body body sql.SqlHistoryParam true "param for query history sql info"
// @Success 200 object response.APIResponse{data=sql.SqlHistoryInfo}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/sql/querySqlHistoryInfo [POST]
// @Security ApiKeyAuth
func QuerySqlHistoryInfo(c *gin.Context) (*sql.SqlHistoryInfo, error) {
	param := &sql.SqlHistoryParam{}
	if err := c.ShouldBindJSON(param); err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	return sqlbiz.QuerySqlHistoryInfo(c, param)
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
	param := &sql.SqlDetailParam{}
	if err := c.ShouldBindJSON(param); err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	return sqlbiz.QuerySqlDetailInfo(c, param)
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
	param := &sql.PlanDetailParam{}
	if err := c.ShouldBindJSON(param); err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	return sqlbiz.QueryPlanDetailInfo(c, param)
}
