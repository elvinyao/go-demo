package mattermost

import (
	"log"
	"sync"
	"time"
)

// EventFilter 定义事件过滤接口
type EventFilter interface {
	// 判断事件是否应该被处理
	ShouldProcess(event *Event) bool
}

// ChannelFilter 基于频道ID的过滤器
type ChannelFilter struct {
	ChannelIDs []string
}

// ShouldProcess 判断事件是否来自关注的频道
func (f *ChannelFilter) ShouldProcess(event *Event) bool {
	if event.Channel == nil {
		return false
	}

	for _, id := range f.ChannelIDs {
		if event.Channel.ID == id {
			return true
		}
	}

	return false
}

// EventTypeFilter 基于事件类型的过滤器
type EventTypeFilter struct {
	EventTypes []EventType
}

// ShouldProcess 判断事件类型是否需要处理
func (f *EventTypeFilter) ShouldProcess(event *Event) bool {
	for _, et := range f.EventTypes {
		if event.Type == et {
			return true
		}
	}

	return false
}

// EventListener 负责监听Mattermost WebSocket事件
type EventListener struct {
	conn       *Connection
	handlers   []EventHandler
	filters    []EventFilter
	mu         sync.Mutex
	isRunning  bool
	stopChan   chan struct{}
	mockEvents bool // 是否使用模拟事件
}

// NewEventListener 创建一个新的事件监听器
func NewEventListener(conn *Connection, mockEvents bool) *EventListener {
	return &EventListener{
		conn:       conn,
		handlers:   make([]EventHandler, 0),
		filters:    make([]EventFilter, 0),
		isRunning:  false,
		stopChan:   make(chan struct{}),
		mockEvents: mockEvents,
	}
}

// AddHandler 添加事件处理器
func (l *EventListener) AddHandler(handler EventHandler) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.handlers = append(l.handlers, handler)
}

// AddFilter 添加事件过滤器
func (l *EventListener) AddFilter(filter EventFilter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.filters = append(l.filters, filter)
}

// StartListening 开始监听事件
func (l *EventListener) StartListening() {
	l.mu.Lock()
	if l.isRunning {
		l.mu.Unlock()
		return
	}
	l.isRunning = true
	l.mu.Unlock()

	log.Println("[MattermostEventListener] Start listening for events...")

	// 连接到Mattermost
	if err := l.conn.Connect(); err != nil {
		log.Printf("[MattermostEventListener] Failed to connect: %v", err)
		return
	}

	// 注册为事件处理器
	l.conn.AddEventHandler(l)

	// 如果启用了模拟事件，则启动模拟事件生成
	if l.mockEvents {
		go l.generateMockEvents()
	}
}

// HandleEvent 实现EventHandler接口
func (l *EventListener) HandleEvent(event *Event) {
	// 应用所有过滤器
	if !l.shouldProcessEvent(event) {
		return
	}

	// 分发事件到所有已注册的处理器
	l.mu.Lock()
	handlers := make([]EventHandler, len(l.handlers))
	copy(handlers, l.handlers)
	l.mu.Unlock()

	for _, handler := range handlers {
		handler.HandleEvent(event)
	}
}

// shouldProcessEvent 判断是否处理事件
func (l *EventListener) shouldProcessEvent(event *Event) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.filters) == 0 {
		return true // 无过滤器，处理所有事件
	}

	for _, filter := range l.filters {
		if filter.ShouldProcess(event) {
			return true
		}
	}

	return false
}

// StopListening 停止监听事件
func (l *EventListener) StopListening() {
	l.mu.Lock()
	if !l.isRunning {
		l.mu.Unlock()
		return
	}
	l.isRunning = false
	close(l.stopChan)
	l.mu.Unlock()

	log.Println("[MattermostEventListener] Stopping event listening.")
	l.conn.Close()
}

// generateMockEvents 生成模拟事件进行测试
func (l *EventListener) generateMockEvents() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 生成一个模拟的消息发布事件
			event := l.createMockPostEvent()
			log.Printf("[MattermostEventListener] Generated mock event: %s", event.Type)

			// 分发事件
			l.conn.DispatchEvent(event)

		case <-l.stopChan:
			return
		}
	}
}

// createMockPostEvent 创建一个模拟的消息发布事件
func (l *EventListener) createMockPostEvent() *Event {
	channelID := "channel1"
	userID := "user1"

	postData := map[string]interface{}{
		"id":         "post1",
		"create_at":  time.Now().Unix(),
		"user_id":    userID,
		"channel_id": channelID,
		"message":    "This is a test message that should trigger a task",
		"type":       "",
	}

	data := map[string]interface{}{
		"post": postData,
		"channel": map[string]interface{}{
			"id":           channelID,
			"name":         "test-channel",
			"display_name": "Test Channel",
		},
		"user": map[string]interface{}{
			"id":       userID,
			"username": "testuser",
			"email":    "test@example.com",
		},
	}

	return NewEvent(EventTypePosted, data)
}
