package main

import (
	"log"
	"net/http"
	"os"
	"ssssshql/oidc-simple/auth/jwt"
	"ssssshql/oidc-simple/config"
	"ssssshql/oidc-simple/controller"
	"ssssshql/oidc-simple/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化 JWT 密钥（从环境变量或默认路径）
	if err := jwt.InitKeysFromEnv(); err != nil {
		log.Fatalf("初始化 JWT 密钥失败: %v", err)
	}
	log.Println("✓ JWT 密钥初始化成功")

	configPath := "/opt/oidc/config.yml"
	if envPath := os.Getenv("OIDC_CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	if _, err := config.LoadConfig(configPath); err != nil {
		log.Fatalf("加载配置文件失败: %v\n请创建配置文件 %s", err, configPath)
	}
	log.Printf("✓ 成功加载配置文件: %s", configPath)

	router := gin.Default()

	// 从嵌入的文件系统加载模板
	templatesFS, err := GetTemplatesFS()
	if err != nil {
		log.Fatalf("加载模板文件系统失败: %v", err)
	}
	router.HTMLRender = gin.HTMLFS(templatesFS)

	// 静态文件路由（无需认证）
	router.GET("/login", controller.Login)
	router.POST("/login", controller.HandleLogin)
	router.GET("/logout", controller.Logout)
	router.GET("favicon.ico", func(c *gin.Context) {
		favicon, err := GetFavicon()
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "image/x-icon", favicon)
	})

	// JWKS 端点（公钥端点，用于验证 Token）
	router.GET("/.well-known/jwks.json", controller.JWKS)

	// OIDC 端点
	v1 := router.Group("/v1")
	{
		// 授权端点需要用户登录（Cookie 会话认证）
		v1.GET("/authorize", middleware.AuthRequired(), controller.Authorize)

		// Token 端点无需认证（使用 client credentials）
		v1.POST("/token", controller.Token)

		// UserInfo 端点需要 Bearer Token 认证
		v1.GET("/userinfo", middleware.BearerAuth(), controller.UserInfo)
	}

	// 需要认证的路由
	router.GET("/", middleware.AuthRequired(), controller.Index)

	// 启动服务器
	log.Println("====================================")
	log.Println("OIDC 服务启动成功")
	log.Printf("访问地址: http://localhost:8080")
	log.Printf("配置文件: %s", configPath)
	log.Println("====================================")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
