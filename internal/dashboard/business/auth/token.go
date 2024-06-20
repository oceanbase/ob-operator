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

package auth

import "github.com/google/uuid"

func GenerateAuthToken(user *AuthUser) Token {
	uid := uuid.New().String()
	token := Token(uid)
	tokenPool.Add(token, user)
	return token
}

func ValidateToken(token Token) (*AuthUser, bool) {
	user, ok := tokenPool.Get(token)
	if !ok {
		return nil, false
	}
	_ = tokenPool.Remove(token)
	return user, true
}
