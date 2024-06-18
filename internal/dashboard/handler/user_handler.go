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
	"context"
	"errors"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/internal/dashboard/business/auth"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/server/constant"
	"github.com/oceanbase/ob-operator/internal/store"
	crypto "github.com/oceanbase/ob-operator/pkg/crypto"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

// @ID Login
// @Summary User login
// @Description User login and return access token with cookie.
// @Tags User
// @Accept application/json
// @Produce application/json
// @Param loginInfo body param.LoginParam true "login"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/login [POST]
func Login(c *gin.Context) (string, error) {
	loginParams := &param.LoginParam{}
	if err := c.BindJSON(loginParams); err != nil {
		return "", httpErr.NewBadRequest(err.Error())
	}
	credentials, err := getDashboardUserCredentials(c)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return "", httpErr.NewBadRequest(err.Error())
		}
		return "", httpErr.NewInternal(err.Error())
	}
	fetchedPwdRaw, exist := credentials.Data[loginParams.Username]
	if !exist {
		return "", httpErr.NewBadRequest("username or password is incorrect")
	}
	fetchedPwd := string(fetchedPwdRaw)
	decryptedPwd, err := crypto.DecryptWithPrivateKey(loginParams.Password)
	if err != nil {
		return "", httpErr.NewBadRequest(err.Error())
	}
	if fetchedPwd != decryptedPwd {
		return "", httpErr.NewBadRequest("username or password is incorrect")
	}
	sess := sessions.Default(c)
	sess.Set("username", loginParams.Username)
	sess.Set("expiration", time.Now().Add(constant.DefaultSessionExpiration*time.Second).Unix())
	if err := sess.Save(); err != nil {
		return "", httpErr.NewInternal(err.Error())
	}
	store.GetCache().Store(loginParams.Username, struct{}{})
	return "login successfully", nil
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

func Authz(c *gin.Context) (*auth.AuthUser, error) {
	urlParam := struct {
		Token auth.Token `uri:"token" binding:"required"`
	}{}

	err := c.BindUri(&urlParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	// authUser, ok := auth.ValidateToken(urlParam.Token)
	// if !ok {
	// 	return nil, httpErr.NewUnauthorized("invalid token")
	// }

	return &auth.AuthUser{
		Username: "mock",
		Nickname: "mock",
	}, nil
}

func getDashboardUserCredentials(c context.Context) (*v1.Secret, error) {
	credentialSecret, exist := os.LookupEnv("USER_CREDENTIALS_SECRET")
	if !exist || credentialSecret == "" {
		return nil, errors.New("env USER_CREDENTIALS_SECRET is not set")
	}
	ns, exist := os.LookupEnv("USER_NAMESPACE")
	if !exist || ns == "" {
		return nil, errors.New("env USER_NAMESPACE is not set")
	}
	clt := client.GetClient()
	return clt.ClientSet.CoreV1().Secrets(ns).Get(c, credentialSecret, metav1.GetOptions{})
}
