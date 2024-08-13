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
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	accountLineSeparator = " <|:SEP:|> "
)

type AccountCreds struct {
	EncryptedPassword string
	Nickname          string
	LastLoginAtUnix   int64
	Description       string
}

func NewAccountCreds(infoLine string) (*AccountCreds, error) {
	parts := strings.SplitN(infoLine, accountLineSeparator, 4)
	if len(parts) != 4 {
		return nil, errors.New("User credentials file is corrupted: invalid format")
	}
	ts, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, errors.New("User credentials file is corrupted: last login time is not a valid timestamp")
	}
	return &AccountCreds{
		EncryptedPassword: parts[0],
		Nickname:          parts[1],
		LastLoginAtUnix:   ts,
		Description:       strings.TrimSpace(parts[3]),
	}, nil
}

type UpdateAccountCreds struct {
	AccountCreds
	Username string
	Delete   bool
}

// key value format -> admin: pwd <SEP> nickname <SEP> lastLogin <SEP> description
func (u *AccountCreds) ToLine() string {
	lastLoginAtStr := fmt.Sprintf("%d", u.LastLoginAtUnix)
	return strings.Join([]string{u.EncryptedPassword, u.Nickname, lastLoginAtStr, u.Description}, accountLineSeparator)
}
