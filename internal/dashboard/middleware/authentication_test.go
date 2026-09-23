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

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/oceanbase/ob-operator/internal/store"
)

func newAuthenticationTestRouter() (*gin.Engine, *gin.RouterGroup) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery(), sessions.Sessions("cookies", cookie.NewStore([]byte("authentication-test-secret-32byt"))))
	return router, router.Group("/api/v1", LoginRequired(), RefreshExpiration())
}

func TestLoginRequiredRejectsAnonymousProtectedRequests(t *testing.T) {
	for _, tc := range []struct {
		method string
		route  string
		uri    string
	}{
		{http.MethodGet, "/obtenants", "/api/v1/obtenants?ns=readonly-check"},
		{http.MethodGet, "/obtenants", "/api/v1/obtenants?ns=info"},
		{http.MethodGet, "/obtenants", "/api/v1/obtenants?ns=login"},
		{http.MethodGet, "/obtenants", "/api/v1/obtenants?ns=monitor/endpoints"},
		{http.MethodGet, "/obtenants", "/api/v1/obtenants?ns=webhook/alert/log"},
		{http.MethodGet, "/obtenants", "/api/v1/obtenants?ns=%69nfo"},
		{http.MethodGet, "/ac/info", "/api/v1/ac/info"},
		{http.MethodGet, "/obtenants/:namespace/:name", "/api/v1/obtenants/default/info"},
		{http.MethodGet, "/obtenants/:namespace/:name", "/api/v1/obtenants/default/login"},
		{http.MethodPatch, "/info", "/api/v1/info"},
		{http.MethodGet, "/login", "/api/v1/login"},
		{http.MethodPost, "/monitor/endpoints", "/api/v1/monitor/endpoints"},
		{http.MethodGet, "/webhook/alert/log", "/api/v1/webhook/alert/log"},
		{http.MethodGet, "/auth/:token", "/api/v1/auth/token"},
		{http.MethodPost, "/auth/:token/extra", "/api/v1/auth/token/extra"},
	} {
		t.Run(tc.method+" "+tc.uri, func(t *testing.T) {
			router, api := newAuthenticationTestRouter()
			called := false
			api.Handle(tc.method, tc.route, func(c *gin.Context) {
				called = true
				c.Status(http.StatusNoContent)
			})
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(tc.method, tc.uri, nil))
			require.Equal(t, http.StatusUnauthorized, response.Code, response.Body.String())
			require.False(t, called, "protected handler must not run")
			require.Empty(t, response.Result().Cookies(), "rejected requests must not refresh sessions")
		})
	}
}

func TestLoginRequiredAllowsOnlyPublicMethodsAndRoutes(t *testing.T) {
	for _, tc := range []struct {
		method string
		route  string
		uri    string
	}{
		{http.MethodPost, "/login", "/api/v1/login"},
		{http.MethodGet, "/info", "/api/v1/info"},
		{http.MethodGet, "/monitor/endpoints", "/api/v1/monitor/endpoints"},
		{http.MethodPost, "/webhook/alert/log", "/api/v1/webhook/alert/log"},
		{http.MethodPost, "/auth/:token", "/api/v1/auth/token"},
	} {
		for _, query := range []string{"", "?source=dashboard"} {
			t.Run(tc.method+" "+tc.uri+query, func(t *testing.T) {
				router, api := newAuthenticationTestRouter()
				api.Handle(tc.method, tc.route, func(c *gin.Context) {
					c.Status(http.StatusNoContent)
				})
				response := httptest.NewRecorder()
				router.ServeHTTP(response, httptest.NewRequest(tc.method, tc.uri+query, nil))
				require.Equal(t, http.StatusNoContent, response.Code, response.Body.String())
				require.Empty(t, response.Result().Cookies(), "anonymous routes must not create expiration-only cookies")
			})
		}
	}
}

func authenticationTestCookie(t *testing.T, router *gin.Engine, values map[string]interface{}) *http.Cookie {
	t.Helper()
	router.POST("/test-session", func(c *gin.Context) {
		session := sessions.Default(c)
		for key, value := range values {
			session.Set(key, value)
		}
		require.NoError(t, session.Save())
		c.Status(http.StatusNoContent)
	})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/test-session", nil))
	require.Len(t, response.Result().Cookies(), 1)
	return response.Result().Cookies()[0]
}

func TestLoginRequiredValidSessionAndRefresh(t *testing.T) {
	for _, uri := range []string{"/api/v1/obtenants?ns=info", "/api/v1/obtenants?ns=login", "/api/v1/obtenants/default/info", "/api/v1/info"} {
		t.Run(uri, func(t *testing.T) {
			router, api := newAuthenticationTestRouter()
			username := t.Name()
			store.GetCache().Store(username, struct{}{})
			t.Cleanup(func() { store.GetCache().Delete(username) })
			expiration := time.Now().Add(time.Minute).Unix()
			cookie := authenticationTestCookie(t, router, map[string]interface{}{"username": username, "expiration": expiration})
			handler := func(c *gin.Context) {
				require.Equal(t, username, c.GetString("username"))
				require.Greater(t, sessions.Default(c).Get("expiration").(int64), expiration)
				c.Status(http.StatusNoContent)
			}
			api.GET("/obtenants", handler)
			api.GET("/obtenants/:namespace/:name", handler)
			api.PATCH("/info", handler)
			method := http.MethodGet
			if uri == "/api/v1/info" {
				method = http.MethodPatch
			}
			request := httptest.NewRequest(method, uri, nil)
			request.AddCookie(cookie)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, http.StatusNoContent, response.Code, response.Body.String())
			require.Len(t, response.Result().Cookies(), 1, "authenticated requests must persist the refreshed session")
		})
	}
}

func TestLoginRequiredRejectsInvalidSessions(t *testing.T) {
	for _, tc := range []struct {
		name       string
		username   interface{}
		expiration interface{}
		cached     bool
		status     int
	}{
		{"missing identity", nil, time.Now().Add(time.Hour).Unix(), false, http.StatusUnauthorized},
		{"empty identity", "", time.Now().Add(time.Hour).Unix(), true, http.StatusUnauthorized},
		{"wrong identity type", 123, time.Now().Add(time.Hour).Unix(), false, http.StatusUnauthorized},
		{"missing cache", "user", time.Now().Add(time.Hour).Unix(), false, http.StatusUnauthorized},
		{"missing expiration", "user", nil, true, http.StatusForbidden},
		{"wrong expiration type", "user", "tomorrow", true, http.StatusForbidden},
		{"negative expiration", "user", int64(-1), true, http.StatusForbidden},
		{"expired", "user", time.Now().Add(-time.Hour).Unix(), true, http.StatusUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router, api := newAuthenticationTestRouter()
			if tc.cached {
				store.GetCache().Store(tc.username, struct{}{})
				t.Cleanup(func() { store.GetCache().Delete(tc.username) })
			}
			values := map[string]interface{}{}
			if tc.username != nil {
				values["username"] = tc.username
			}
			if tc.expiration != nil {
				values["expiration"] = tc.expiration
			}
			cookie := authenticationTestCookie(t, router, values)
			called := false
			api.GET("/obtenants", func(c *gin.Context) {
				called = true
				c.Status(http.StatusNoContent)
			})
			request := httptest.NewRequest(http.MethodGet, "/api/v1/obtenants?ns=info", nil)
			request.AddCookie(cookie)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, tc.status, response.Code, response.Body.String())
			require.False(t, called)
			if tc.name == "expired" {
				_, exists := store.GetCache().Load(tc.username)
				require.False(t, exists)
				require.Len(t, response.Result().Cookies(), 1)
				require.Negative(t, response.Result().Cookies()[0].MaxAge)
			}
		})
	}
}
