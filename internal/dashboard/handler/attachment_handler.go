/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package handler

import (
	"fmt"
	"io"
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
// @Produce application/x-gzip
// @Param id path string true "attachment id"
// @Success 200 {file} application/x-gzip
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/attachments/{id} [GET]
// @Security ApiKeyAuth
func DownloadAttachment(c *gin.Context) {
	id := c.Param("id")
	attachmentFile := biz.GetAttachment(id)

	file, err := os.Open(attachmentFile)
	if err != nil {
		c.String(http.StatusNotFound, "file not found")
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		c.String(http.StatusInternalServerError, "can't get file stat")
		return
	}
	c.Header("Content-Type", "application/x-gzip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", id))
	c.Header("Content-Length", fmt.Sprintf("%d", stat.Size()))
	io.Copy(c.Writer, file)
}
