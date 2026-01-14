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

package sql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	analyticmodel "github.com/oceanbase/ob-operator/internal/sql-analyzer/model"
	"github.com/pkg/errors"
)

func QuerySqlStats(host string, tenantName string, req model.QuerySqlStatsRequest) (*model.SqlStatsResponse, error) {
	url := fmt.Sprintf("http://%s:8080/api/v1/tenants/%s/sql-stats", host, tenantName)

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
		Data       *model.SqlStatsResponse `json:"data"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response body")
	}

	if !apiResp.Successful {
		return nil, fmt.Errorf("sql-analyzer returned error: %s", apiResp.Message)
	}

	return apiResp.Data, nil
}

func QueryRequestStatistics(host string, tenantName string, req model.RequestStatisticsRequest) (*model.RequestStatisticsResponse, error) {
	url := fmt.Sprintf("http://%s:8080/api/v1/tenants/%s/request-stats", host, tenantName)

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
		Successful bool                             `json:"successful"`
		Message    string                           `json:"message"`
		Data       *model.RequestStatisticsResponse `json:"data"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response body")
	}

	if !apiResp.Successful {
		return nil, fmt.Errorf("sql-analyzer returned error: %s", apiResp.Message)
	}

	return apiResp.Data, nil
}

func QuerySqlHistory(host string, tenantName string, req model.SqlHistoryRequest) (*model.SqlHistoryResponse, error) {
	url := fmt.Sprintf("http://%s:8080/api/v1/tenants/%s/sql-history", host, tenantName)

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
		Successful bool                      `json:"successful"`
		Message    string                    `json:"message"`
		Data       *model.SqlHistoryResponse `json:"data"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response body")
	}

	if !apiResp.Successful {
		return nil, fmt.Errorf("sql-analyzer returned error: %s", apiResp.Message)
	}

	return apiResp.Data, nil
}

func QuerySqlDetail(host string, tenantName string, req model.SqlDetailRequest) (*model.SqlDetailResponse, error) {
	url := fmt.Sprintf("http://%s:8080/api/v1/tenants/%s/sql-detail", host, tenantName)

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
		Successful bool                     `json:"successful"`
		Message    string                   `json:"message"`
		Data       *model.SqlDetailResponse `json:"data"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response body")
	}

	if !apiResp.Successful {
		return nil, fmt.Errorf("sql-analyzer returned error: %s", apiResp.Message)
	}

	return apiResp.Data, nil
}

func QueryPlanDetail(host string, tenantName string, req analyticmodel.SqlPlanIdentifier) ([]analyticmodel.SqlPlan, error) {
	url := fmt.Sprintf("http://%s:8080/api/v1/tenants/%s/plan_detail", host, tenantName)

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
