package handler

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestObjectStorageBindRejectsUnknownAndOversizedWithoutEcho(t *testing.T) {
	for _, body := range []string{`{"password":"DO_NOT_ECHO"}`, `{"bucketURL":"x"} {}`, `{"bucketURL":"DO_NOT_ECHO"`, `{"bucketURL":"` + strings.Repeat("DO_NOT_ECHO", 2000) + `"}`} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/", strings.NewReader(body))
		var p struct {
			BucketURL string `json:"bucketURL"`
		}
		if err := bindObjectStorage(c, &p); err == nil || strings.Contains(err.Error(), "DO_NOT_ECHO") {
			t.Fatal("expected sanitized bind error", err)
		}
	}
}

func TestObjectStorageCredentialsRejectPlaintext(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"test","encryptedAccessID":"DO_NOT_ECHO","encryptedAccessKey":"plaintext"}`))
	if _, err := CreateObjectStorageCredentials(c); err == nil || strings.Contains(err.Error(), "DO_NOT_ECHO") {
		t.Fatal("plaintext credentials must be rejected before Kubernetes access", err)
	}
}
