package handler

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/param"
	"github.com/oceanbase/oceanbase-dashboard/internal/server/constant"
	"github.com/oceanbase/oceanbase-dashboard/internal/store"
	crypto "github.com/oceanbase/oceanbase-dashboard/pkg/crypto"
	oberr "github.com/oceanbase/oceanbase-dashboard/pkg/errors"
	"github.com/oceanbase/oceanbase-dashboard/pkg/k8s/client"

	v1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		return "", oberr.NewBadRequest(err.Error())
	}
	credentials, err := getDashboardUserCredentials(c)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			logHandlerError(c, err)
			return "", oberr.NewBadRequest(err.Error())
		} else {
			return "", oberr.NewInternal(err.Error())
		}
	}
	fetchedPwdRaw, exist := credentials.Data[loginParams.Username]
	if !exist {
		return "", oberr.NewBadRequest("username or password is incorrect")
	}
	fetchedPwd := string(fetchedPwdRaw)
	decryptedPwd, err := crypto.DecryptWithPrivateKey(loginParams.Password)
	if err != nil {
		return "", oberr.NewBadRequest(err.Error())
	}
	if fetchedPwd != decryptedPwd {
		return "", oberr.NewBadRequest("username or password is incorrect")
	}
	sess := sessions.Default(c)
	sess.Set("username", loginParams.Username)
	sess.Set("expiration", time.Now().Add(constant.DefaultSessionExpiration*time.Second).Unix())
	if err := sess.Save(); err != nil {
		logHandlerError(c, err)
		return "", oberr.NewInternal(err.Error())
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
		logHandlerError(c, err)
		return "", oberr.NewInternal(err.Error())
	}
	if usernameEntry != nil {
		store.GetCache().Delete(usernameEntry.(string))
	}
	return "logout successfully", nil
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
