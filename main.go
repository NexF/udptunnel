package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// ===============================
// 主程序入口
// ===============================

// printUsage 打印使用说明
func printUsage() {
	fmt.Println("隧道程序 - 支持 UDP 和 TCP 转发")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  UDP隧道客户端: -mode=client -protocol=udp -local=<UDP监听地址> -remote=<TCP服务端地址>")
	fmt.Println("  UDP隧道服务端: -mode=server -protocol=udp -local=<TCP监听地址> -remote=<UDP目标地址>")
	fmt.Println("  TCP隧道客户端: -mode=client -protocol=tcp -local=<TCP监听地址> -remote=<TCP服务端地址>")
	fmt.Println("  TCP隧道服务端: -mode=server -protocol=tcp -local=<TCP监听地址> -remote=<TCP目标地址>")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  UDP客户端: -mode=client -protocol=udp -local=:8080 -remote=server.example.com:9090")
	fmt.Println("  UDP服务端: -mode=server -protocol=udp -local=:9090 -remote=127.0.0.1:53")
	fmt.Println("  TCP客户端: -mode=client -protocol=tcp -local=:8080 -remote=server.example.com:9090")
	fmt.Println("  TCP服务端: -mode=server -protocol=tcp -local=:9090 -remote=127.0.0.1:22")
	fmt.Println()
	fmt.Println("功能说明:")
	fmt.Println("  UDP隧道:")
	fmt.Println("    - 客户端: 监听本地 UDP 端口，将数据通过 TCP 转发到服务端")
	fmt.Println("    - 服务端: 接收 TCP 连接，将数据转发到目标 UDP 服务")
	fmt.Println("  TCP隧道:")
	fmt.Println("    - 客户端: 监听本地 TCP 端口，将连接通过 TCP 转发到服务端")
	fmt.Println("    - 服务端: 接收 TCP 连接，将连接转发到目标 TCP 服务")
}

// validateArgs 验证命令行参数
func validateArgs(mode, protocol, localAddr, remoteAddr string) error {
	if mode == "" {
		return fmt.Errorf("缺少运行模式参数")
	}
	if protocol == "" {
		return fmt.Errorf("缺少协议类型参数")
	}
	if localAddr == "" {
		return fmt.Errorf("缺少本地地址参数")
	}
	if remoteAddr == "" {
		return fmt.Errorf("缺少远程地址参数")
	}
	if mode != "client" && mode != "server" {
		return fmt.Errorf("无效的运行模式: %s（必须是 'client' 或 'server'）", mode)
	}
	if protocol != "udp" && protocol != "tcp" {
		return fmt.Errorf("无效的协议类型: %s（必须是 'udp' 或 'tcp'）", protocol)
	}
	return nil
}

func main() {
	var (
		mode       = flag.String("mode", "", "运行模式: client 或 server")
		protocol   = flag.String("protocol", "udp", "协议类型: udp 或 tcp (默认: udp)")
		localAddr  = flag.String("local", "", "本地地址")
		remoteAddr = flag.String("remote", "", "远程地址")
		help       = flag.Bool("help", false, "显示帮助信息")
	)
	flag.Parse()

	if *help {
		printUsage()
		os.Exit(0)
	}

	if err := validateArgs(*mode, *protocol, *localAddr, *remoteAddr); err != nil {
		fmt.Printf("参数错误: %v\n\n", err)
		printUsage()
		os.Exit(1)
	}

	log.Printf("启动 %s 隧道程序 - 模式: %s", strings.ToUpper(*protocol), *mode)

	switch *protocol {
	case "udp":
		switch *mode {
		case "client":
			runClient(*localAddr, *remoteAddr)
		case "server":
			runServer(*localAddr, *remoteAddr)
		}
	case "tcp":
		switch *mode {
		case "client":
			runTCPClient(*localAddr, *remoteAddr)
		case "server":
			runTCPServer(*localAddr, *remoteAddr)
		}
	}
}
