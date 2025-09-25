package main

import (
	"fmt"
	"net"
	"log"
)

// 简单的 UDP 回显服务器，用于测试隧道功能
func main() {
	addr, err := net.ResolveUDPAddr("udp", ":12345")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Println("测试 UDP 服务器启动，监听端口 :12345")
	fmt.Println("这个服务器会回显所有收到的消息")

	buffer := make([]byte, 1024)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("读取错误: %v", err)
			continue
		}

		message := string(buffer[:n])
		fmt.Printf("收到来自 %s 的消息: %s\n", clientAddr.String(), message)

		// 回显消息
		response := fmt.Sprintf("回显: %s", message)
		_, err = conn.WriteToUDP([]byte(response), clientAddr)
		if err != nil {
			log.Printf("发送错误: %v", err)
		}
	}
}
