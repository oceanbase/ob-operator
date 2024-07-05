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
	role, err := getAccountRole(username)
	if err != nil {
		return nil, err
	}
	account.Role = *role
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
	if account.password != password {
		return nil, httpErr.NewBadRequest("username or password is incorrect")
	}
	role, err := getAccountRole(username)
	if err != nil {
		logrus.WithError(err).Error("failed to get role for user")
	} else {
		account.Role = *role
	}
	return &account.Account, nil
}

func CreateAccount(ctx context.Context, param *acmodel.CreateAccountParam) (*acmodel.Account, error) {
	credentials, err := getDashboardUserCredentials(ctx)
	if err != nil {
		return nil, err
	}
	if _, ok := credentials.Data[param.Username]; ok {
		return nil, httpErr.NewBadRequest("username already exists")
	}

	roles, err := enforcer.GetRoleManager().GetRoles(param.RoleName)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return nil, httpErr.NewBadRequest("role does not exist")
	}
	ok, err := enforcer.AddRoleForUser(param.Username, param.RoleName)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, httpErr.NewInternal("failed to add role for user")
	}

	credentials.Data[param.Username] = []byte(strings.Join([]string{param.Password, param.Nickname, "0", param.Description}, " "))
	clt := client.GetClient()
	_, err = clt.ClientSet.CoreV1().Secrets(os.Getenv("USER_NAMESPACE")).Update(ctx, credentials, metav1.UpdateOptions{})
	if err != nil {
		return nil, httpErr.NewInternal("failed to update user credentials")
	}

	acc, err := fetchAccount(credentials, param.Username)
	if err != nil {
		return nil, err
	}
	return &acc.Account, nil
}

func PatchAccount(ctx context.Context, username string, param *acmodel.PatchAccountParam) (*acmodel.Account, error) {
	credentials, err := getDashboardUserCredentials(ctx)
	if err != nil {
		return nil, err
	}
	acc, err := fetchAccount(credentials, username)
	if err != nil {
		return nil, err
	}
	if param.RoleName != "" && acc.Role.Name != param.RoleName {
		roles, err := enforcer.GetRoleManager().GetRoles(param.RoleName)
		if err != nil {
			return nil, err
		}
		if len(roles) == 0 {
			return nil, httpErr.NewBadRequest("role does not exist")
		}
		ok, err := enforcer.DeleteRoleForUser(username, acc.Role.Name)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, httpErr.NewInternal("failed to delete role for user")
		}
		ok, err = enforcer.AddRoleForUser(username, param.RoleName)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, httpErr.NewInternal("failed to add role for user")
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
		credentials.Data[username] = []byte(strings.Join([]string{acc.password, acc.Nickname, strconv.FormatInt(time.Now().Unix(), 10), acc.Description}, " "))
		clt := client.GetClient()
		_, err = clt.ClientSet.CoreV1().Secrets(os.Getenv("USER_NAMESPACE")).Update(ctx, credentials, metav1.UpdateOptions{})
		if err != nil {
			return nil, httpErr.NewInternal("failed to update user credentials")
		}
	}
	return &acc.Account, nil
}

func DeleteAccount(ctx context.Context, username string) (*acmodel.Account, error) {
	credentials, err := getDashboardUserCredentials(ctx)
	if err != nil {
		return nil, err
	}
	acc, err := fetchAccount(credentials, username)
	if err != nil {
		return nil, err
	}
	ok, err := enforcer.DeleteRoleForUser(username, acc.Role.Name)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, httpErr.NewInternal("failed to delete role for user")
	}

	delete(credentials.Data, username)
	clt := client.GetClient()
	_, err = clt.ClientSet.CoreV1().Secrets(os.Getenv("USER_NAMESPACE")).Update(ctx, credentials, metav1.UpdateOptions{})
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
	lastLoginAt := time.Unix(ts, 0)
	return &account{
		Account: acmodel.Account{
			Username:    username,
			Nickname:    parts[1],
			LastLoginAt: &lastLoginAt,
			Description: parts[3],
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

func getAccountRole(username string) (*acmodel.Role, error) {
	roles, err := enforcer.GetRolesForUser(username)
	if err != nil {
		return nil, err
	}
	policyLines, err := enforcer.GetPermissionsForUser(username)
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
	return &acmodel.Role{
		Name:     roles[0],
		Policies: policies,
	}, nil
}
