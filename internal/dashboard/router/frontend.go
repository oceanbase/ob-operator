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
