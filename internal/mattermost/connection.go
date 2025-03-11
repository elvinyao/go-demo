package mattermost

import (
	"log"
	"sync"
	"time"
)

// Connection 管理与Mattermost的连接
type Connection struct {
	ServerURL         string
	Token             string
	Connected         bool
	ReconnectInterval time.Duration
	stopChan          chan struct{}
	mu                sync.Mutex
	eventHandlers     []EventHandler
}

// EventHandler 事件处理器接口
type EventHandler interface {
	HandleEvent(event *Event)
}

// NewConnection 创建一个新的连接实例
func NewConnection(serverURL, token string, reconnectInterval time.Duration) *Connection {
	return &Connection{
		ServerURL:         serverURL,
		Token:             token,
		Connected:         false,
		ReconnectInterval: reconnectInterval,
		stopChan:          make(chan struct{}),
		eventHandlers:     make([]EventHandler, 0),
	}
}

// Connect 连接到Mattermost WebSocket
func (c *Connection) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Connected {
		log.Println("[MattermostConnection] Already connected")
		return nil
	}

	// 模拟连接过程
	log.Printf("[MattermostConnection] Connecting to %s", c.ServerURL)
	time.Sleep(500 * time.Millisecond) // 模拟连接延迟
	c.Connected = true

	// 启动心跳和重连机制
	go c.maintainConnection()

	return nil
}

// AddEventHandler 添加事件处理器
func (c *Connection) AddEventHandler(handler EventHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventHandlers = append(c.eventHandlers, handler)
}

// 维持连接的后台协程
func (c *Connection) maintainConnection() {
	ticker := time.NewTicker(c.ReconnectInterval * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !c.Connected {
				log.Println("[MattermostConnection] Connection lost, reconnecting...")
				c.Connect()
			}
		case <-c.stopChan:
			return
		}
	}
}

// DispatchEvent 分发事件到所有处理器
func (c *Connection) DispatchEvent(event *Event) {
	c.mu.Lock()
	handlers := make([]EventHandler, len(c.eventHandlers))
	copy(handlers, c.eventHandlers)
	c.mu.Unlock()

	for _, handler := range handlers {
		handler.HandleEvent(event)
	}
}

// Close 关闭连接
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.Connected {
		return nil
	}

	close(c.stopChan)
	c.Connected = false
	log.Println("[MattermostConnection] Connection closed")
	return nil
}
