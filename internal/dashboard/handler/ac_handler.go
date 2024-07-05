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
	"github.com/gin-gonic/gin"

	acbiz "github.com/oceanbase/ob-operator/internal/dashboard/business/ac"
	acmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/ac"
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
func GetAccountInfo(_ *gin.Context) (*acmodel.Account, error) {
	return nil, httpErr.New(httpErr.ErrNotImplemented, "Not implemented")
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
func ListAccounts(_ *gin.Context) ([]acmodel.Account, error) {
	return nil, httpErr.New(httpErr.ErrNotImplemented, "Not implemented")
}

// @ID CreateAccount
// @Summary Create an account
// @Description Create an account
// @Tags AccessControl
// @Accept application/json
// @Produce application/json
// @Param account body ac.CreateUserParam true "Account information"
// @Success 200 object response.APIResponse{data=ac.Account}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/ac/accounts [POST]
// @Security ApiKeyAuth
func CreateAccount(_ *gin.Context) (*acmodel.Account, error) {
	return nil, httpErr.New(httpErr.ErrNotImplemented, "Not implemented")
}

// @ID PatchAccount
// @Summary Patch an account
// @Description Patch an account
// @Tags AccessControl
// @Accept application/json
// @Produce application/json
// @Param account body ac.PatchUserParam true "Account information"
// @Param username path string true "Username"
// @Success 200 object response.APIResponse{data=ac.Account}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/ac/accounts/{username} [PATCH]
// @Security ApiKeyAuth
func PatchAccount(_ *gin.Context) (*acmodel.Account, error) {
	return nil, httpErr.New(httpErr.ErrNotImplemented, "Not implemented")
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
func DeleteAccount(_ *gin.Context) (*acmodel.Account, error) {
	return nil, httpErr.New(httpErr.ErrNotImplemented, "Not implemented")
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
func ListRoles(_ *gin.Context) ([]acmodel.Role, error) {
	return nil, httpErr.New(httpErr.ErrNotImplemented, "Not implemented")
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
func CreateRole(_ *gin.Context) (*acmodel.Role, error) {
	return nil, httpErr.New(httpErr.ErrNotImplemented, "Not implemented")
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
func PatchRole(_ *gin.Context) (*acmodel.Role, error) {
	return nil, httpErr.New(httpErr.ErrNotImplemented, "Not implemented")
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
func DeleteRole(_ *gin.Context) (*acmodel.Role, error) {
	return nil, httpErr.New(httpErr.ErrNotImplemented, "Not implemented")
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
