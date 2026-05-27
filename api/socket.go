package api

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"webhelper/global"

	"github.com/gorilla/websocket"
)

// WebSocketClient WebSocket 客户端
type WebSocketClient struct {
	conn        *websocket.Conn
	isConnected bool
	mu          sync.RWMutex
	onMessage   func(message string)
	onError     func(err error)
	onClose     func()
}

// WebSocketConfig WebSocket 配置
type WebSocketConfig struct {
	URL      string            `json:"url"`
	Headers  map[string]string `json:"headers"`
	Insecure bool              `json:"insecure"`
}

// WebSocketMessage WebSocket 消息
type WebSocketMessage struct {
	Type    string `json:"type"` // text, binary, ping, pong, error, system
	Content string `json:"content"`
	Time    string `json:"time"`
}

var wsClient *WebSocketClient

// NewWebSocketClient 创建 WebSocket 客户端
func NewWebSocketClient() *WebSocketClient {
	return &WebSocketClient{
		isConnected: false,
	}
}

// Connect 连接 WebSocket
func (c *WebSocketClient) Connect(config *WebSocketConfig, onMessage func(string), onError func(error), onClose func()) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isConnected {
		return fmt.Errorf("已连接")
	}

	// 解析 URL
	u, err := url.Parse(config.URL)
	if err != nil {
		return fmt.Errorf("URL 格式错误: %v", err)
	}

	// 创建 HTTP Header（只添加用户自定义头）
	header := http.Header{}
	for key, value := range config.Headers {
		header.Set(key, value)
	}

	// 创建 Dialer（最简配置）
	dialer := &websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	// 连接
	global.Log.Infof("正在连接 WebSocket: %s", u.String())
	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}

	c.conn = conn
	c.isConnected = true
	c.onMessage = onMessage
	c.onError = onError
	c.onClose = onClose

	global.Log.Infof("WebSocket 连接成功: %s", u.String())

	// 启动消息监听
	go c.readMessages()

	return nil
}

// readMessages 读取消息
func (c *WebSocketClient) readMessages() {
	defer func() {
		c.mu.Lock()
		c.isConnected = false
		c.mu.Unlock()
		if c.onClose != nil {
			c.onClose()
		}
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if c.onError != nil {
				c.onError(err)
			}
			return
		}
		if c.onMessage != nil {
			c.onMessage(string(message))
		}
	}
}

// Send 发送消息
func (c *WebSocketClient) Send(message string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isConnected || c.conn == nil {
		return fmt.Errorf("未连接")
	}

	return c.conn.WriteMessage(websocket.TextMessage, []byte(message))
}

// Close 关闭连接
func (c *WebSocketClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.isConnected = false
		return err
	}
	return nil
}

// IsConnected 是否已连接
func (c *WebSocketClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}

// GetWebSocketClient 获取 WebSocket 客户端
func GetWebSocketClient() *WebSocketClient {
	if wsClient == nil {
		wsClient = NewWebSocketClient()
	}
	return wsClient
}

// ConnectWebSocket 连接 WebSocket
func ConnectWebSocket(config *WebSocketConfig, onMessage func(string), onError func(error), onClose func()) error {
	client := GetWebSocketClient()
	return client.Connect(config, onMessage, onError, onClose)
}

// SendWebSocketMessage 发送 WebSocket 消息
func SendWebSocketMessage(message string) error {
	client := GetWebSocketClient()
	return client.Send(message)
}

// CloseWebSocket 关闭 WebSocket 连接
func CloseWebSocket() error {
	client := GetWebSocketClient()
	return client.Close()
}

// IsWebSocketConnected WebSocket 是否已连接
func IsWebSocketConnected() bool {
	client := GetWebSocketClient()
	return client.IsConnected()
}

// TCPClient TCP 客户端
type TCPClient struct {
	conn        net.Conn
	isConnected bool
	mu          sync.RWMutex
	onMessage   func(message string)
	onError     func(err error)
	onClose     func()
}

// TCPConfig TCP 配置
type TCPConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// TCPMessage TCP 消息
type TCPMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Time    string `json:"time"`
}

var tcpClient *TCPClient

// NewTCPClient 创建 TCP 客户端
func NewTCPClient() *TCPClient {
	return &TCPClient{
		isConnected: false,
	}
}

// Connect 连接 TCP
func (c *TCPClient) Connect(config *TCPConfig, onMessage func(string), onError func(error), onClose func()) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isConnected {
		return fmt.Errorf("已连接")
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}

	c.conn = conn
	c.isConnected = true
	c.onMessage = onMessage
	c.onError = onError
	c.onClose = onClose

	// 启动消息监听
	go c.readMessages()

	return nil
}

// readMessages 读取消息
func (c *TCPClient) readMessages() {
	defer func() {
		c.mu.Lock()
		c.isConnected = false
		c.mu.Unlock()
		if c.onClose != nil {
			c.onClose()
		}
	}()

	buf := make([]byte, 4096)
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			if c.onError != nil {
				c.onError(err)
			}
			return
		}
		if c.onMessage != nil {
			c.onMessage(string(buf[:n]))
		}
	}
}

// Send 发送消息
func (c *TCPClient) Send(message string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isConnected || c.conn == nil {
		return fmt.Errorf("未连接")
	}

	_, err := c.conn.Write([]byte(message))
	return err
}

// Close 关闭连接
func (c *TCPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.isConnected = false
		return err
	}
	return nil
}

// IsConnected 是否已连接
func (c *TCPClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}

// GetTCPClient 获取 TCP 客户端
func GetTCPClient() *TCPClient {
	if tcpClient == nil {
		tcpClient = NewTCPClient()
	}
	return tcpClient
}

// ConnectTCP 连接 TCP
func ConnectTCP(config *TCPConfig, onMessage func(string), onError func(error), onClose func()) error {
	client := GetTCPClient()
	return client.Connect(config, onMessage, onError, onClose)
}

// SendTCPMessage 发送 TCP 消息
func SendTCPMessage(message string) error {
	client := GetTCPClient()
	return client.Send(message)
}

// CloseTCP 关闭 TCP 连接
func CloseTCP() error {
	client := GetTCPClient()
	return client.Close()
}

// IsTCPConnected TCP 是否已连接
func IsTCPConnected() bool {
	client := GetTCPClient()
	return client.isConnected
}
