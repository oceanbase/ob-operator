package middleware

import (
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/internal/server/constant"
	"github.com/oceanbase/oceanbase-dashboard/internal/store"

	log "github.com/sirupsen/logrus"
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
			c.AbortWithStatusJSON(401, gin.H{
				"message": "login expired, please login again",
			})
			session.Clear()
			err := session.Save()
			if err != nil {
				log.Errorf("failed to save session: %v", err)
				c.AbortWithStatusJSON(500, gin.H{
					"message": "failed to save session",
				})
			}
			store.GetCache().Delete(username)
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
