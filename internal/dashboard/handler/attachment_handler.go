package handler

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	biz "github.com/oceanbase/ob-operator/internal/dashboard/business/attachment"
)

// @ID DownloadAttachment
// @Summary Download attachment
// @Description Download attachment by id
// @Tags Attachment
// @Accept application/json
// @Produce application/zip
// @Param id path string true "attachment id"
// @Success 200 {file} application/zip
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/attachments/{id} [GET]
// @Security ApiKeyAuth
func DownloadAttachment(c *gin.Context) {
	id := c.Param("id")
	zipFile, err := biz.GetAttachment(id)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to get attachment: %s", err.Error())
		return
	}
	defer os.Remove(zipFile)
	c.File(zipFile)
}
