# 构建阶段
FROM golang:1.26.3-alpine AS builder

# 安装必要工具（git 用于 Go 模块代理解析）
RUN apk add --no-cache git

# 设置 Go 代理（构建与运行时统一配置）
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# 安装 swag
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 生成 swagger 文档
RUN swag init -g cmd/main.go -o docs

# 编译
RUN go build -o travel-backend .

# 运行阶段（关键：用固定版本，不走 latest 网络拉取）
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/travel-backend .

# 复制静态文件（关键！）
COPY --from=builder /app/static ./static

EXPOSE 8080
CMD ["./travel-backend"]
