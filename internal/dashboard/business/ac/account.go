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

package ac

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	acmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/ac"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

type account struct {
	acmodel.Account
	password string
}

func ListAccounts(ctx context.Context) ([]acmodel.Account, error) {
	return fetchAccounts(ctx)
}

func GetAccount(ctx context.Context, username string) (*acmodel.Account, error) {
	credentials, err := getDashboardUserCredentials(ctx)
	if err != nil {
		return nil, err
	}
	account, err := fetchAccount(credentials, username)
	if err != nil {
		return nil, err
	}
	roles, err := getAccountRoles(username)
	if err != nil {
		return nil, err
	}
	account.Roles = roles
	return &account.Account, nil
}

func Enforce(_ context.Context, username string, policy *acmodel.Policy) (bool, error) {
	return enforcer.Enforce(username, policy.Object, policy.Action)
}

func ValidateAccount(ctx context.Context, username, password string) (*acmodel.Account, error) {
	credentials, err := getDashboardUserCredentials(ctx)
	if err != nil {
		return nil, err
	}
	account, err := fetchAccount(credentials, username)
	if err != nil {
		return nil, err
	}
	bts := sha256.Sum256([]byte(password))
	sha256EncodedPwd := hex.EncodeToString(bts[:])
	if account.password != sha256EncodedPwd {
		return nil, httpErr.NewBadRequest("username or password is incorrect")
	}
	roles, err := getAccountRoles(username)
	if err != nil {
		logrus.WithError(err).Error("failed to get role for user")
	} else {
		account.Roles = roles
	}

	now := time.Now().Unix()
	err = updateUserCredentials(ctx, credentials, username, account.password, account.Nickname, strconv.FormatInt(now, 10), account.Description)
	if err != nil {
		logrus.WithError(err).Warn("failed to update user credentials")
	}

	return &account.Account, nil
}

func CreateAccount(ctx context.Context, param *acmodel.CreateAccountParam) (*acmodel.Account, error) {
	enforcer.accMu.Lock()
	defer enforcer.accMu.Unlock()
	credentials, err := getDashboardUserCredentials(ctx)
	if err != nil {
		return nil, err
	}
	if _, ok := credentials.Data[param.Username]; ok {
		return nil, httpErr.NewBadRequest("username already exists")
	}

	roles, err := enforcer.GetFilteredPolicy(0, param.Roles...)
	if err != nil {
		return nil, err
	}
	if len(roles) != len(param.Roles) {
		return nil, httpErr.NewBadRequest("role does not exist")
	}
	for _, role := range param.Roles {
		ok, err := enforcer.AddRoleForUser(param.Username, role)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, httpErr.NewInternal("failed to add role for user")
		}
	}
	bts := sha256.Sum256([]byte(param.Password))
	sha256EncodedPwd := hex.EncodeToString(bts[:])
	err = updateUserCredentials(ctx, credentials, param.Username, sha256EncodedPwd, param.Nickname, "0", param.Description)
	if err != nil {
		return nil, httpErr.NewInternal("failed to update user credentials")
	}
	err = persistPolicies(ctx, enforcer.policyPath, enforcer.configMapPath)
	if err != nil {
		return nil, err
	}
	acc, err := fetchAccount(credentials, param.Username)
	if err != nil {
		return nil, err
	}
	return &acc.Account, nil
}

func PatchAccount(ctx context.Context, username string, param *acmodel.PatchAccountParam) (*acmodel.Account, error) {
	enforcer.accMu.Lock()
	defer enforcer.accMu.Unlock()
	credentials, err := getDashboardUserCredentials(ctx)
	if err != nil {
		return nil, err
	}
	acc, err := fetchAccount(credentials, username)
	if err != nil {
		return nil, err
	}
	if len(param.Roles) > 0 {
		_, err := enforcer.GetFilteredPolicy(0, param.Roles...)
		if err != nil {
			return nil, err
		}
		for _, role := range acc.Roles {
			ok, err := enforcer.DeleteRoleForUser(username, role.Name)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, httpErr.NewInternal("failed to delete role for user")
			}
		}
		for _, role := range param.Roles {
			ok, err := enforcer.AddRoleForUser(username, role)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, httpErr.NewInternal("failed to add role for user")
			}
		}
	}

	accountChanged := false
	if param.Password != "" && acc.password != param.Password {
		acc.password = param.Password
		accountChanged = true
	}
	if param.Nickname != "" && acc.Nickname != param.Nickname {
		acc.Nickname = param.Nickname
		accountChanged = true
	}
	if param.Description != "" && acc.Description != param.Description {
		acc.Description = param.Description
		accountChanged = true
	}
	if accountChanged {
		err = updateUserCredentials(ctx, credentials, username, acc.password, acc.Nickname, strconv.FormatInt(time.Now().Unix(), 10), acc.Description)
		if err != nil {
			return nil, httpErr.NewInternal("failed to update user credentials")
		}
	}
	acc, err = fetchAccount(credentials, username)
	if err != nil {
		return nil, err
	}
	return &acc.Account, nil
}

func DeleteAccount(ctx context.Context, username string) (*acmodel.Account, error) {
	enforcer.accMu.Lock()
	defer enforcer.accMu.Unlock()
	credentials, err := getDashboardUserCredentials(ctx)
	if err != nil {
		return nil, err
	}
	acc, err := fetchAccount(credentials, username)
	if err != nil {
		return nil, err
	}
	for _, role := range acc.Roles {
		ok, err := enforcer.DeleteRoleForUser(username, role.Name)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, httpErr.NewInternal("failed to delete role for user")
		}
	}
	err = persistPolicies(ctx, enforcer.policyPath, enforcer.configMapPath)
	if err != nil {
		return nil, err
	}
	err = updateUserCredentials(ctx, credentials, username, "", "", "0", "")
	if err != nil {
		return nil, httpErr.NewInternal("failed to update user credentials")
	}
	return &acc.Account, nil
}

func getDashboardUserCredentials(c context.Context) (*v1.Secret, error) {
	secretName, exist := os.LookupEnv("USER_CREDENTIALS_SECRET")
	if !exist || secretName == "" {
		return nil, httpErr.NewBadRequest("env USER_CREDENTIALS_SECRET is not set")
	}
	ns, exist := os.LookupEnv("USER_NAMESPACE")
	if !exist || ns == "" {
		return nil, httpErr.NewBadRequest("env USER_NAMESPACE is not set")
	}
	clt := client.GetClient()
	return clt.ClientSet.CoreV1().Secrets(ns).Get(c, secretName, metav1.GetOptions{})
}

func updateUserCredentials(c context.Context, credentials *v1.Secret, username, password, nickname, lastLoginAtUnix, description string) error {
	if password == "" {
		delete(credentials.Data, username)
	} else {
		credentials.Data[username] = []byte(strings.Join([]string{password, nickname, lastLoginAtUnix, description}, " "))
	}
	clt := client.GetClient()
	_, err := clt.ClientSet.CoreV1().Secrets(os.Getenv("USER_NAMESPACE")).Update(c, credentials, metav1.UpdateOptions{})
	return err
}

// pwd nickname lastLogin description
func fetchAccount(credentials *v1.Secret, username string) (*account, error) {
	if _, ok := credentials.Data[username]; !ok {
		return nil, httpErr.NewBadRequest("username or password is incorrect")
	}
	infoLine := string(credentials.Data[username])

	parts := strings.SplitN(infoLine, " ", 4)
	if len(parts) != 4 {
		return nil, httpErr.NewInternal("User credentials file is corrupted: invalid format")
	}
	ts, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, httpErr.NewInternal("User credentials file is corrupted: last login time is not a valid timestamp")
	}
	roles, err := getAccountRoles(username)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return nil, httpErr.NewInternal("User credentials file is corrupted: user has no role")
	}
	lastLoginAt := time.Unix(ts, 0)
	return &account{
		Account: acmodel.Account{
			Username:    username,
			Nickname:    parts[1],
			LastLoginAt: &lastLoginAt,
			Description: parts[3],
			Roles:       roles,
		},
		password: parts[0],
	}, nil
}

func fetchAccounts(ctx context.Context) ([]acmodel.Account, error) {
	credentials, err := getDashboardUserCredentials(ctx)
	if err != nil {
		return nil, err
	}
	accounts := make([]acmodel.Account, 0, len(credentials.Data))
	for username := range credentials.Data {
		account, err := fetchAccount(credentials, username)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account.Account)
	}
	return accounts, nil
}

func getAccountRoles(username string) ([]acmodel.Role, error) {
	// g, username, role
	roles, err := enforcer.GetRolesForUser(username)
	if err != nil {
		return nil, err
	}
	modelRoles := make([]acmodel.Role, 0, len(roles))
	for _, role := range roles {
		// treat role name as username
		// p, role, object, action
		policyLines, err := enforcer.GetPermissionsForUser(role)
		if err != nil {
			return nil, err
		}
		policies := make([]acmodel.Policy, 0, len(policyLines))
		for _, line := range policyLines {
			policies = append(policies, acmodel.Policy{
				Object: acmodel.Object(line[1]),
				Action: acmodel.Action(line[2]),
			})
		}
		modelRoles = append(modelRoles, acmodel.Role{
			Name:     role,
			Policies: policies,
		})
	}
	return modelRoles, nil
}
