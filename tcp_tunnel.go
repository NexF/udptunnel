package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

// ===============================
// TCP 隧道模块
// ===============================

// TCPTunnelClient TCP隧道客户端
type TCPTunnelClient struct {
	localTCP    string
	remoteTCP   string
	listener    net.Listener
	connections map[string]*TCPClientConnection
	mu          sync.RWMutex
}

// NewTCPTunnelClient 创建新的TCP隧道客户端
func NewTCPTunnelClient(localTCP, remoteTCP string) *TCPTunnelClient {
	return &TCPTunnelClient{
		localTCP:    localTCP,
		remoteTCP:   remoteTCP,
		connections: make(map[string]*TCPClientConnection),
	}
}

// Start 启动TCP客户端
func (c *TCPTunnelClient) Start() error {
	log.Printf("启动TCP客户端模式 - 本地 TCP: %s, 远程 TCP: %s", c.localTCP, c.remoteTCP)

	var err error
	c.listener, err = net.Listen("tcp", c.localTCP)
	if err != nil {
		return fmt.Errorf("监听本地 TCP 失败: %w", err)
	}
	defer c.listener.Close()

	log.Printf("TCP 隧道客户端已启动，监听地址: %s", c.localTCP)

	return c.acceptConnections()
}

// acceptConnections 接受客户端连接
func (c *TCPTunnelClient) acceptConnections() error {
	for {
		localConn, err := c.listener.Accept()
		if err != nil {
			log.Printf("接受本地连接失败: %v", err)
			continue
		}

		log.Printf("接受来自 %s 的本地连接", localConn.RemoteAddr().String())

		// 为每个连接启动处理协程
		go c.handleLocalConnection(localConn)
	}
}

// handleLocalConnection 处理本地连接
func (c *TCPTunnelClient) handleLocalConnection(localConn net.Conn) {
	defer localConn.Close()

	// 连接到远程服务端
	remoteConn, err := net.DialTimeout("tcp", c.remoteTCP, tcpConnTimeout)
	if err != nil {
		log.Printf("连接到远程服务端失败: %v", err)
		return
	}
	defer remoteConn.Close()

	clientKey := localConn.RemoteAddr().String()
	log.Printf("为客户端 %s 建立了到远程服务端 %s 的连接", clientKey, c.remoteTCP)

	// 创建连接管理对象
	tcpConn := &TCPClientConnection{
		localConn:  localConn,
		remoteConn: remoteConn,
		clientKey:  clientKey,
		client:     c,
	}

	// 注册连接
	c.registerConnection(clientKey, tcpConn)
	defer c.removeConnection(clientKey)

	// 启动双向数据转发
	tcpConn.startForwarding()
}

// registerConnection 注册连接
func (c *TCPTunnelClient) registerConnection(clientKey string, conn *TCPClientConnection) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connections[clientKey] = conn
}

// removeConnection 移除连接
func (c *TCPTunnelClient) removeConnection(clientKey string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.connections, clientKey)
}

// TCPClientConnection TCP客户端连接管理
type TCPClientConnection struct {
	localConn  net.Conn
	remoteConn net.Conn
	clientKey  string
	client     *TCPTunnelClient
}

// startForwarding 启动双向转发
func (c *TCPClientConnection) startForwarding() {
	var wg sync.WaitGroup
	wg.Add(2)

	// 本地到远程的转发
	go func() {
		defer wg.Done()
		c.forwardData(c.localConn, c.remoteConn, "local->remote")
	}()

	// 远程到本地的转发
	go func() {
		defer wg.Done()
		c.forwardData(c.remoteConn, c.localConn, "remote->local")
	}()

	wg.Wait()
	log.Printf("[客户端 %s] 连接已关闭", c.clientKey)
}

// forwardData 转发数据
func (c *TCPClientConnection) forwardData(src, dst net.Conn, direction string) {
	defer func() {
		// 关闭目标连接的写入，触发对方读取结束
		if tcpConn, ok := dst.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
	}()

	written, err := io.Copy(dst, src)
	if err != nil {
		log.Printf("[客户端 %s] %s 数据转发出错: %v", c.clientKey, direction, err)
	} else {
		log.Printf("[客户端 %s] %s 数据转发完成，传输 %d 字节", c.clientKey, direction, written)
	}
}

// ===============================
// TCP 隧道服务端
// ===============================

// TCPTunnelServer TCP隧道服务端
type TCPTunnelServer struct {
	listenTCP string
	targetTCP string
	listener  net.Listener
}

// NewTCPTunnelServer 创建新的TCP隧道服务端
func NewTCPTunnelServer(listenTCP, targetTCP string) *TCPTunnelServer {
	return &TCPTunnelServer{
		listenTCP: listenTCP,
		targetTCP: targetTCP,
	}
}

// Start 启动TCP服务端
func (s *TCPTunnelServer) Start() error {
	log.Printf("启动TCP服务端模式 - 监听 TCP: %s, 目标 TCP: %s", s.listenTCP, s.targetTCP)

	var err error
	s.listener, err = net.Listen("tcp", s.listenTCP)
	if err != nil {
		return fmt.Errorf("监听 TCP 失败: %w", err)
	}
	defer s.listener.Close()

	log.Printf("TCP 隧道服务端已启动，监听地址: %s", s.listenTCP)

	return s.acceptConnections()
}

// acceptConnections 接受客户端连接
func (s *TCPTunnelServer) acceptConnections() error {
	for {
		clientConn, err := s.listener.Accept()
		if err != nil {
			log.Printf("接受连接失败: %v", err)
			continue
		}

		log.Printf("接受来自 %s 的连接", clientConn.RemoteAddr().String())

		// 为每个连接启动处理协程
		go s.handleClientConnection(clientConn)
	}
}

// handleClientConnection 处理客户端连接
func (s *TCPTunnelServer) handleClientConnection(clientConn net.Conn) {
	defer clientConn.Close()

	// 连接到目标TCP服务
	targetConn, err := net.DialTimeout("tcp", s.targetTCP, tcpConnTimeout)
	if err != nil {
		log.Printf("[客户端 %s] 连接到目标TCP服务失败: %v", clientConn.RemoteAddr().String(), err)
		return
	}
	defer targetConn.Close()

	clientAddr := clientConn.RemoteAddr().String()
	log.Printf("[客户端 %s] 连接已建立，目标: %s", clientAddr, s.targetTCP)

	// 创建服务端连接管理对象
	serverConn := &TCPServerConnection{
		clientConn: clientConn,
		targetConn: targetConn,
		clientAddr: clientAddr,
		targetTCP:  s.targetTCP,
	}

	// 启动双向数据转发
	serverConn.startForwarding()
}

// TCPServerConnection TCP服务端连接管理
type TCPServerConnection struct {
	clientConn net.Conn
	targetConn net.Conn
	clientAddr string
	targetTCP  string
}

// startForwarding 启动双向转发
func (s *TCPServerConnection) startForwarding() {
	var wg sync.WaitGroup
	wg.Add(2)

	// 客户端到目标的转发
	go func() {
		defer wg.Done()
		s.forwardData(s.clientConn, s.targetConn, "client->target")
	}()

	// 目标到客户端的转发
	go func() {
		defer wg.Done()
		s.forwardData(s.targetConn, s.clientConn, "target->client")
	}()

	wg.Wait()
	log.Printf("[客户端 %s] 连接已关闭", s.clientAddr)
}

// forwardData 转发数据
func (s *TCPServerConnection) forwardData(src, dst net.Conn, direction string) {
	defer func() {
		// 关闭目标连接的写入，触发对方读取结束
		if tcpConn, ok := dst.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
	}()

	written, err := io.Copy(dst, src)
	if err != nil {
		log.Printf("[客户端 %s] %s 数据转发出错: %v", s.clientAddr, direction, err)
	} else {
		log.Printf("[客户端 %s] %s 数据转发完成，传输 %d 字节", s.clientAddr, direction, written)
	}
}

// ===============================
// 启动函数（保持向后兼容）
// ===============================

// runTCPClient 启动TCP客户端
func runTCPClient(localTCP, remoteTCP string) {
	client := NewTCPTunnelClient(localTCP, remoteTCP)
	if err := client.Start(); err != nil {
		log.Fatalf("TCP客户端启动失败: %v", err)
	}
}

// runTCPServer 启动TCP服务端
func runTCPServer(listenTCP, targetTCP string) {
	server := NewTCPTunnelServer(listenTCP, targetTCP)
	if err := server.Start(); err != nil {
		log.Fatalf("TCP服务端启动失败: %v", err)
	}
}
