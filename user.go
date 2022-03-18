package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

//创建一个用户
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	//启动监听广播消息
	go user.ListenMsg()

	return user
}

//用户上线
func (user *User) Online() {
	//把用户加入到map里面
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()

	//对上线消息进行广播
	user.server.BroadCast(user, "上线了")
}

//用户下线
func (user *User) Offline() {
	//把用户加入到map里面
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()

	//对下线消息进行广播
	user.server.BroadCast(user, "下线了")
}

//把消息发送给当前用户
func (user *User) sendMessage(msg string) {
	user.conn.Write([]byte(msg + "\n"))
}

//获取在线的用户
func (user *User) GetOnlineUser() {
	user.server.mapLock.Lock()
	for _, v := range user.server.OnlineMap {
		msg := "[" + v.Addr + "]" + v.Name + "在线..."
		user.sendMessage(msg)
	}
	user.server.mapLock.Unlock()
}

//用户重命名
func (user *User) Rename(newName string) {

	user.server.mapLock.Lock()
	_, ok := user.server.OnlineMap[newName]
	user.server.mapLock.Unlock()

	if ok {
		user.sendMessage("用户名已存在！")
		return
	} else {
		user.server.mapLock.Lock()
		delete(user.server.OnlineMap, user.Name)
		user.server.OnlineMap[newName] = user
		user.server.mapLock.Unlock()

		user.Name = newName
		user.sendMessage("您已更新用户名：" + user.Name)
	}
}

//发送消息给某人
func (user *User) SendMsgTo(remoteName, cont string) {

	//判断用户是否存在
	remoteUser, ok := user.server.OnlineMap[remoteName]
	if !ok {
		user.sendMessage("该用户不存在！")
		return
	}

	remoteUser.sendMessage("[" + user.Name + "]" + cont)

}

//用户输入后的信息处理
func (user *User) DoMessage(msg string) {

	if len(msg) < 1 {
		return
	}

	if msg == "who" {
		user.GetOnlineUser()
		return
	}

	//rename|张三
	if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]
		user.Rename(newName)
		return
	}
	//to|李四|消息内容
	if len(msg) > 4 && msg[:3] == "to|" {

		//获取用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			user.sendMessage("消息格式不正确，请使用\"to|张三|你好啊\"格式。")
			return
		}

		//判断消息是否为空
		cont := strings.Split(msg, "|")
		if len(cont) < 3 {
			user.sendMessage("消息格式不正确，请使用\"to|张三|你好啊\"格式。")
			return
		}
		if cont[2] == "" {
			user.sendMessage("无消息内容，请重发")
			return
		}

		user.SendMsgTo(remoteName, cont[2])
		return
	}

	user.server.BroadCast(user, msg)
}

//监听广播的消息，得到消息发送给用户
func (user *User) ListenMsg() {

	for {
		msg := <-user.C
		user.sendMessage(msg)
	}

}
