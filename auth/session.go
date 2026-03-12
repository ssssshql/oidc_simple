package auth

import (
	"time"

	"github.com/gin-gonic/gin"
)

const (
	CookieName    = "oidc_session"
	CookieMaxAge  = 30 * 24 * 60 * 60 // 30天
	SessionUserID = "session_user_id"
)

// SetSessionCookie 设置会话 Cookie
func SetSessionCookie(c *gin.Context, userID string) {
	c.SetCookie(
		CookieName,
		userID,
		CookieMaxAge,
		"/",
		"",
		false, // secure
		true,  // httpOnly
	)
}

// GetSessionCookie 获取会话 Cookie
func GetSessionCookie(c *gin.Context) (string, error) {
	return c.Cookie(CookieName)
}

// ClearSessionCookie 清除会话 Cookie
func ClearSessionCookie(c *gin.Context) {
	c.SetCookie(
		CookieName,
		"",
		-1,
		"/",
		"",
		false,
		true,
	)
}

// IsLoggedIn 检查是否已登录
func IsLoggedIn(c *gin.Context) bool {
	cookie, err := GetSessionCookie(c)
	if err != nil || cookie == "" {
		return false
	}
	return true
}

// GetCurrentUserID 获取当前登录用户ID
func GetCurrentUserID(c *gin.Context) string {
	cookie, err := GetSessionCookie(c)
	if err != nil {
		return ""
	}
	return cookie
}

// SetSessionTime 设置会话时间到上下文
func SetSessionTime(c *gin.Context) {
	c.Set("session_time", time.Now())
}
