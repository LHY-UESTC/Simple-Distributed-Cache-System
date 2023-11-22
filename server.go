package main

import (
	pb "Simple-Distributed-Cache-System/UESTC-LHY/cache"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var server cacheServer       // 服务器实例
var address [4]string        // 地址
var client [2]pb.CacheClient // 此数组存储了两个客户端对象，用于与两个不同的 RPC 服务器建立通信
var conn [2]*grpc.ClientConn // 存储了两个连接对象，分别用于与两个不同的 RPC 服务器建立连接

func setupClient() {
	// 存储 gRPC 连接选项的切片
	var opts []grpc.DialOption
	var err error
	// 将安全传输凭证选项添加到 opts 中，insecure.NewCredentials() 创建一个不安全的传输凭证，表示在连接时不进行身份验证
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// 使用 opts 和 address[2] 连接到指定的 gRPC 服务器。将返回的连接对象赋值给 conn[0]
	conn[0], err = grpc.Dial(address[2], opts...)
	if err != nil {
		fmt.Printf("dial失败: %v", err)
	}
	fmt.Println("设置客户端:", address[2])
	// 使用 opts 和 address[3] 连接到指定的 gRPC 服务器。将返回的连接对象赋值给 conn[1]
	conn[1], err = grpc.Dial(address[3], opts...)
	if err != nil {
		fmt.Printf("dial失败: %v", err)
	}
	fmt.Println("设置客户端:", address[3])
	// 使用 pb.NewCacheClient() 创建两个 pb.CacheClient 类型的客户端对象，并将其赋值给 client[0] 和 client[1]
	client[0] = pb.NewCacheClient(conn[0])
	client[1] = pb.NewCacheClient(conn[1])
}

// 发送 gRPC 客户端的 Get 请求
func CacheGet(client pb.CacheClient, req *pb.GetRequest) *pb.GetReply {
	// 使用 context.Background() 创建一个空的上下文对象
	// 并使用 context.WithTimeout 方法设置一个超时时间为 10 秒的上下文对象 ctx
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// 延迟执行 cancel 函数，以确保在函数返回之前取消上下文对象，释放相关资源
	defer cancel()
	// 发送 Get 请求
	GetReply, err := client.GetCache(ctx, req)
	if err != nil {
		fmt.Println("client.GetCache 失败")
		return GetReply
	}
	return GetReply
}

// 发送 gRPC 客户端的 Post 请求
func CachePost(client pb.CacheClient, req *pb.PostRequest) *pb.PostReply {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 发送 Post 请求
	PostReply, err := client.PostCache(ctx, req)
	if err != nil {
		fmt.Println("client.PostCache 失败")
		return PostReply
	}
	return PostReply
}

// 发送 gRPC 客户端的 Delete 请求
func CacheDelete(client pb.CacheClient, req *pb.DeleteRequest) *pb.DeleteReply {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 发送 Delete 请求
	DeleteReply, err := client.DeleteCache(ctx, req)
	if err != nil {
		fmt.Println("client.DeleteCache 失败")
		return DeleteReply
	}
	return DeleteReply
}

func setAddress() {
	// 通过服务器索引设置地址变量
	if os.Args[1] == "1" {
		// 此服务器的http服务器端口
		address[0] = "127.0.0.1:9527"
		// 此服务器的rpc服务器端口
		address[1] = "127.0.0.1:9530"
		// 另一个服务器的rpc服务器端口
		address[2] = "127.0.0.1:9531"
		// 另一个服务器的rpc服务器端口
		address[3] = "127.0.0.1:9532"
	} else if os.Args[1] == "2" {
		address[0] = "127.0.0.1:9528"
		address[1] = "127.0.0.1:9531"
		address[2] = "127.0.0.1:9530"
		address[3] = "127.0.0.1:9532"
	} else if os.Args[1] == "3" {
		address[0] = "127.0.0.1:9529"
		address[1] = "127.0.0.1:9532"
		address[2] = "127.0.0.1:9530"
		address[3] = "127.0.0.1:9531"
	} else {
		fmt.Println("只有3个缓存服务器.")
	}
}

// 处理 HTTP GET 请求
// 函数接受两个参数：w 是 http.ResponseWriter 类型，用于构造 HTTP 响应；key 是一个字符串类型的参数，表示请求的键名
func handleGet(w http.ResponseWriter, key string) {
	// fmt.Println("get", key)
	// 检查 server.cache 中是否存在键名为 key 的缓存项
	server.mutex.Lock()
	if value, ok := server.cache[key]; ok {
		server.mutex.Unlock()
		// 将 HTTP 响应的状态码设置为 200 OK
		w.WriteHeader(http.StatusOK)
		// 将 Content-Type 头部设置为 application/json,表示响应的内容类型为 JSON 格式
		w.Header().Set("Content-Type", "application/json")
		// 将 JSON 格式的响应数据写入到 `w` 中，即向客户端返回一个 JSON 字符串，格式为 `{"key":"value"}`
		// 其中 `key` 是请求的键名，`value` 是 `server.cache[key]` 对应的值
		fmt.Fprintln(w, "{\""+key+"\":\""+value+"\"}")
		return
	}
	server.mutex.Unlock()
	// 如果 server.cache 中不存在键名为 key 的缓存项,就通过 gRPC 客户端向两个缓存服务器发送 Get 请求
	GetReply1 := CacheGet(client[0], &pb.GetRequest{Key: key})
	if GetReply1.IsOk == 1 {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, "{\""+key+"\":\""+GetReply1.Value+"\"}")
		return
	}
	GetReply2 := CacheGet(client[1], &pb.GetRequest{Key: key})
	if GetReply2.IsOk == 1 {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, "{\""+key+"\":\""+GetReply2.Value+"\"}")
		return
	}
	// 将 HTTP 响应的状态码设置为 404 Not Found，表示请求的资源不存在
	w.WriteHeader(http.StatusNotFound)
}

// 处理 HTTP POST 请求
func handlePost(w http.ResponseWriter, jsonstr string) {
	// 使用了一个正则表达式，将形如 {"key":"value"} 的 JSON 字符串中的键名和键值提取出来
	reg := regexp.MustCompile(`{\s*"(.*)"\s*:\s*"(.*)"\s*}`)
	if reg == nil {
		fmt.Println("regexp err")
		return
	}
	result := reg.FindAllStringSubmatch(jsonstr, -1)
	key, value := result[0][1], result[0][2]

	// fmt.Println("set", key, ":", value)
	// 存在于当前server中，直接修改即可
	server.mutex.Lock()
	if _, ok := server.cache[key]; ok {
		server.cache[key] = value
		server.mutex.Unlock()
		w.WriteHeader(http.StatusOK)
		return
	}
	server.mutex.Unlock()
	// 可能在其他server中，需要向其他server发送请求
	GetReply1 := CacheGet(client[0], &pb.GetRequest{Key: key})
	if GetReply1.IsOk == 1 {
		PostRely1 := CachePost(client[0], &pb.PostRequest{Key: key, Value: value})
		if PostRely1.IsOk == 1 {
			w.WriteHeader(http.StatusOK)
		}
		return
	}
	GetReply2 := CacheGet(client[1], &pb.GetRequest{Key: key})
	if GetReply2.IsOk == 1 {
		PostRely2 := CachePost(client[1], &pb.PostRequest{Key: key, Value: value})
		if PostRely2.IsOk == 1 {
			w.WriteHeader(http.StatusOK)
		}
		return
	}
	// 说明是个新的key
	server.mutex.Lock()
	server.cache[key] = value
	server.mutex.Unlock()
	// 将 HTTP 响应的状态码设置为 200 OK
	w.WriteHeader(http.StatusOK)
}

// 处理 HTTP DELETE 请求
func handleDelete(w http.ResponseWriter, key string) {
	// fmt.Println("delete", key)
	server.mutex.Lock()
	if _, ok := server.cache[key]; ok {
		// 删除 server.cache 中的对应缓存项
		delete(server.cache, key)
		server.mutex.Unlock()
		// 将 HTTP 响应的状态码设置为 200 OK，表示删除成功
		w.WriteHeader(http.StatusOK)
		// 将字符串 "1" 写入 `w` 中，即向客户端返回一个值为 "1" 的响应
		fmt.Fprintln(w, "1")
		return
	}
	server.mutex.Unlock()
	// 通过 gRPC 客户端向两个缓存服务器发送删除缓存项的请求
	DeleteReply1 := CacheDelete(client[0], &pb.DeleteRequest{Key: key})
	DeleteReply2 := CacheDelete(client[1], &pb.DeleteRequest{Key: key})
	if DeleteReply1.IsOk == 1 || DeleteReply2.IsOk == 1 {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "1")
		return
	}
	// 将字符串 "0" 写入 w 中，即向客户端返回一个值为 "0" 的响应，表示缓存项不存在
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "0")
}

// 处理 HTTP 请求
func handleHttpRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet { // 请求的方法是 GET
		// r.URL.String()[1:] 是请求 URL 的路径部分（去除了第一个字符）
		handleGet(w, r.URL.String()[1:])
	} else if r.Method == http.MethodPost { // 请求的方法是 POST
		// 使用 ioutil.ReadAll 函数读取请求的 body 数据，将其存储在 body 变量中
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			// 如果读取过程中发生错误，会返回一个非空的 err，此时会返回一个 HTTP 500 错误响应
			http.Error(w, "Unable to read request body.", http.StatusInternalServerError)
			return
		}
		handlePost(w, string(body))
	} else if r.Method == http.MethodDelete { // 请求的方法是 DELETE
		handleDelete(w, r.URL.String()[1:])
	} else {
		// 如果请求的方法不是 GET、POST 或 DELETE，则返回一个 HTTP 405 错误响应，表示不支持的请求方法
		http.Error(w, "Unsupport http request.", http.StatusMethodNotAllowed)
	}
}

// cacheServer 的结构体定义
type cacheServer struct {
	// 未实现的 gRPC 服务端接口，表示 cacheServer 结构体实现了 pb.CacheServer 接口，该接口定义了与缓存服务相关的 gRPC 方法
	pb.UnimplementedCacheServer
	// 服务器的内存缓存
	cache map[string]string
	// 互斥锁保护共享内存资源
	mutex sync.Mutex
}

// gRPC 服务器的 Get 请求处理程序，cache_grpc.pb.go中调用
// ctx：context.Context 类型的参数，表示请求的上下文。它提供了请求的元数据和取消信号等功能
// req：*pb.GetRequest 类型的参数，表示 Get 请求的内容。pb.GetRequest 是一个自动生成的结构体类型，包含了 gRPC 定义文件中定义的请求字段
func (s *cacheServer) GetCache(ctx context.Context, req *pb.GetRequest) (*pb.GetReply, error) {
	s.mutex.Lock()
	if value, ok := s.cache[req.Key]; ok {
		s.mutex.Unlock()
		return &pb.GetReply{IsOk: 1, Key: req.Key, Value: value}, nil
	}
	s.mutex.Unlock()
	return &pb.GetReply{IsOk: 0}, nil
}

// gRPC 服务器的 Post 请求处理程序
func (s *cacheServer) PostCache(ctx context.Context, req *pb.PostRequest) (*pb.PostReply, error) {
	s.mutex.Lock()
	s.cache[req.Key] = req.Value
	s.mutex.Unlock()
	return &pb.PostReply{IsOk: 1}, nil
}

// gRPC 服务器的 Delete 请求处理程序
func (s *cacheServer) DeleteCache(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteReply, error) {
	// IsOk 字段为 1 表示成功删除了一个缓存项，为 0 表示未找到要删除的缓存项
	s.mutex.Lock()
	if _, ok := s.cache[req.Key]; ok {
		delete(s.cache, req.Key)
		s.mutex.Unlock()
		return &pb.DeleteReply{IsOk: 1}, nil
	}
	s.mutex.Unlock()
	return &pb.DeleteReply{IsOk: 0}, nil
}

// 启动Http服务器
func startHttpServer() {
	// 当HTTP请求的路径为"/"时，将调用handleHttpRequest函数来处理该请求
	http.HandleFunc("/", handleHttpRequest)
	// HTTP服务器正在监听的地址
	fmt.Println("HTTP服务器正在监听的地址:", address[0])
	// 启动HTTP服务器并开始监听指定的地址
	err := http.ListenAndServe(address[0], nil)
	if err != nil {
		fmt.Println("监听失败:", err)
	}
}

// 启动 gRPC 服务器
func startRpcServer() {
	// gRPC 服务器正在监听指定地址
	fmt.Println("gRPC 服务器正在监听指定地址:", address[1])
	// 创建一个 TCP 监听器 lis，该监听器用于接收客户端的连接请求
	lis, err := net.Listen("tcp", address[1])
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}
	// 函数创建一个 gRPC 服务器实例 grpcServer。通过传入 opts 切片作为参数，可以指定服务器的选项
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	// 将 server 注册到 grpcServer 上，并将该 grpcServer 关联到一个 cacheServer 实例上
	server = cacheServer{cache: make(map[string]string)}
	pb.RegisterCacheServer(grpcServer, &server)
	// 启动 gRPC 服务器，并开始监听来自客户端的请求。这会导致程序阻塞在此处，直到服务器关闭或出现错误
	grpcServer.Serve(lis)
}

func main() {
	// 检查命令行参数的数量是否为2
	if len(os.Args) != 2 {
		fmt.Println("请指定服务器(1-3)")
		return
	}
	// 设置服务器的地址
	setAddress()
	// 两个函数将在独立的goroutine中执行，不会阻塞主程序的继续执行
	go startHttpServer()
	go startRpcServer()
	// 设置客户端
	setupClient()
	// 使用select {}语句创建一个无限循环，会一直阻塞，等待通道的输入，使主程序保持运行状态
	select {}
}
