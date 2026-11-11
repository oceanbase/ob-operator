package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/ac"
	h "github.com/oceanbase/ob-operator/internal/dashboard/handler"
)

func InitObjectStorageRoutes(g *gin.RouterGroup) {
	// Separate capability: cluster readers/writers do not implicitly gain access
	// to a credential inventory, credential creation or signed endpoint probes.
	g.GET("/objectstorage/:namespace/credentials", h.Wrap(h.ListObjectStorageCredentials, ac.PathGuard("objectstorage", ":namespace", "read")))
	g.POST("/objectstorage/:namespace/credentials", h.Wrap(h.CreateObjectStorageCredentials, ac.PathGuard("objectstorage", ":namespace", "write")))
	g.POST("/objectstorage/:namespace/check", h.Wrap(h.CheckObjectStorage, ac.PathGuard("objectstorage", ":namespace", "write")))
}
