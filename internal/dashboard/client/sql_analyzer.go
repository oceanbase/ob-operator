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

	"github.com/oceanbase/ob-operator/internal/dashboard/model"
	"github.com/pkg/errors"
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

func (c *SqlAnalyzerClient) QuerySqlStats(tenantName string, req model.QuerySqlStatsRequest) (*model.SqlStatsResponse, error) {
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

	var sqlStatsResp model.SqlStatsResponse
	if err := json.Unmarshal(respBody, &sqlStatsResp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response body")
	}

	return &sqlStatsResp, nil
}
