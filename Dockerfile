FROM golang:1.24.0-alpine AS builder
WORKDIR /app

# 安装git，配置Go模块代理
RUN apk add --no-cache git
ENV GOPROXY=https://proxy.golang.org,direct

# 复制依赖清单
COPY go.mod go.sum ./

# 修复：移除-v，改用-x（清理缓存后下载）
RUN go clean -modcache && go mod download -x

# 复制代码并编译（替换为你的真实入口文件）
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o idx-stt ./cmd/main.go

FROM alpine:3.18
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /app/idx-stt .
EXPOSE 9002
CMD ["./idx-stt"]
