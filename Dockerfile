# 使用国内镜像源的构建阶段
FROM golang:1.21-alpine AS builder

# 设置 GOPROXY 以提高在中国的下载速度
ENV GOPROXY=https://goproxy.cn,direct

# 设置工作目录
WORKDIR /app

# 复制 go mod 和 sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o audiobookshelf-manager cmd/bot/main.go

# 最终阶段
FROM alpine:latest

# 更换 Alpine 镜像源以提高在中国的下载速度
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

# 安装 ca-certificates 以支持 HTTPS 请求
RUN apk --no-cache add ca-certificates

# 创建工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/audiobookshelf-manager .

# 复制环境变量文件（如果存在）
COPY --from=builder /app/.env .env

# 暴露端口（虽然 Telegram Bot 不需要监听端口，但以防万一）
EXPOSE 8080

# 运行应用
CMD ["./audiobookshelf-manager"]