# Simple-Distributed-Cache-System
An Implementation of Distributed System Class [Assignment](https://uestc.feishu.cn/docx/C7ajdHwq9oppWXxhyelcLVvHngc).

Usage:

cd Simple-Distributed-Cache-System
# 生成protoc文件
protoc --go_out=. cache.proto

protoc --go-grpc_out=. cache.proto

go build
# 分别开四个窗口执行下面四条语句
./Simple-Distributed-Cache-System 1

./Simple-Distributed-Cache-System 2

./Simple-Distributed-Cache-System 3

bash sdcs-test.sh 3
