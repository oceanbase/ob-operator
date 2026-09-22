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

package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/ob-operator/internal/dashboard/server/constant"
	"github.com/oceanbase/ob-operator/internal/store"
)

// authentication

// Match registered routes, not request text: query values and path parameters
// must never make a protected endpoint anonymous.
func isAnonymousRoute(c *gin.Context) bool {
	switch c.FullPath() {
	case "/api/v1/info", "/api/v1/monitor/endpoints":
		return c.Request.Method == http.MethodGet
	case "/api/v1/login", "/api/v1/webhook/alert/log", "/api/v1/auth/:token":
		return c.Request.Method == http.MethodPost
	default:
		return false
	}
}

func LoginRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isAnonymousRoute(c) {
			c.Next()
			return
		}
		session := sessions.Default(c)
		username, ok := session.Get("username").(string)
		if !ok || username == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"message": "login required",
			})
			return
		}

		_, exist := store.GetCache().Load(username)
		if !exist {
			c.AbortWithStatusJSON(401, gin.H{
				"message": "login required",
			})
			return
		}

		expr, ok := session.Get("expiration").(int64)
		if !ok || expr < 0 {
			c.AbortWithStatusJSON(403, gin.H{
				"message": "cookie broken",
			})
			return
		}
		expriration := time.Unix(expr, 0)
		if expriration.Before(time.Now()) {
			session.Clear()
			session.Options(sessions.Options{Path: "/", MaxAge: -1}) // this sets the cookie with a MaxAge of 0
			err := session.Save()
			if err != nil {
				log.Errorf("Failed to save session: %v", err)
				c.AbortWithStatusJSON(500, gin.H{
					"message": "failed to save session",
				})
			}
			store.GetCache().Delete(username)
			c.AbortWithStatusJSON(401, gin.H{
				"message": "login expired, please login again",
			})
			return
		}
		c.Set("username", username)
		c.Next()
	}
}

func RefreshExpiration() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isAnonymousRoute(c) {
			c.Next()
			return
		}
		session := sessions.Default(c)
		expiration := time.Now().Add(constant.DefaultSessionExpiration * time.Second)
		session.Set("expiration", expiration.Unix())
		err := session.Save()
		if err != nil {
			log.Errorf("Failed to save session: %v", err)
			c.AbortWithStatusJSON(500, gin.H{
				"message": "failed to save session",
			})
			return
		}
		c.Next()
	}
}
