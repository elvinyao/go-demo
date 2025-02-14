package mattermost

import (
	"log"
	"time"
)

// EventListener 示例，用于监听websocket事件
type EventListener struct {
	conn *Connection
}

func NewEventListener(conn *Connection) *EventListener {
	return &EventListener{conn: conn}
}

func (l *EventListener) StartListening() {
	// 示例: 假装在goroutine里读取事件
	log.Println("[MattermostEventListener] Start listening for events...")
	go func() {
		for {
			// 这里仅演示log循环
			time.Sleep(5 * time.Second)
			log.Println("[MattermostEventListener] Received a mock event.")
		}
	}()
}

func (l *EventListener) StopListening() {
	log.Println("[MattermostEventListener] Stopping event listening.")
}
