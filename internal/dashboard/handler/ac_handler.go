/*
Copyright (c) 2024 OceanBase
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
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	acbiz "github.com/oceanbase/ob-operator/internal/dashboard/business/ac"
	acmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/ac"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	crypto "github.com/oceanbase/ob-operator/pkg/crypto"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
)

// @ID GetAccountInfo
// @Summary get account info
// @Description get account info
// @Tags AccessControl
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=ac.Account}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/ac/info [GET]
// @Security ApiKeyAuth
func GetAccountInfo(c *gin.Context) (*acmodel.Account, error) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		return nil, httpErr.New(httpErr.ErrUnauthorized, "Unauthorized")
	}
	return acbiz.GetAccount(c, username.(string))
}

// @ID ResetPassword
// @Summary Reset user's own password
// @Description Reset user's own password
// @Tags AccessControl
// @Accept application/json
// @Produce application/json
// @Param resetParam body param.ResetPasswordParam true "reset password"
// @Success 200 object response.APIResponse{data=ac.Account}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/ac/password [POST]
// @Security ApiKeyAuth
func ResetPassword(c *gin.Context) (*acmodel.Account, error) {
	username := c.GetString("username")
	if username == "" {
		return nil, httpErr.NewUnauthorized("unauthorized")
	}
	param := &param.ResetPasswordParam{}
	if err := c.BindJSON(param); err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	decryptedPwd, err := crypto.DecryptWithPrivateKey(param.Password)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	param.Password = decryptedPwd

	if param.OldPassword != "" {
		decryptedPwd, err := crypto.DecryptWithPrivateKey(param.OldPassword)
		if err != nil {
			return nil, httpErr.NewBadRequest(err.Error())
		}
		param.OldPassword = decryptedPwd
	}

	acc, err := acbiz.ResetAccountPassword(c, username, param)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

// @ID ListAllAccounts
// @Summary List all accounts
// @Description List all accounts
// @Tags AccessControl
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]ac.Account}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/ac/accounts [GET]
// @Security ApiKeyAuth
func ListAccounts(c *gin.Context) ([]acmodel.Account, error) {
	return acbiz.ListAccounts(c)
}

// @ID CreateAccount
// @Summary Create an account
// @Description Create an account
// @Tags AccessControl
// @Accept application/json
// @Produce application/json
// @Param account body ac.CreateAccountParam true "Account information"
// @Success 200 object response.APIResponse{data=ac.Account}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/ac/accounts [POST]
// @Security ApiKeyAuth
func CreateAccount(c *gin.Context) (*acmodel.Account, error) {
	param := acmodel.CreateAccountParam{}
	err := c.BindJSON(&param)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	return acbiz.CreateAccount(c, &param)
}

// @ID PatchAccount
// @Summary Patch an account
// @Description Patch an account
// @Tags AccessControl
// @Accept application/json
// @Produce application/json
// @Param account body ac.PatchAccountParam true "Account information"
// @Param username path string true "Username"
// @Success 200 object response.APIResponse{data=ac.Account}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/ac/accounts/{username} [PATCH]
// @Security ApiKeyAuth
func PatchAccount(c *gin.Context) (*acmodel.Account, error) {
	username := c.Param("username")
	if username == "" {
		return nil, httpErr.NewBadRequest("Username is required")
	}
	param := acmodel.PatchAccountParam{}
	err := c.BindJSON(&param)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	return acbiz.PatchAccount(c, username, &param)
}

// @ID DeleteAccount
// @Summary Delete an account
// @Description Delete an account
// @Tags AccessControl
// @Accept application/json
// @Produce application/json
// @Param username path string true "Username"
// @Success 200 object response.APIResponse{data=ac.Account}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/ac/accounts/{username} [DELETE]
// @Security ApiKeyAuth
func DeleteAccount(c *gin.Context) (*acmodel.Account, error) {
	session := sessions.Default(c)
	currentUser := session.Get("username").(string)
	if currentUser == c.Param("username") {
		return nil, httpErr.NewBadRequest("You can't delete yourself")
	}
	username := c.Param("username")
	if username == "" {
		return nil, httpErr.NewBadRequest("Username is required")
	}
	return acbiz.DeleteAccount(c, username)
}

// @ID ListAllRoles
// @Summary List all roles
// @Description List all roles
// @Tags AccessControl
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]ac.Role}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/ac/roles [GET]
// @Security ApiKeyAuth
func ListRoles(c *gin.Context) ([]*acmodel.Role, error) {
	return acbiz.ListRoles(c)
}

// @ID CreateRole
// @Summary Create an role
// @Description Create an role
// @Tags AccessControl
// @Accept application/json
// @Produce application/json
// @Param role body ac.CreateRoleParam true "Role information"
// @Success 200 object response.APIResponse{data=ac.Role}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/ac/roles [POST]
// @Security ApiKeyAuth
func CreateRole(c *gin.Context) (*acmodel.Role, error) {
	param := acmodel.CreateRoleParam{}
	err := c.BindJSON(&param)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	return acbiz.CreateRole(c, &param)
}

// @ID PatchRole
// @Summary Patch an role
// @Description Patch an role
// @Tags AccessControl
// @Accept application/json
// @Produce application/json
// @Param role body ac.PatchRoleParam true "Role information"
// @Param name path string true "Role name"
// @Success 200 object response.APIResponse{data=ac.Role}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/ac/roles/{name} [PATCH]
// @Security ApiKeyAuth
func PatchRole(c *gin.Context) (*acmodel.Role, error) {
	name := c.Param("name")
	if name == "" {
		return nil, httpErr.NewBadRequest("Role name is required")
	}
	param := acmodel.PatchRoleParam{}
	err := c.BindJSON(&param)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	return acbiz.PatchRole(c, name, &param)
}

// @ID DeleteRole
// @Summary Delete an role
// @Description Delete an role
// @Tags AccessControl
// @Accept application/json
// @Produce application/json
// @Param name path string true "Role name"
// @Success 200 object response.APIResponse{data=ac.Role}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/ac/roles/{name} [DELETE]
// @Security ApiKeyAuth
func DeleteRole(c *gin.Context) (*acmodel.Role, error) {
	name := c.Param("name")
	if name == "" {
		return nil, httpErr.NewBadRequest("Role name is required")
	}
	return acbiz.DeleteRole(c, name)
}

// @ID ListAllPolicies
// @Summary List all policies
// @Description List all policies
// @Tags AccessControl
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]ac.Policy}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/ac/policies [GET]
// @Security ApiKeyAuth
func ListAllPolicies(_ *gin.Context) ([]acmodel.Policy, error) {
	return acbiz.AllPolicies, nil
}
