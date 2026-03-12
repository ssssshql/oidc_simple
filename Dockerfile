# 构建阶段
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

WORKDIR /app

# 安装 UPX 压缩工具
RUN apk add --no-cache upx

# 安装依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建参数：目标平台
ARG TARGETOS TARGETARCH

# 构建可执行文件（使用 -ldflags="-s -w" 去除调试信息）
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o oidc-simple .

# 使用 UPX 压缩二进制文件（最佳压缩比）
RUN upx --best --lzma oidc-simple

# 运行阶段 - 使用更小的基础镜像
FROM alpine:3.20

WORKDIR /app

# 仅安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata && \
    mkdir -p /opt/oidc

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
