package main

import (
	"fmt"
	"net"
	"time"
	"log"
)

// 简单的 UDP 客户端，用于测试隧道功能
func main() {
	// 连接到隧道客户端的监听端口
	conn, err := net.Dial("udp", "127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Println("测试客户端启动，连接到隧道客户端 127.0.0.1:8080")

	// 发送测试消息
	messages := []string{
		"Hello, UDP Tunnel!",
		"这是第二条测试消息",
		"Test message 3",
		"最后一条消息",
	}

	for i, message := range messages {
		fmt.Printf("发送消息 %d: %s\n", i+1, message)
		
		_, err := conn.Write([]byte(message))
		if err != nil {
			log.Printf("发送错误: %v", err)
			continue
		}

		// 读取响应
		buffer := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("读取响应错误: %v", err)
			continue
		}

		response := string(buffer[:n])
		fmt.Printf("收到响应: %s\n\n", response)

		time.Sleep(1 * time.Second)
	}

	fmt.Println("测试完成")
}
