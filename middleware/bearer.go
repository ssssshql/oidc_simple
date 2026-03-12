package middleware

import (
	"net/http"
	"strings"

	"ssssshql/oidc-simple/auth/jwt"

	"github.com/gin-gonic/gin"
)

// BearerAuth 验证 Bearer Token 的中间件
func BearerAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Authorization header 获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":             "missing_authorization_header",
				"error_description": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// 验证 Bearer token 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":             "invalid_authorization_header",
				"error_description": "Authorization header must be Bearer token",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证 Access Token
		userID, err := jwt.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":             "invalid_token",
				"error_description": err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("username", userID)
		c.Next()
	}
}
