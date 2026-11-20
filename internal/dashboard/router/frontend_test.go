package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestFrontendRevalidatesEntryAndAssets(t *testing.T) {
	directory := t.TempDir()
	for name, content := range map[string]string{
		"index.html":       "<div id=\"root\"></div>",
		"umi.12345678.js":  "console.log('release');",
		"umi.12345678.css": "body{margin:0}",
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	router := gin.New()
	router.Use(serveFrontend(directory))
	router.GET("/api/v1/info", func(c *gin.Context) { c.Status(http.StatusOK) })
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		for _, path := range []string{"/", "/index.html", "/umi.12345678.js", "/umi.12345678.css"} {
			t.Run(method+path, func(t *testing.T) {
				res := httptest.NewRecorder()
				router.ServeHTTP(res, httptest.NewRequest(method, path, nil))
				if got := res.Header().Get("Cache-Control"); got != "no-cache, must-revalidate" {
					t.Fatalf("unexpected cache control: %q", got)
				}
				if res.Code != http.StatusOK && !(path == "/index.html" && res.Code == http.StatusMovedPermanently) {
					t.Fatalf("unexpected HTTP status: %d", res.Code)
				}
				if modified := res.Header().Get("Last-Modified"); modified != "" {
					req := httptest.NewRequest(method, path, nil)
					req.Header.Set("If-Modified-Since", modified)
					cached := httptest.NewRecorder()
					router.ServeHTTP(cached, req)
					if cached.Code != http.StatusNotModified || cached.Header().Get("Cache-Control") != "no-cache, must-revalidate" {
						t.Fatalf("revalidation did not preserve policy: %d %v", cached.Code, cached.Header())
					}
				}
			})
		}
	}
	for _, path := range []string{"/api/v1/info", "/missing.js"} {
		res := httptest.NewRecorder()
		router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
		if res.Header().Get("Cache-Control") != "" {
			t.Fatalf("frontend cache policy leaked to %s", path)
		}
	}
}
