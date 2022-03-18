package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	Conn       net.Conn
	flag       int
}

func NewClient(ip string, port int) *Client {
	//创建一个用户
	client := &Client{
		ServerIp:   ip,
		ServerPort: port,
		flag:       999,
	}

	//连接服务器
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))

	if err != nil {
		fmt.Println("连接服务器失败!", err)
		return nil
	}
	client.Conn = conn

	//返回
	return client

}

func (cli *Client) DealResponse() {
	//一旦cli.Conn有数据，就拷贝到标准输出到用户界面
	io.Copy(os.Stdout, cli.Conn)
	//上面这个io.Copy() 等价于下面的for循环
	// for{
	// 	buff := make([]byte,4096)
	// 	cli.Conn.Read(buff)
	// }

}

//显示菜单
func (cli *Client) Menu() bool {
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		cli.flag = flag
		return true
	}
	fmt.Println(">>请输入合法范围的数字<<")
	return false
}

func (cli *Client) PublicChat() bool {
	var chatMsg string

	fmt.Println("===请输入聊天内容,exit退出===")
	fmt.Scan(&chatMsg)
	for chatMsg != "exit" {
		if len(chatMsg) > 0 {
			_, err := cli.Conn.Write([]byte(chatMsg + "\n"))
			if err != nil {
				fmt.Println("发送信息失败", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println("===请输入聊天内容,exit退出===")
		fmt.Scan(&chatMsg)
	}

	return true

}

//查询在线用户
func (cli *Client) SelectUsers() {
	_, err := cli.Conn.Write([]byte("who\n"))
	if err != nil {
		fmt.Println("cli write err", err)
	}
	return
}

func (cli *Client) PrivateChat() {
	//用户菜单
	time.Sleep(500 * time.Millisecond)
	cli.SelectUsers()

	var romoteName string
	var chatMsg string

	time.Sleep(200 * time.Millisecond)
	fmt.Println("===请输入聊天对象的用户名,exit退出===")

	fmt.Scanln(&romoteName)

	for len(romoteName) > 0 && romoteName != "exit" {
		fmt.Println("===请输入聊天内容,exit退出===")
		fmt.Scanln(&chatMsg)
		chatMsg = chatMsg
		for len(chatMsg) > 0 && chatMsg != "exit" {
			if len(chatMsg) > 0 {
				_, err := cli.Conn.Write([]byte("to|" + romoteName + "|" + chatMsg + "\n"))
				if err != nil {
					fmt.Println("发送信息失败", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println("===请输入聊天内容,exit退出===")
			fmt.Scanln(&chatMsg)
		}

		cli.SelectUsers()
		time.Sleep(200 * time.Millisecond)
		fmt.Println("===请输入聊天对象的用户名,exit退出===")
		fmt.Scanln(&romoteName)
	}

}

func (cli *Client) UpdateName() {
	fmt.Println("请输入用户名")

	fmt.Scanln(&cli.Name)
	sendMsg := "rename|" + cli.Name + "\n"
	_, err := cli.Conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("修改用户名请求失败", err)
	}

}

func (cli *Client) Run() {

	//如果客户端的flag值不为0 则继续循环
	for cli.flag != 0 {
		//循环进入显示菜单，//不等于true再次进入循环重新输入
		for cli.Menu() != true {
		}
		switch cli.flag {
		case 1:
			cli.PublicChat()
		case 2:
			cli.PrivateChat()
		case 3:
			cli.UpdateName()
			break
		case 0:
			fmt.Println("退出")
			break
		}
	}

}

var ServerIp string
var ServerPort int

func init() {
	flag.StringVar(&ServerIp, "ip", "127.0.0.1", "设置服务器ip地址")
	flag.IntVar(&ServerPort, "port", 9091, "设置服务器端口")
}

func main() {
	//参数解析
	flag.Parse()

	//创建一个客户端服务连接
	cli := NewClient(ServerIp, ServerPort)
	if cli == nil {
		fmt.Println(">>>>>连接服务器失败...")
		return
	}

	fmt.Println(">>>>>连接服务器成功...")

	//单独开启一个goroutine处理服务器的消息
	go cli.DealResponse()

	//启动业务
	cli.Run()
}
