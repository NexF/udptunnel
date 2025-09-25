# UDP 隧道程序

这是一个用 Go 语言编写的 UDP 隧道程序，可以通过 TCP 连接转发 UDP 数据包。

## 功能

- **客户端模式**: 监听本地 UDP 端口，将接收到的 UDP 数据包通过 TCP 连接转发到服务端
- **服务端模式**: 监听 TCP 连接，接收客户端转发的数据，然后转发到目标 UDP 服务

## 项目结构

```
udptunnel/
├── main.go           # 主程序入口，命令行参数处理
├── client.go         # 客户端模块：TunnelClient, ClientConnection
├── server.go         # 服务端模块：TunnelServer, ServerConnection  
├── packet.go         # 数据包处理：TCPPacketHandler, 接口定义
├── go.mod           # Go 模块配置
├── README.md        # 项目文档
└── tests/           # 测试文件目录
    ├── test_client.go   # 测试客户端
    ├── test_udp_server.go # 测试 UDP 服务器
    └── test.sh          # 自动化测试脚本
```

## 编译

```bash
go build -o udptunnel main.go client.go server.go packet.go
```

推荐使用 Go 模块方式：

```bash
go build -o udptunnel .
```

或者指定文件编译：

```bash
go build -o udptunnel main.go client.go server.go packet.go
```

## 使用方法

### 客户端模式

```bash
./udptunnel -mode=client -local=:8080 -remote=server.example.com:9090
```

参数说明：
- `-local`: 本地 UDP 监听地址和端口
- `-remote`: 远程 TCP 服务端地址和端口

### 服务端模式

```bash
./udptunnel -mode=server -local=:9090 -remote=127.0.0.1:53
```

参数说明：
- `-local`: TCP 监听地址和端口
- `-remote`: 目标 UDP 服务地址和端口

## 使用场景示例

### DNS 隧道

假设您想通过 TCP 隧道访问远程的 DNS 服务器：

1. 在有 DNS 服务器访问权限的机器上运行服务端：
   ```bash
   ./udptunnel -mode=server -local=:9090 -remote=8.8.8.8:53
   ```

2. 在客户端机器上运行客户端：
   ```bash
   ./udptunnel -mode=client -local=:5353 -remote=dns-server.example.com:9090
   ```

3. 现在您可以向本地的 5353 端口发送 DNS 查询，它会通过 TCP 隧道转发到 8.8.8.8:53

## 运行测试

项目包含完整的测试套件，可以验证隧道功能：

```bash
cd tests
chmod +x test.sh
./test.sh
```

测试脚本会自动：
1. 编译主程序和测试程序
2. 启动测试 UDP 服务器（端口 12345）
3. 启动隧道服务端（TCP 9090 -> UDP 12345）
4. 启动隧道客户端（UDP 8080 -> TCP 9090）
5. 运行测试客户端发送消息验证隧道功能
6. 自动清理所有进程

## 代码架构

### 模块化设计

代码采用模块化设计，拆分为多个文件，每个文件负责特定功能：

#### 📁 main.go - 主程序入口
- 命令行参数解析和验证
- 程序启动逻辑
- 帮助信息显示

#### 📁 packet.go - 数据包处理模块
- `PacketWriter/PacketReader` 接口：统一的数据包读写接口
- `TCPPacketHandler`：TCP 数据包的封装和解封装处理
- 常量定义：包大小、超时时间等

#### 📁 client.go - 客户端模块
- `TunnelClient`：客户端主控制器，管理 UDP 监听和连接池
- `ClientConnection`：管理单个客户端连接的生命周期
- UDP 数据包接收和转发逻辑

#### 📁 server.go - 服务端模块
- `TunnelServer`：服务端主控制器，管理 TCP 监听
- `ServerConnection`：管理服务端到客户端的连接
- TCP 连接处理和 UDP 转发逻辑

### 工作原理

1. **客户端流程**：
   ```
   UDP客户端 -> TunnelClient -> ClientConnection -> TCPPacketHandler -> 服务端
   ```
   - 监听本地 UDP 端口，接收客户端数据
   - 为每个 UDP 客户端建立独立的 TCP 连接
   - 使用连接池管理多个客户端连接

2. **服务端流程**：
   ```
   客户端 -> TunnelServer -> ServerConnection -> UDP目标服务 -> 响应返回
   ```
   - 监听 TCP 端口，接受客户端连接
   - 为每个 TCP 连接建立到目标 UDP 服务的连接
   - 双向转发数据包

### 数据包格式

TCP 传输的数据包格式：
- 前 2 字节：数据长度（大端序，uint16）
- 后续字节：原始 UDP 数据

这种格式确保了 TCP 流中数据包的正确分割和重组。
