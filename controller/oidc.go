package controller

import (
	"encoding/base64"
	"net/http"
	"ssssshql/oidc-simple/auth/jwt"
	"ssssshql/oidc-simple/config"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// 存储授权码的缓存
var (
	authCodeStore = make(map[string]*AuthCodeData)
	codeMutex     sync.RWMutex
)

type AuthCodeData struct {
	UserID    string
	State     string
	ExpiresAt time.Time
}

// Authorize 302重定向并附加code参数
func Authorize(c *gin.Context) {
	redirectURI, _ := c.GetQuery("redirect_uri")
	state, _ := c.GetQuery("state")

	if redirectURI == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "redirect_uri is required"})
		return
	}

	// 获取当前用户（通过中间件已经验证）
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userInterface.(*config.User)

	// 生成授权码
	authCode := jwt.GenerateAuthCode()

	// 存储授权码（有效期10分钟）
	codeMutex.Lock()
	authCodeStore[authCode] = &AuthCodeData{
		UserID:    user.Username,
		State:     state,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	codeMutex.Unlock()

	// 清理过期的授权码
	go cleanExpiredCodes()

	// 拼接重定向 URL
	redirectURL := redirectURI + "?code=" + authCode
	if state != "" {
		redirectURL += "&state=" + state
	}

	// 执行302重定向
	c.Redirect(http.StatusFound, redirectURL)
}

// Token 处理令牌请求
func Token(c *gin.Context) {
	grantType := c.PostForm("grant_type")
	code := c.PostForm("code")

	if grantType != "authorization_code" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "unsupported_grant_type",
			"error_description": "Only authorization_code grant type is supported",
		})
		return
	}

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "code is required",
		})
		return
	}

	// 验证授权码
	codeMutex.RLock()
	codeData, exists := authCodeStore[code]
	codeMutex.RUnlock()

	if !exists || time.Now().After(codeData.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_grant",
			"error_description": "invalid or expired authorization code",
		})
		return
	}

	// 删除已使用的授权码（一次性使用）
	codeMutex.Lock()
	delete(authCodeStore, code)
	codeMutex.Unlock()

	// 获取用户信息
	user := config.GetUser(codeData.UserID)
	if user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "user not found",
		})
		return
	}

	// 生成 Access Token
	accessToken, err := jwt.GenerateAccessToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "failed to generate access token",
		})
		return
	}

	// 生成 ID Token
	idToken, err := jwt.GenerateIDToken(user.Username, user.Email, user.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "failed to generate ID token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"id_token":     idToken,
	})
}

// UserInfo 返回用户信息
func UserInfo(c *gin.Context) {
	// 从上下文中获取用户名（由 Bearer Auth 中间件设置）
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "unauthorized",
			"error_description": "user not found in token",
		})
		return
	}

	// 从配置中获取完整用户信息
	user := config.GetUser(username.(string))
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "user_not_found",
			"error_description": "user not found in database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sub":   user.Username,
		"name":  user.Name,
		"email": user.Email,
	})
}

// JWKS 返回公钥信息（JSON Web Key Set）
func JWKS(c *gin.Context) {
	pubKey := jwt.GetPublicKey()
	if pubKey == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "public key not available"})
		return
	}

	// 将 RSA 公钥转换为 JWK 格式
	n := pubKey.N.Bytes()

	// 编码 e (exponent)
	eBytes := make([]byte, 4)
	eBytes[0] = byte(pubKey.E >> 24)
	eBytes[1] = byte(pubKey.E >> 16)
	eBytes[2] = byte(pubKey.E >> 8)
	eBytes[3] = byte(pubKey.E)

	// 去除前导零
	i := 0
	for i < 3 && eBytes[i] == 0 {
		i++
	}
	eBytes = eBytes[i:]

	// Base64 URL 编码（无填充）
	nBase64 := base64.RawURLEncoding.EncodeToString(n)
	eBase64 := base64.RawURLEncoding.EncodeToString(eBytes)

	c.JSON(http.StatusOK, gin.H{
		"keys": []gin.H{
			{
				"kty": "RSA",
				"alg": "RS256",
				"use": "sig",
				"n":   nBase64,
				"e":   eBase64,
			},
		},
	})
}

// cleanExpiredCodes 清理过期的授权码
func cleanExpiredCodes() {
	codeMutex.Lock()
	defer codeMutex.Unlock()

	now := time.Now()
	for code, data := range authCodeStore {
		if now.After(data.ExpiresAt) {
			delete(authCodeStore, code)
		}
	}
}
