package middleware

import (
	"net/http"
	"net/url"

	"ssssshql/oidc-simple/auth"
	"ssssshql/oidc-simple/config"

	"github.com/gin-gonic/gin"
)

// AuthRequired 验证用户是否已登录的中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否已登录
		if !auth.IsLoggedIn(c) {
			// 保存原始请求 URL
			originalURL := c.Request.URL.RequestURI()

			// 重定向到登录页面，带上原始 URL 作为参数
			if originalURL != "" {
				loginURL := "/login?redirect=" + url.QueryEscape(originalURL)
				c.Redirect(http.StatusFound, loginURL)
			} else {
				c.Redirect(http.StatusFound, "/login")
			}
			c.Abort()
			return
		}

		// 获取用户ID
		userID := auth.GetCurrentUserID(c)
		user := config.GetUser(userID)
		if user == nil {
			// 用户不存在，清除 Cookie 并重定向
			auth.ClearSessionCookie(c)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user", user)
		c.Next()
	}
}
