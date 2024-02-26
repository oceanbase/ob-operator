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
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/ob-operator/internal/dashboard/server/constant"
	"github.com/oceanbase/ob-operator/internal/store"
)

// authentication

func LoginRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasSuffix(c.Request.RequestURI, "login") || strings.HasSuffix(c.Request.RequestURI, "info") {
			c.Next()
			return
		}
		session := sessions.Default(c)
		if session.Get("username") == nil {
			c.AbortWithStatusJSON(401, gin.H{
				"message": "login required",
			})
			return
		}

		username := session.Get("username").(string)
		_, exist := store.GetCache().Load(username)
		if !exist {
			c.AbortWithStatusJSON(401, gin.H{
				"message": "login required",
			})
			return
		}

		expr := session.Get("expiration")
		if expr == nil || expr.(int64) < 0 {
			c.AbortWithStatusJSON(403, gin.H{
				"message": "cookie broken",
			})
			return
		}
		expriration := time.Unix(expr.(int64), 0)
		if expriration.Before(time.Now()) {
			session.Clear()
			session.Options(sessions.Options{Path: "/", MaxAge: -1}) // this sets the cookie with a MaxAge of 0
			err := session.Save()
			if err != nil {
				log.Errorf("failed to save session: %v", err)
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
		c.Next()
	}
}

func RefreshExpiration() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasSuffix(c.Request.RequestURI, "login") || strings.HasSuffix(c.Request.RequestURI, "info") {
			c.Next()
			return
		}
		session := sessions.Default(c)
		expiration := time.Now().Add(constant.DefaultSessionExpiration * time.Second)
		session.Set("expiration", expiration.Unix())
		err := session.Save()
		if err != nil {
			log.Errorf("failed to save session: %v", err)
			c.AbortWithStatusJSON(500, gin.H{
				"message": "failed to save session",
			})
			return
		}
		c.Next()
	}
}
