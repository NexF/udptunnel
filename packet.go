package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	// 数据包最大大小
	maxPacketSize = 65536
	// UDP 读取超时时间
	udpReadTimeout = 30 * time.Second
	// TCP 连接超时时间
	tcpConnTimeout = 10 * time.Second
	// UDP 连接重试次数
	udpRetryCount = 3
	// UDP 连接重试间隔
	udpRetryInterval = 1 * time.Second
	// 数据包长度字段大小
	packetLengthSize = 2
)

// ===============================
// 数据包处理模块
// ===============================

// PacketWriter 数据包写入接口
type PacketWriter interface {
	WritePacket(data []byte) error
}

// PacketReader 数据包读取接口
type PacketReader interface {
	ReadPacket() ([]byte, error)
}

// TCPPacketHandler TCP 数据包处理器
type TCPPacketHandler struct {
	conn net.Conn
}

// NewTCPPacketHandler 创建 TCP 数据包处理器
func NewTCPPacketHandler(conn net.Conn) *TCPPacketHandler {
	return &TCPPacketHandler{conn: conn}
}

// WritePacket 写入数据包到 TCP 连接
func (h *TCPPacketHandler) WritePacket(data []byte) error {
	length := uint16(len(data))
	lengthBytes := make([]byte, packetLengthSize)
	binary.BigEndian.PutUint16(lengthBytes, length)

	if _, err := h.conn.Write(lengthBytes); err != nil {
		return fmt.Errorf("写入数据长度失败: %w", err)
	}

	if _, err := h.conn.Write(data); err != nil {
		return fmt.Errorf("写入数据内容失败: %w", err)
	}

	return nil
}

// ReadPacket 从 TCP 连接读取数据包
func (h *TCPPacketHandler) ReadPacket() ([]byte, error) {
	lengthBytes := make([]byte, packetLengthSize)
	if _, err := io.ReadFull(h.conn, lengthBytes); err != nil {
		return nil, fmt.Errorf("读取数据长度失败: %w", err)
	}

	length := binary.BigEndian.Uint16(lengthBytes)
	if length == 0 {
		return []byte{}, nil
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(h.conn, data); err != nil {
		return nil, fmt.Errorf("读取数据内容失败: %w", err)
	}

	return data, nil
}
