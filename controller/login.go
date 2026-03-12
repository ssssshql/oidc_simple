package controller

import (
	"net/http"
	"net/url"

	"ssssshql/oidc-simple/auth"
	"ssssshql/oidc-simple/config"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

// Login 显示登录页面
func Login(c *gin.Context) {
	// 如果已经登录
	if auth.IsLoggedIn(c) {
		// 检查是否有 redirect 参数
		redirectURL := c.Query("redirect")
		if redirectURL != "" {
			// 重定向到原始 URL
			c.Redirect(http.StatusFound, redirectURL)
			return
		}
		c.Redirect(http.StatusFound, "/")
		return
	}

	// 获取 redirect 参数并传递给模板
	redirectURL := c.Query("redirect")
	c.HTML(http.StatusOK, "login.html", gin.H{
		"redirect": redirectURL,
	})
}

// HandleLogin 处理登录请求
func HandleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码不能为空"})
		return
	}

	// 验证用户
	user := config.ValidateUser(req.Username, req.Password)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 设置会话 Cookie
	auth.SetSessionCookie(c, user.Username)

	// 检查是否有 redirect 参数（从 POST 表单或 URL 查询参数中读取）
	redirectURL := c.DefaultPostForm("redirect", c.Query("redirect"))
	if redirectURL != "" {
		// 解码 URL
		decodedURL, err := url.QueryUnescape(redirectURL)
		if err != nil {
			decodedURL = redirectURL
		}

		// 登录成功，返回原始 URL
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"message":  "登录成功",
			"redirect": decodedURL,
		})
	} else {
		// 登录成功，重定向到首页
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"message":  "登录成功",
			"redirect": "/",
		})
	}
}

// Logout 处理登出请求
func Logout(c *gin.Context) {
	// 清除会话 Cookie
	auth.ClearSessionCookie(c)

	// 重定向到登录页面
	c.Redirect(http.StatusFound, "/login")
}
