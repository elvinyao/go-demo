package mattermost

import "time"

// EventType 定义Mattermost事件类型
type EventType string

// 定义常见的事件类型常量
const (
	EventTypePosted         EventType = "posted"          // 新消息发布
	EventTypePostEdited     EventType = "post_edited"     // 消息编辑
	EventTypePostDeleted    EventType = "post_deleted"    // 消息删除
	EventTypeChannelCreated EventType = "channel_created" // 频道创建
	EventTypeUserAdded      EventType = "user_added"      // 用户添加到频道
	EventTypeUserRemoved    EventType = "user_removed"    // 用户从频道移除
	EventTypeTyping         EventType = "typing"          // 用户正在输入
	EventTypeDirectAdded    EventType = "direct_added"    // 添加直接消息
	EventTypeLeaveTeam      EventType = "leave_team"      // 用户离开团队
	EventTypeUpdateTeam     EventType = "update_team"     // 团队更新
)

// Post 表示一个Mattermost消息结构
type Post struct {
	ID        string                 `json:"id"`
	CreateAt  int64                  `json:"create_at"`
	UpdateAt  int64                  `json:"update_at"`
	DeleteAt  int64                  `json:"delete_at"`
	UserID    string                 `json:"user_id"`
	ChannelID string                 `json:"channel_id"`
	RootID    string                 `json:"root_id"`
	ParentID  string                 `json:"parent_id"`
	Message   string                 `json:"message"`
	Type      string                 `json:"type"`
	Props     map[string]interface{} `json:"props"`
	Hashtags  string                 `json:"hashtags"`
	FileIDs   []string               `json:"file_ids"`
}

// User 表示一个Mattermost用户
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

// Channel 表示一个Mattermost频道
type Channel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	TeamID      string `json:"team_id"`
}

// Event 表示从Mattermost WebSocket接收到的事件
type Event struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Post      *Post                  `json:"post"`
	Channel   *Channel               `json:"channel"`
	User      *User                  `json:"user"`
	Raw       map[string]interface{} `json:"raw"`
}

// NewEvent 创建一个新的事件
func NewEvent(eventType EventType, data map[string]interface{}) *Event {
	event := &Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
		Raw:       data,
	}

	// 解析Post
	if postData, ok := data["post"].(map[string]interface{}); ok {
		event.Post = &Post{
			ID:        postData["id"].(string),
			Message:   postData["message"].(string),
			ChannelID: postData["channel_id"].(string),
			UserID:    postData["user_id"].(string),
		}
	}

	// 解析Channel
	if channelData, ok := data["channel"].(map[string]interface{}); ok {
		event.Channel = &Channel{
			ID:   channelData["id"].(string),
			Name: channelData["name"].(string),
		}
	}

	return event
}
