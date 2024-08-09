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
	"fmt"
	"strings"
)

type UpdateAccountCreds struct {
	Username          string
	EncryptedPassword string
	Nickname          string
	LastLoginAtUnix   int64
	Description       string
	Delete            bool
}

// key value format -> admin: pwd nickname lastLogin description
func (u *UpdateAccountCreds) String() string {
	lastLoginAtStr := fmt.Sprintf("%d", u.LastLoginAtUnix)
	return strings.Join([]string{u.EncryptedPassword, u.Nickname, lastLoginAtStr, u.Description}, " ")
}
