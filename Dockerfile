# 构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的系统依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建参数，用于指定要构建的服务
ARG SERVICE=api
ARG VERSION=latest
ARG BUILD_TIME
ARG GIT_COMMIT

# 设置构建标志
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# 构建应用
RUN go build -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o /app/bin/app ./cmd/${SERVICE}

# 运行阶段
FROM alpine:latest AS runtime

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata curl

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/bin/app /app/app

# 复制配置文件
COPY --from=builder /app/configs /app/configs

# 设置文件权限
RUN chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["./app"] 