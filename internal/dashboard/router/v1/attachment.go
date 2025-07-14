package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/ob-operator/internal/dashboard/handler"
)

func InitAttachmentRoutes(r *gin.RouterGroup) {
	r.GET("/attachments/:id", handler.DownloadAttachment)
}
