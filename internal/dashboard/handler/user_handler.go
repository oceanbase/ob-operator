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
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	acbiz "github.com/oceanbase/ob-operator/internal/dashboard/business/ac"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/auth"
	acmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/ac"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/server/constant"
	"github.com/oceanbase/ob-operator/internal/store"
	crypto "github.com/oceanbase/ob-operator/pkg/crypto"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
)

// @ID Login
// @Summary User login
// @Description User login and return access token with cookie.
// @Tags User
// @Accept application/json
// @Produce application/json
// @Param loginInfo body param.LoginParam true "login"
// @Success 200 object response.APIResponse{data=ac.Account}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/login [POST]
func Login(c *gin.Context) (*acmodel.Account, error) {
	loginParams := &param.LoginParam{}
	if err := c.BindJSON(loginParams); err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	decryptedPwd, err := crypto.DecryptWithPrivateKey(loginParams.Password)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	acc, err := acbiz.ValidateAccount(c, loginParams.Username, decryptedPwd)
	if err != nil {
		return nil, httpErr.NewBadRequest("username or password is incorrect")
	}
	if acc.LastLoginAt == nil || acc.LastLoginAt.IsZero() {
		acc.NeedReset = true
	}
	sess := sessions.Default(c)
	sess.Set("username", loginParams.Username)
	sess.Set("expiration", time.Now().Add(constant.DefaultSessionExpiration*time.Second).Unix())
	if err := sess.Save(); err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	store.GetCache().Store(loginParams.Username, struct{}{})
	return acc, nil
}

// @ID Logout
// @Summary User logout
// @Description User logout and clear session.
// @Tags User
// @Produce application/json
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/logout [POST]
// @Security ApiKeyAuth
func Logout(c *gin.Context) (string, error) {
	sess := sessions.Default(c)
	usernameEntry := sess.Get("username")
	sess.Clear()
	sess.Options(sessions.Options{Path: "/", MaxAge: -1}) // this sets the cookie with a MaxAge of 0
	if err := sess.Save(); err != nil {
		return "", httpErr.NewInternal(err.Error())
	}
	if usernameEntry != nil {
		store.GetCache().Delete(usernameEntry.(string))
	}
	return "logout successfully", nil
}

// Authorization handler that does not show in swagger
func Authz(c *gin.Context) (*auth.AuthUser, error) {
	urlParam := struct {
		Token auth.Token `uri:"token" binding:"required"`
	}{}

	err := c.BindUri(&urlParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	authUser, ok := auth.ValidateToken(urlParam.Token)
	if !ok {
		return nil, httpErr.NewUnauthorized("invalid token")
	}

	return authUser, nil
}
