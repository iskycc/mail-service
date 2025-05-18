# 构建阶段
FROM golang:1.21-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git make

# 设置工作目录
WORKDIR /app

# 复制依赖定义
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建可执行文件
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o mail-service .

# 运行阶段
FROM alpine:3.19

# 安装运行时依赖
RUN apk add --no-cache tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 设置工作目录
WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/mail-service .

# 暴露端口
EXPOSE 22125

# 设置环境变量默认值
ENV REDIS_HOST=redis-server \
    REDIS_PORT=6379 \
    REDIS_PWD="" \
    MYSQL_HOST=mysql-server \
    MYSQL_PORT=3306 \
    MYSQL_USER=libv \
    MYSQL_PWD=MiaoDiapp \
    MYSQL_DBNAME=libv

# 启动命令
CMD ["/app/mail-service"]
