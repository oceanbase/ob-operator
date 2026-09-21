package router

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
)

func TestSSRoutesRequireLoginAndCollectorRejectsRemoteAccess(t *testing.T) {
	t.Setenv("DEBUG_DASHBOARD", "")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	InitRoutes(r)
	for _, test := range []struct {
		method, path string
		status       int
	}{
		{"GET", "/api/v1/logservices", 401}, {"POST", "/api/v1/logservices", 401},
		{"GET", "/api/v1/logservices/ns/ls", 401}, {"PATCH", "/api/v1/logservices/ns/ls/replicas", 401}, {"DELETE", "/api/v1/logservices/ns/ls", 401},
		{"GET", "/api/v1/logservices/ns/ls/metrics", 401}, {"GET", "/api/v1/obclusters/ns/ob/ss-metrics", 401},
		{"GET", "/internal/ss-metrics", 403},
		{"GET", "/api/v1/objectstorage/ns/credentials", 401},
		{"POST", "/api/v1/objectstorage/ns/credentials", 401},
		{"POST", "/api/v1/objectstorage/ns/check", 401},
	} {
		t.Run(test.method+test.path, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, nil)
			req.RemoteAddr = "192.0.2.1:1234"
			req.Header.Set("X-Forwarded-For", "127.0.0.1")
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)
			if res.Code != test.status {
				t.Fatalf("got %d: %s", res.Code, res.Body.String())
			}
		})
	}
}
