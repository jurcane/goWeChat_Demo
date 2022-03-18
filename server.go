package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

//服务的结构
type Server struct {
	Ip   string
	Port int

	//在线用户
	OnlineMap map[string]*User
	//在线用户map锁
	mapLock sync.RWMutex
	//广播的channel
	Message chan string
}

//创建一个服务
func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}

//监听广播的消息，有消息就发送给所有人
func (srv *Server) ListenMessage() {
	//循环监听广播，没有消息就会阻塞
	for {
		msg := <-srv.Message
		srv.mapLock.Lock()
		for _, user := range srv.OnlineMap {
			user.C <- msg
		}
		srv.mapLock.Unlock()
	}

}

//把消息放到广播里面
func (srv *Server) BroadCast(user *User, msg string) {

	sendMsg := "[" + user.Name + "]" + msg + "\n"

	//把消息放到信道里
	srv.Message <- sendMsg

}

//每当有请求进来时
func (srv *Server) Handler(conn net.Conn) {
	// fmt.Println("连接成功。。。")

	//创建一个用户
	user := NewUser(conn, srv)
	//用户上线行为
	user.Online()
	//超时踢下线
	isLive := make(chan bool)

	//监听用户的输入
	go func() {
		buff := make([]byte, 4096)
		for {
			// fmt.Println("读取输入")
			n, err := conn.Read(buff) //每次循环都要读一次,没有读到就阻塞
			//客户端合法关闭了连接，进行广播通知，放在读取判断之前，因为下线的话无法读取
			if n == 0 {
				//用户下线操作
				user.Offline()
				return
			}
			// time.Sleep(5 * time.Second)
			// fmt.Println("读取输入结束")
			if err != nil && err != io.EOF { //读取错误并且也不是EOF文件末尾
				fmt.Println("conn read err:", err)
				return
			}

			//提取用户的消息并处理，去掉换行符
			msg := string(buff[:n-1])
			//用户发送消息
			user.DoMessage(msg)
			//更新活跃状态
			isLive <- true
		}

	}()

	for {
		select {
		case <-time.After(600 * time.Second):
			// 发送消息给当前用户
			user.sendMessage("你已被踢下线\n")

			//销毁用户的资源
			close(user.C)

			//关闭当前连接
			conn.Close()
			//退出当前的handler

			return //runtime.Goexit() //这个没搞懂

		case <-isLive:
			//活跃状态，继续回头监听
		}
	}

}

//启动
func (srv *Server) Start() {

	// fmt.Println("启动服务socket...")
	//socket listen，这里会开启两个队列，一个syn队列用来接收第一次握手请求
	//一个accept队列用来，第三次握手请求时把syn队列的请求移出，新链接放入到accept队列中
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", srv.Ip, srv.Port))
	if err != nil {
		log.Fatal("net listen err:", err)
		return
	}
	//close listen socket
	defer listener.Close()
	// defer fmt.Println("主程序关闭")

	//启动监听广播的消息
	go srv.ListenMessage()

	for {
		//循环监听端口请求
		// fmt.Println("循环监听消息。。。")
		//accept
		conn, err := listener.Accept() //accept队列进入阻塞状态，队列中有消息时会拷贝套接字到用户态内存，并返回新的sockfd
		// fmt.Println("监听到请求。。。")
		if err != nil {
			fmt.Println("listen accept err:", err)
			continue
		}
		//handle
		go srv.Handler(conn)
	}

}
