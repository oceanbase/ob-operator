/*
Copyright (c) 2026 OceanBase
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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestListAllTenantsPreservesUnauthorized(t *testing.T) {
	// Exercise the handler and response wrapper without LoginRequired, so a
	// business-layer authentication error must itself remain HTTP 401.
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/obtenants", Wrap(ListAllTenants))
	for _, uri := range []string{"/api/v1/obtenants?ns=info", "/api/v1/obtenants?obcluster=example"} {
		t.Run(uri, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, uri, nil))
			require.Equal(t, http.StatusUnauthorized, response.Code, response.Body.String())
			require.JSONEq(t, `{"data":null,"message":"Error Unauthorized: login required","successful":false}`, response.Body.String())
		})
	}
}
