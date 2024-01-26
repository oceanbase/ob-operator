package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/oceanbase/oceanbase-dashboard/internal/model/response"
	crypto "github.com/oceanbase/oceanbase-dashboard/pkg/crypto"
)

var (
	Version    = ""
	CommitHash = ""
	BuildTime  = ""
)

// @ID GetProcessInfo
// @Summary Get process info
// @Description Get process info of OceanBase Dashboard, including process name etc.
// @Tags Info
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.DashboardInfo}
// @Failure 500 object response.APIResponse
// @Router /api/v1/info [GET]
func GetProcessInfo(c *gin.Context) (*response.DashboardInfo, error) {
	pubBytes, err := crypto.PublicKeyToBytes()
	if err != nil {
		logHandlerError(c, err)
		return nil, err
	}
	return &response.DashboardInfo{
		AppName:   "oceanbase-dashboard",
		Version:   strings.Join([]string{Version, CommitHash, BuildTime}, "-"),
		PublicKey: string(pubBytes),
	}, nil
}
