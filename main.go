package main

import "fmt"

func main() {

	fmt.Println("服务运行中...")
	server := NewServer("127.0.0.1", 9090)
	server.Start()
}
