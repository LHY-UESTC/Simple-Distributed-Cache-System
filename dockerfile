FROM ubuntu:20.04 as builder

WORKDIR /app_build

COPY Simple-Distributed-Cache-System/ /app_build/

RUN apt-get update && apt-get install --reinstall ca-certificates -y \
    && tar -C /usr/local -xzf go1.20.3.linux-amd64.tar.gz\
    && rm go1.20.3.linux-amd64.tar.gz

# set env for go compiler
ENV GOPATH="/root/go"
ENV PATH="/usr/local/go/bin:/root/go/bin:$PATH"

RUN go env -w GOPROXY=https://goproxy.cn \
    && go mod download \
    && go build -o ./Simple-Distributed-Cache-System server.go

# 最终阶段
FROM ubuntu:20.04

WORKDIR /LuoHongyu_server_application

COPY --from=builder /app_build/Simple-Distributed-Cache-System /LuoHongyu_server_application/
