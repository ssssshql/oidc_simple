package controller

import (
	"net/http"
	"ssssshql/oidc-simple/auth"
	"ssssshql/oidc-simple/config"

	"github.com/gin-gonic/gin"
)

// Index 显示首页
func Index(c *gin.Context) {
	// 获取当前用户
	userID := auth.GetCurrentUserID(c)
	user := config.GetUser(userID)

	// 准备模板数据
	data := gin.H{
		"user": user,
	}

	c.HTML(http.StatusOK, "index.html", data)
}
