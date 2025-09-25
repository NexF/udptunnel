package main

import (
	"fmt"
	"log"
	"net"
	"sync"
)

// ===============================
// 客户端模块
// ===============================

// TunnelClient UDP 隧道客户端
type TunnelClient struct {
	localUDP    string
	remoteTCP   string
	udpConn     *net.UDPConn
	connections map[string]*ClientConnection
	mu          sync.RWMutex
}

// NewTunnelClient 创建新的隧道客户端
func NewTunnelClient(localUDP, remoteTCP string) *TunnelClient {
	return &TunnelClient{
		localUDP:    localUDP,
		remoteTCP:   remoteTCP,
		connections: make(map[string]*ClientConnection),
	}
}

// Start 启动客户端
func (c *TunnelClient) Start() error {
	log.Printf("启动客户端模式 - 本地 UDP: %s, 远程 TCP: %s", c.localUDP, c.remoteTCP)

	// 监听本地 UDP
	udpAddr, err := net.ResolveUDPAddr("udp", c.localUDP)
	if err != nil {
		return fmt.Errorf("解析 UDP 地址失败: %w", err)
	}

	c.udpConn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("监听 UDP 失败: %w", err)
	}
	defer c.udpConn.Close()

	log.Printf("UDP 隧道客户端已启动，监听地址: %s", c.localUDP)

	// 处理 UDP 数据包
	return c.handleUDPPackets()
}

// handleUDPPackets 处理 UDP 数据包
func (c *TunnelClient) handleUDPPackets() error {
	buffer := make([]byte, maxPacketSize)
	for {
		n, clientAddr, err := c.udpConn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("读取 UDP 数据失败: %v", err)
			continue
		}

		if err := c.forwardToServer(clientAddr, buffer[:n]); err != nil {
			log.Printf("转发数据到服务端失败: %v", err)
		}
	}
}

// forwardToServer 转发数据到服务端
func (c *TunnelClient) forwardToServer(clientAddr *net.UDPAddr, data []byte) error {
	clientKey := clientAddr.String()

	c.mu.RLock()
	conn, exists := c.connections[clientKey]
	c.mu.RUnlock()

	if !exists || !c.isConnectionValid(conn) {
		// 如果连接不存在或无效，清理旧连接并创建新连接
		if exists {
			c.removeConnection(clientKey)
		}

		var err error
		conn, err = c.createClientConnection(clientAddr)
		if err != nil {
			return fmt.Errorf("创建客户端连接失败: %w", err)
		}
		log.Printf("为客户端 %s 建立了新的 TCP 连接", clientKey)
	}

	if err := conn.SendToServer(data); err != nil {
		// 清理失效连接
		c.removeConnection(clientKey)
		return err
	}

	return nil
}

// isConnectionValid 检查连接是否有效
func (c *TunnelClient) isConnectionValid(conn *ClientConnection) bool {
	if conn == nil || conn.tcpHandler == nil || conn.tcpHandler.conn == nil {
		return false
	}
	// 可以添加更多的连接健康检查逻辑
	return true
}

// createClientConnection 创建客户端连接
func (c *TunnelClient) createClientConnection(clientAddr *net.UDPAddr) (*ClientConnection, error) {
	// 使用带超时的连接
	tcpConn, err := net.DialTimeout("tcp", c.remoteTCP, tcpConnTimeout)
	if err != nil {
		return nil, fmt.Errorf("连接到服务端失败: %w", err)
	}

	conn := &ClientConnection{
		tcpHandler: NewTCPPacketHandler(tcpConn),
		udpConn:    c.udpConn,
		clientAddr: clientAddr,
		client:     c,
	}

	c.mu.Lock()
	c.connections[clientAddr.String()] = conn
	c.mu.Unlock()

	// 启动从服务端接收数据的协程
	go conn.HandleServerResponse()

	return conn, nil
}

// removeConnection 移除连接
func (c *TunnelClient) removeConnection(clientKey string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if conn, exists := c.connections[clientKey]; exists {
		delete(c.connections, clientKey)
		// 关闭TCP连接，但不再递归调用removeConnection
		conn.Close()
	}
}

// ClientConnection 客户端连接管理
type ClientConnection struct {
	tcpHandler *TCPPacketHandler
	udpConn    *net.UDPConn
	clientAddr *net.UDPAddr
	client     *TunnelClient
}

// SendToServer 发送数据到服务端
func (c *ClientConnection) SendToServer(data []byte) error {
	return c.tcpHandler.WritePacket(data)
}

// HandleServerResponse 处理服务端响应
func (c *ClientConnection) HandleServerResponse() {
	defer func() {
		// 确保连接被清理
		c.Close()
		if c.client != nil {
			c.client.removeConnection(c.clientAddr.String())
		}
	}()

	for {
		data, err := c.tcpHandler.ReadPacket()
		if err != nil {
			log.Printf("读取服务端响应失败: %v", err)
			return
		}

		if len(data) == 0 {
			continue
		}

		// 将数据发送回原始 UDP 客户端
		if err := c.sendUDPResponse(data); err != nil {
			log.Printf("发送 UDP 响应失败: %v", err)
			return
		}
	}
}

// sendUDPResponse 发送 UDP 响应
func (c *ClientConnection) sendUDPResponse(data []byte) error {
	_, err := c.udpConn.WriteToUDP(data, c.clientAddr)
	if err != nil {
		return fmt.Errorf("向 %s 发送 UDP 响应失败: %w", c.clientAddr.String(), err)
	}
	return nil
}

// Close 关闭连接
func (c *ClientConnection) Close() {
	if c.tcpHandler != nil && c.tcpHandler.conn != nil {
		c.tcpHandler.conn.Close()
	}
}

// runClient 启动客户端（保持向后兼容）
func runClient(localUDP, remoteTCP string) {
	client := NewTunnelClient(localUDP, remoteTCP)
	if err := client.Start(); err != nil {
		log.Fatalf("客户端启动失败: %v", err)
	}
}
