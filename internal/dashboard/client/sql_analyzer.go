/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pkg/errors"
	logger "github.com/sirupsen/logrus"

	apimodel "github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	analyticmodel "github.com/oceanbase/ob-operator/internal/sql-analyzer/model"
)

type SqlAnalyzerClient struct {
	Host string
}

func NewSqlAnalyzerClient() *SqlAnalyzerClient {
	host := os.Getenv("SQL_ANALYZER_SERVICE_HOST")
	if host == "" {
		host = "http://localhost:8080" // Default for local development
	}
	return &SqlAnalyzerClient{
		Host: host,
	}
}

func NewClient(host string) *SqlAnalyzerClient {
	return &SqlAnalyzerClient{
		Host: host,
	}
}

func (c *SqlAnalyzerClient) QuerySqlStats(tenantName string, req apimodel.QuerySqlStatsRequest) (*apimodel.SqlStatsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/%s/sql-stats", c.Host, tenantName)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request body")
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		logger.Errorf("failed to send request to sql-analyzer: %v", err)
		return &apimodel.SqlStatsResponse{Items: []apimodel.SqlStatsItem{}, TotalCount: 0}, nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("failed to read response body from sql-analyzer: %v", err)
		return &apimodel.SqlStatsResponse{Items: []apimodel.SqlStatsItem{}, TotalCount: 0}, nil
	}

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("sql-analyzer returned non-200 status: %d, body: %s", resp.StatusCode, string(respBody))
		return &apimodel.SqlStatsResponse{Items: []apimodel.SqlStatsItem{}, TotalCount: 0}, nil
	}

	var apiResp struct {
		Successful bool                       `json:"successful"`
		Message    string                     `json:"message"`
		Data       *apimodel.SqlStatsResponse `json:"data"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		logger.Errorf("failed to unmarshal response body from sql-analyzer: %v, body: %s", err, string(respBody))
		return &apimodel.SqlStatsResponse{Items: []apimodel.SqlStatsItem{}, TotalCount: 0}, nil
	}

	if !apiResp.Successful {
		return nil, fmt.Errorf("sql-analyzer returned error: %s", apiResp.Message)
	}

	return apiResp.Data, nil
}

func (c *SqlAnalyzerClient) QueryRequestStatistics(tenantName string, req apimodel.RequestStatisticsRequest) (*apimodel.RequestStatisticsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/%s/request-stats", c.Host, tenantName)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request body")
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request to sql-analyzer")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sql-analyzer returned non-200 status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var apiResp struct {
		Successful bool                                `json:"successful"`
		Message    string                              `json:"message"`
		Data       *apimodel.RequestStatisticsResponse `json:"data"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response body")
	}

	if !apiResp.Successful {
		return nil, fmt.Errorf("sql-analyzer returned error: %s", apiResp.Message)
	}

	return apiResp.Data, nil
}

func (c *SqlAnalyzerClient) QuerySqlHistory(tenantName string, req apimodel.SqlHistoryRequest) (*apimodel.SqlHistoryResponse, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/%s/sql-history", c.Host, tenantName)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request body")
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request to sql-analyzer")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sql-analyzer returned non-200 status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var apiResp struct {
		Successful bool                         `json:"successful"`
		Message    string                       `json:"message"`
		Data       *apimodel.SqlHistoryResponse `json:"data"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response body")
	}

	if !apiResp.Successful {
		return nil, fmt.Errorf("sql-analyzer returned error: %s", apiResp.Message)
	}

	return apiResp.Data, nil
}

func (c *SqlAnalyzerClient) QuerySqlDetail(tenantName string, req apimodel.SqlDetailRequest) (*apimodel.SqlDetailResponse, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/%s/sql-detail", c.Host, tenantName)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request body")
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request to sql-analyzer")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sql-analyzer returned non-200 status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var apiResp struct {
		Successful bool                        `json:"successful"`
		Message    string                      `json:"message"`
		Data       *apimodel.SqlDetailResponse `json:"data"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response body")
	}

	if !apiResp.Successful {
		return nil, fmt.Errorf("sql-analyzer returned error: %s", apiResp.Message)
	}

	return apiResp.Data, nil
}

func (c *SqlAnalyzerClient) QueryPlanDetail(tenantName string, req analyticmodel.SqlPlanIdentifier) ([]analyticmodel.SqlPlan, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/%s/plan_detail", c.Host, tenantName)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request body")
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request to sql-analyzer")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sql-analyzer returned non-200 status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var apiResp struct {
		Successful bool                    `json:"successful"`
		Message    string                  `json:"message"`
		Data       []analyticmodel.SqlPlan `json:"data"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response body")
	}

	if !apiResp.Successful {
		return nil, fmt.Errorf("sql-analyzer returned error: %s", apiResp.Message)
	}

	return apiResp.Data, nil
}
