package models

// 示例: 定义一些与Mattermost事件相关的数据结构
type Event struct {
	EventType string `json:"event_type"`
	ChannelID string `json:"channel_id"`
	Message   string `json:"message"`
}

// 也可定义更多字段, 以满足解析event需要
