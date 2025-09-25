package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

// ===============================
// 服务端模块
// ===============================

// TunnelServer UDP 隧道服务端
type TunnelServer struct {
	listenTCP string
	targetUDP string
	listener  net.Listener
}

// NewTunnelServer 创建新的隧道服务端
func NewTunnelServer(listenTCP, targetUDP string) *TunnelServer {
	return &TunnelServer{
		listenTCP: listenTCP,
		targetUDP: targetUDP,
	}
}

// Start 启动服务端
func (s *TunnelServer) Start() error {
	log.Printf("启动服务端模式 - 监听 TCP: %s, 目标 UDP: %s", s.listenTCP, s.targetUDP)

	var err error
	s.listener, err = net.Listen("tcp", s.listenTCP)
	if err != nil {
		return fmt.Errorf("监听 TCP 失败: %w", err)
	}
	defer s.listener.Close()

	log.Printf("UDP 隧道服务端已启动，监听地址: %s", s.listenTCP)

	return s.acceptConnections()
}

// acceptConnections 接受客户端连接
func (s *TunnelServer) acceptConnections() error {
	for {
		tcpConn, err := s.listener.Accept()
		if err != nil {
			log.Printf("接受连接失败: %v", err)
			continue
		}

		log.Printf("接受来自 %s 的连接", tcpConn.RemoteAddr().String())

		// 为每个连接启动处理协程
		go s.handleClientConnection(tcpConn)
	}
}

// handleClientConnection 处理客户端连接（新版）
func (s *TunnelServer) handleClientConnection(tcpConn net.Conn) {
	serverConn, err := NewServerConnection(tcpConn, s.targetUDP)
	if err != nil {
		log.Printf("创建服务端连接失败: %v", err)
		tcpConn.Close()
		return
	}

	log.Printf("[客户端 %s] 连接已建立，开始处理数据", serverConn.clientAddr)
	serverConn.Start()
}

// ServerConnection 服务端连接管理
type ServerConnection struct {
	tcpHandler *TCPPacketHandler
	udpConn    *net.UDPConn
	clientAddr string
	targetUDP  string // 保存目标UDP地址
}

// NewServerConnection 创建新的服务端连接
func NewServerConnection(tcpConn net.Conn, targetUDP string) (*ServerConnection, error) {
	// 连接到目标 UDP 服务
	udpAddr, err := net.ResolveUDPAddr("udp", targetUDP)
	if err != nil {
		return nil, fmt.Errorf("解析目标 UDP 地址失败: %w", err)
	}

	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, fmt.Errorf("连接到目标 UDP 失败: %w", err)
	}

	return &ServerConnection{
		tcpHandler: NewTCPPacketHandler(tcpConn),
		udpConn:    udpConn,
		clientAddr: tcpConn.RemoteAddr().String(),
		targetUDP:  targetUDP,
	}, nil
}

// Start 启动连接处理
func (sc *ServerConnection) Start() {
	defer sc.Close()

	// 启动 UDP 响应处理协程
	go sc.handleUDPResponse()

	// 处理来自客户端的数据
	sc.handleClientData()
}

// handleUDPResponse 处理 UDP 响应
func (sc *ServerConnection) handleUDPResponse() {
	buffer := make([]byte, maxPacketSize)
	for {
		sc.udpConn.SetReadDeadline(time.Now().Add(udpReadTimeout))
		n, err := sc.udpConn.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.Printf("[客户端 %s] 读取 UDP 响应失败: %v", sc.clientAddr, err)
			return
		}

		// 发送响应回客户端
		if err := sc.tcpHandler.WritePacket(buffer[:n]); err != nil {
			log.Printf("[客户端 %s] 发送 TCP 响应失败: %v", sc.clientAddr, err)
			return
		}
	}
}

// handleClientData 处理客户端数据
func (sc *ServerConnection) handleClientData() {
	for {
		data, err := sc.tcpHandler.ReadPacket()
		if err != nil {
			log.Printf("[客户端 %s] 读取客户端数据失败: %v", sc.clientAddr, err)
			return
		}

		if len(data) == 0 {
			continue
		}

		// 转发到目标 UDP 服务
		if err := sc.forwardToUDP(data); err != nil {
			log.Printf("[客户端 %s] 转发到 UDP 失败: %v", sc.clientAddr, err)
			return
		}
	}
}

// forwardToUDP 转发数据到 UDP
func (sc *ServerConnection) forwardToUDP(data []byte) error {
	_, err := sc.udpConn.Write(data)
	if err != nil {
		// 如果UDP连接失败，尝试重新建立连接
		if sc.reconnectUDP() == nil {
			// 重试发送
			_, retryErr := sc.udpConn.Write(data)
			if retryErr == nil {
				log.Printf("[客户端 %s] UDP连接重建成功，数据发送完成", sc.clientAddr)
				return nil
			}
		}
		return fmt.Errorf("向 UDP 服务写入数据失败: %w", err)
	}
	return nil
}

// reconnectUDP 重新连接UDP（带重试机制）
func (sc *ServerConnection) reconnectUDP() error {
	if sc.udpConn != nil {
		sc.udpConn.Close()
	}

	// 使用保存的目标UDP地址重新连接
	udpAddr, err := net.ResolveUDPAddr("udp", sc.targetUDP)
	if err != nil {
		return fmt.Errorf("解析目标 UDP 地址失败: %w", err)
	}

	// 重试连接
	var lastErr error
	for i := 0; i < udpRetryCount; i++ {
		newConn, err := net.DialUDP("udp", nil, udpAddr)
		if err == nil {
			sc.udpConn = newConn
			log.Printf("[客户端 %s] UDP连接已重建到 %s (重试 %d/%d)", sc.clientAddr, sc.targetUDP, i+1, udpRetryCount)
			return nil
		}

		lastErr = err
		log.Printf("[客户端 %s] UDP重连失败 (尝试 %d/%d): %v", sc.clientAddr, i+1, udpRetryCount, err)

		if i < udpRetryCount-1 {
			time.Sleep(udpRetryInterval)
		}
	}

	return fmt.Errorf("重新连接到目标 UDP 失败（已重试 %d 次）: %w", udpRetryCount, lastErr)
}

// Close 关闭连接
func (sc *ServerConnection) Close() {
	if sc.tcpHandler != nil && sc.tcpHandler.conn != nil {
		sc.tcpHandler.conn.Close()
	}
	if sc.udpConn != nil {
		sc.udpConn.Close()
	}
	log.Printf("[客户端 %s] 连接已关闭", sc.clientAddr)
}

// runServer 启动服务端（保持向后兼容）
func runServer(listenTCP, targetUDP string) {
	server := NewTunnelServer(listenTCP, targetUDP)
	if err := server.Start(); err != nil {
		log.Fatalf("服务端启动失败: %v", err)
	}
}
