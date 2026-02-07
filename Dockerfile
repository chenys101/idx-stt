# 核心修改：编译阶段使用Go 1.24.0 alpine官方镜像
FROM golang:1.24.0-alpine AS builder
WORKDIR /app

# 安装git，配置Go模块代理（海外环境用官方代理，稳定）
RUN apk add --no-cache git
ENV GOPROXY=https://proxy.golang.org,direct

# 复制依赖清单
COPY go.mod go.sum ./

# 保留-v参数（1.24.0支持），清理缓存后下载依赖
RUN go clean -modcache && go mod download -v

# 复制项目代码并编译（替换为你的真实入口文件，如./cmd/main.go）
COPY . .
# 编译为静态二进制文件（CGO禁用，保证可移植性）
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o idx-stt ./cmd/main.go

# 运行阶段（无需修改，仅运行二进制文件）
FROM alpine:3.18
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /app/idx-stt .

# 暴露项目实际端口（根据你的项目调整，如9002）
EXPOSE 9002

# 启动二进制文件
CMD ["./idx-stt"]
