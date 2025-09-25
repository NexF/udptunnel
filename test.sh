#!/bin/bash

# UDP 隧道测试脚本

echo "=== 编译程序 ==="
go build -o udptunnel main.go client.go server.go packet.go
go build -o test_udp_server test_udp_server.go  
go build -o test_client test_client.go

echo "=== 启动测试 UDP 服务器 (端口 12345) ==="
echo "在后台启动测试 UDP 服务器..."
./test_udp_server &
UDP_SERVER_PID=$!
sleep 1

echo "=== 启动隧道服务端 (TCP:9090 -> UDP:12345) ==="
echo "在后台启动隧道服务端..."
./udptunnel -mode=server -local=:9090 -remote=127.0.0.1:12345 &
TUNNEL_SERVER_PID=$!
sleep 1

echo "=== 启动隧道客户端 (UDP:8080 -> TCP:9090) ==="
echo "在后台启动隧道客户端..."
./udptunnel -mode=client -local=:8080 -remote=127.0.0.1:9090 &
TUNNEL_CLIENT_PID=$!
sleep 2

echo "=== 运行测试客户端 ==="
echo "发送测试消息通过隧道..."
./test_client

echo "=== 清理进程 ==="
echo "停止所有测试进程..."
kill $UDP_SERVER_PID 2>/dev/null
kill $TUNNEL_SERVER_PID 2>/dev/null  
kill $TUNNEL_CLIENT_PID 2>/dev/null

echo "=== 测试完成 ==="
echo ""
echo "测试流程说明："
echo "1. 测试客户端 -> UDP:8080 (隧道客户端)"
echo "2. 隧道客户端 -> TCP:9090 (隧道服务端)"  
echo "3. 隧道服务端 -> UDP:12345 (测试 UDP 服务器)"
echo "4. 响应按相反路径返回"
