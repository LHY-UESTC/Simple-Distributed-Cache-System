# 指定基础镜像为 Ubuntu 20.04，并将其命名为 builder 阶段
FROM ubuntu:20.04 as builder
# 设置工作目录
WORKDIR /app_build
# 将当前目录下的 Simple-Distributed-Cache-System 目录复制到容器的 /app_build 目录下
COPY Simple-Distributed-Cache-System/ /app_build/
# 使用 RUN 指令运行一系列命令
RUN apt-get update && apt-get install --reinstall ca-certificates -y \
    && tar -C /usr/local -xzf go1.20.3.linux-amd64.tar.gz\
    && rm go1.20.3.linux-amd64.tar.gz

# 为go编译器设置env
ENV GOPATH="/root/go"
ENV PATH="/usr/local/go/bin:/root/go/bin:$PATH"

RUN go env -w GOPROXY=https://goproxy.cn \
    && go mod download \
    && go build -o ./Simple-Distributed-Cache-System server.go

# 最终阶段
FROM ubuntu:20.04

WORKDIR /LuoHongyu_server_application

COPY --from=builder /app_build/Simple-Distributed-Cache-System /LuoHongyu_server_application/
