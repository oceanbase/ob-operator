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
