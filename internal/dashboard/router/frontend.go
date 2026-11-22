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

// Package router registers the Dashboard HTTP API and serves the web frontend.
package router

import (
	"net/http"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

// Revalidate the entry page and assets on each load. Together with content-hashed
// build output, this prevents a new release from reusing an old router/menu bundle.
func serveFrontend(directory string) gin.HandlerFunc {
	fs := static.LocalFile(directory, false)
	serve := static.Serve("/", fs)
	return func(c *gin.Context) {
		if (c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead) && fs.Exists("/", c.Request.URL.Path) {
			c.Header("Cache-Control", "no-cache, must-revalidate")
		}
		serve(c)
	}
}
