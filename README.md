# OIDC 认证服务

一个轻量级、符合标准的 OpenID Connect 身份认证服务。

## ✨ 特性

- 🎯 **标准兼容** - 完整实现 OAuth 2.0 和 OpenID Connect 协议
- 🔐 **安全可靠** - RS256 签名算法，JWT Token 验证
- 🐳 **容器化** - 支持 Docker 部署，多架构镜像
- 🚀 **开箱即用** - 零配置启动，自动生成密钥
- 📦 **轻量级** - 最小依赖，快速部署

## 🚀 快速开始

### 方式一：直接运行

```bash
# 1. 创建配置目录
sudo mkdir -p /opt/oidc

# 2. 创建用户配置文件
sudo tee /opt/oidc/config.yml <<EOF
users:
  - username: admin
    password: admin123
    email: admin@example.com
    name: 管理员
  - username: user
    password: user123
    email: user@example.com
    name: 普通用户
EOF

# 3. 设置文件权限
sudo chmod 600 /opt/oidc/config.yml

# 4. 运行服务
go run main.go
```

访问 http://localhost:8080/login 即可使用。

### 方式二：Docker 运行

```bash
# 1. 创建配置目录
sudo mkdir -p /opt/oidc

# 2. 创建用户配置文件
sudo tee /opt/oidc/config.yml <<EOF
users:
  - username: admin
    password: admin123
    email: admin@example.com
    name: 管理员
EOF

# 3. 构建镜像
docker build -t oidc-server:latest .

# 4. 运行容器
docker run -d \
  --name oidc-server \
  -p 8080:8080 \
  -v /opt/oidc:/opt/oidc \
  --restart unless-stopped \
  oidc-server:latest
```

### 方式三：使用 Docker Hub 镜像

```bash
# 1. 创建配置目录和文件
sudo mkdir -p /opt/oidc
sudo tee /opt/oidc/config.yml <<EOF
users:
  - username: admin
    password: admin123
    email: admin@example.com
    name: 管理员
EOF

# 2. 拉取并运行
docker run -d \
  --name oidc-server \
  -p 8080:8080 \
  -v /opt/oidc:/opt/oidc \
  --restart unless-stopped \
  your-dockerhub-username/oidc-server:latest
```

## 📁 配置说明

### 用户配置文件 (`/opt/oidc/config.yml`)

```yaml
users:
  # 用户1
  - username: admin          # 用户名
    password: admin123       # 密码（明文）
    email: admin@example.com # 邮箱
    name: 管理员             # 显示名称
  
  # 用户2
  - username: user
    password: user123
    email: user@example.com
    name: 普通用户
  
  # 可以添加更多用户...
```

### 密钥文件

服务启动时会自动生成 RSA 密钥对：

- **私钥**: `/opt/oidc/private.pem` (用于签名 Token)
- **公钥**: `/opt/oidc/public.pem` (用于验证 Token)

⚠️ **重要**：
- 首次启动会自动生成密钥对
- 密钥文件应妥善保管，不要泄露
- 定期更换密钥以提高安全性

## 🔧 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `OIDC_CONFIG_PATH` | 用户配置文件路径 | `/opt/oidc/config.yml` |
| `OIDC_PRIVATE_KEY_PATH` | 私钥文件路径 | `/opt/oidc/private.pem` |
| `OIDC_PUBLIC_KEY_PATH` | 公钥文件路径 | `/opt/oidc/public.pem` |

### 使用环境变量示例

```bash
# Docker Compose
version: '3'
services:
  oidc-server:
    image: oidc-server:latest
    ports:
      - "8080:8080"
    environment:
      - OIDC_CONFIG_PATH=/data/config.yml
      - OIDC_PRIVATE_KEY_PATH=/data/private.pem
      - OIDC_PUBLIC_KEY_PATH=/data/public.pem
    volumes:
      - ./data:/data

# Docker run
docker run -d \
  -p 8080:8080 \
  -e OIDC_CONFIG_PATH=/data/config.yml \
  -v ./data:/data \
  oidc-server:latest
```

## 🌐 OIDC 端点

| 端点 | 方法 | 认证方式 | 说明 |
|------|------|----------|------|
| `/login` | GET | 无 | 登录页面 |
| `/login` | POST | 无 | 处理登录请求 |
| `/logout` | GET | 无 | 退出登录 |
| `/v1/authorize` | GET | Cookie 会话 | OAuth 2.0 授权端点 |
| `/v1/token` | POST | 无 | OAuth 2.0 令牌端点 |
| `/v1/userinfo` | GET | Bearer Token | OpenID Connect 用户信息端点 |
| `/.well-known/jwks.json` | GET | 无 | JWKS 公钥端点 |

## 📖 使用示例

### 1. 完整的 OAuth 2.0 授权码流程

```bash
# 步骤1: 引导用户访问授权端点
# 浏览器访问：
http://localhost:8080/v1/authorize?response_type=code&redirect_uri=http://localhost:3000/callback&client_id=your-client-id&state=random-state&scope=openid%20email

# 步骤2: 用户登录并授权
# 系统会自动跳转到登录页面，用户输入用户名密码

# 步骤3: 获取授权码
# 登录成功后跳转回：
# http://localhost:3000/callback?code=AUTHORIZATION_CODE&state=random-state

# 步骤4: 用授权码换取 Token
curl -X POST http://localhost:8080/v1/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "code=AUTHORIZATION_CODE" \
  -d "redirect_uri=http://localhost:3000/callback" \
  -d "client_id=your-client-id"

# 响应示例：
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "id_token": "eyJhbGciOiJSUzI1NiIs..."
}

# 步骤5: 使用 Access Token 获取用户信息
curl -X GET http://localhost:8080/v1/userinfo \
  -H "Authorization: Bearer ACCESS_TOKEN"

# 响应示例：
{
  "sub": "admin",
  "name": "管理员",
  "email": "admin@example.com"
}
```

### 2. 验证 ID Token

```javascript
// 使用 JWKS 验证 ID Token
const { JWK, JWT } = require('node-jose');

// 1. 获取公钥
const jwks = await fetch('http://localhost:8080/.well-known/jwks.json');
const keyStore = await JWK.asKeyStore(await jwks.json());

// 2. 验证并解码 Token
const result = await JWT.verify(idToken, keyStore);
console.log(result);
// { sub: 'admin', email: 'admin@example.com', name: '管理员', ... }
```

## 🔄 GitHub Actions 自动部署

本项目已配置 GitHub Actions，推送到 GitHub 时会自动构建 Docker 镜像并推送到 Docker Hub。

### 配置步骤

1. **Fork 本仓库**

2. **配置 GitHub Secrets**
   
   在 GitHub 仓库设置中添加以下 Secrets：
   - `DOCKER_USERNAME`: Docker Hub 用户名
   - `DOCKER_PASSWORD`: Docker Hub 密码或访问令牌

3. **修改镜像名称**
   
   编辑 `.github/workflows/docker-publish.yml`，修改 `IMAGE_NAME`：
   ```yaml
   env:
     IMAGE_NAME: your-dockerhub-username/oidc-server
   ```

4. **推送代码触发构建**
   ```bash
   git add .
   git commit -m "Update configuration"
   git push
   ```

5. **使用自动构建的镜像**
   ```bash
   docker run -d \
     -p 8080:8080 \
     -v /opt/oidc:/opt/oidc \
     your-dockerhub-username/oidc-server:latest
   ```

## 🔒 安全建议

### 生产环境部署

1. **使用 HTTPS**
   ```bash
   # 使用反向代理（Nginx/Caddy）
   server {
       listen 443 ssl;
       server_name oidc.yourdomain.com;
       
       ssl_certificate /path/to/cert.pem;
       ssl_certificate_key /path/to/key.pem;
       
       location / {
           proxy_pass http://localhost:8080;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
       }
   }
   ```

2. **限制配置文件权限**
   ```bash
   chmod 600 /opt/oidc/config.yml
   chmod 600 /opt/oidc/private.pem
   chown root:root /opt/oidc/*
   ```

3. **使用强密码**
   ```yaml
   users:
     - username: admin
       password: "Str0ng_P@ssw0rd!2024"  # 使用强密码
       email: admin@example.com
       name: 管理员
   ```

4. **定期更换密钥**
   ```bash
   # 备份旧密钥
   cp /opt/oidc/private.pem /opt/oidc/private.pem.bak
   cp /opt/oidc/public.pem /opt/oidc/public.pem.bak
   
   # 删除旧密钥，服务重启时会自动生成新密钥
   rm /opt/oidc/private.pem /opt/oidc/public.pem
   
   # 重启服务
   docker restart oidc-server
   ```

5. **监控和日志**
   ```bash
   # 查看服务日志
   docker logs -f oidc-server
   
   # 使用 Docker 日志驱动
   docker run -d \
     --log-driver json-file \
     --log-opt max-size=10m \
     --log-opt max-file=3 \
     oidc-server:latest
   ```

## 🛠️ 开发

### 本地开发

```bash
# 克隆仓库
git clone https://github.com/yourusername/oidc-server.git
cd oidc-server

# 安装依赖
go mod download

# 创建配置文件
mkdir -p /opt/oidc
tee /opt/oidc/config.yml <<EOF
users:
  - username: admin
    password: admin123
    email: admin@example.com
    name: 管理员
EOF

# 运行
go run main.go

# 构建
go build -o oidc-server
```

### 项目结构

```
.
├── auth/
│   ├── jwt/          # JWT 工具
│   └── session.go    # 会话管理
├── config/
│   └── user.go       # 配置管理
├── controller/
│   ├── login.go      # 登录控制器
│   ├── oidc.go       # OIDC 控制器
│   └── static.go     # 静态页面
├── middleware/
│   ├── auth.go       # 会话认证中间件
│   └── bearer.go     # Bearer Token 认证中间件
├── templates/
│   ├── login.html    # 登录页面
│   └── index.html    # 首页
├── Dockerfile
├── main.go
└── README.md
```

## 📝 License

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

**如有问题或建议，欢迎提交 Issue**
