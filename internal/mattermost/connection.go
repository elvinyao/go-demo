package mattermost

import "log"

type Connection struct {
	ServerURL string
	Token     string
}

func NewConnection(serverURL, token string) *Connection {
	return &Connection{
		ServerURL: serverURL,
		Token:     token,
	}
}

func (c *Connection) Connect() error {
	// 示例: 仅打印log
	log.Printf("[MattermostConnection] Connecting to %s with token %s", c.ServerURL, c.Token)
	// 实际中使用websocket.Dial等
	return nil
}

func (c *Connection) Close() error {
	log.Println("[MattermostConnection] Closing connection.")
	return nil
}
