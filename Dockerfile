# 构建阶段
FROM golang:1.26-alpine AS builder

WORKDIR /app

# 安装依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建可执行文件
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o oidc-simple .

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 安装 ca-certificates（用于 HTTPS）
RUN apk --no-cache add ca-certificates tzdata

# 创建配置目录
RUN mkdir -p /opt/oidc

# 从构建阶段复制可执行文件
COPY --from=builder /app/oidc-simple .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/assets ./assets

# 暴露端口
EXPOSE 8080

# 设置环境变量
ENV GIN_MODE=release

# 运行服务
CMD ["./oidc-simple"]
