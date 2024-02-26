/*
Copyright (c) 2023 OceanBase
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
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	crypto "github.com/oceanbase/ob-operator/pkg/crypto"
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
func GetProcessInfo(_ *gin.Context) (*response.DashboardInfo, error) {
	pubBytes, err := crypto.PublicKeyToBytes()
	if err != nil {
		return nil, err
	}
	return &response.DashboardInfo{
		AppName:   "oceanbase-dashboard",
		Version:   strings.Join([]string{Version, CommitHash, BuildTime}, "-"),
		PublicKey: string(pubBytes),
	}, nil
}
