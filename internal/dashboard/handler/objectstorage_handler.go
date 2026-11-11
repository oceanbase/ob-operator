package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	storage "github.com/oceanbase/ob-operator/internal/dashboard/business/objectstorage"
	"github.com/oceanbase/ob-operator/pkg/crypto"
	oberr "github.com/oceanbase/ob-operator/pkg/errors"
)

func bindObjectStorage(c *gin.Context, p any) error {
	d := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 16*1024))
	d.DisallowUnknownFields()
	if err := d.Decode(p); err != nil {
		return oberr.NewBadRequest("Invalid object storage request")
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return oberr.NewBadRequest("Expected one JSON object")
	}
	return nil
}
func ListObjectStorageCredentials(c *gin.Context) ([]storage.CredentialReference, error) {
	return storage.Default().List(c.Request.Context(), c.Param("namespace"))
}
func CreateObjectStorageCredentials(c *gin.Context) (*storage.CredentialReference, error) {
	var p struct {
		Name      string `json:"name"`
		AccessID  string `json:"encryptedAccessID"`
		AccessKey string `json:"encryptedAccessKey"`
	}
	if err := bindObjectStorage(c, &p); err != nil {
		return nil, err
	}
	id, err := crypto.DecryptWithPrivateKey(p.AccessID)
	if err != nil {
		return nil, oberr.NewBadRequest("Cannot decrypt credentials; reload the page and retry")
	}
	key, err := crypto.DecryptWithPrivateKey(p.AccessKey)
	if err != nil {
		return nil, oberr.NewBadRequest("Cannot decrypt credentials; reload the page and retry")
	}
	return storage.Default().Create(c.Request.Context(), c.Param("namespace"), p.Name, id, key)
}
func CheckObjectStorage(c *gin.Context) (*storage.CheckResult, error) {
	var p storage.CheckRequest
	if err := bindObjectStorage(c, &p); err != nil {
		return nil, err
	}
	return storage.Default().Check(c.Request.Context(), c.Param("namespace"), p)
}
