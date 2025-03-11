package service

import (
	"log"
	"my-scheduler-go/internal/config"
	"my-scheduler-go/internal/mattermost"
	"time"
)

// MattermostService 处理与Mattermost交互
type MattermostService struct {
	appConfig *config.AppConfig
	conn      *mattermost.Connection
}

// NewMattermostService 创建新的Mattermost服务
func NewMattermostService(appConfig *config.AppConfig) *MattermostService {
	// 创建连接对象
	reconnectInterval := time.Duration(appConfig.Mattermost.ReconnectInterval)
	conn := mattermost.NewConnection(appConfig.Mattermost.ServerURL, appConfig.Mattermost.Token, reconnectInterval)

	return &MattermostService{
		appConfig: appConfig,
		conn:      conn,
	}
}

// Connect 连接到Mattermost服务器
func (s *MattermostService) Connect() error {
	log.Printf("[MattermostService] Connecting to %s", s.appConfig.Mattermost.ServerURL)
	return s.conn.Connect()
}

// Disconnect 断开连接
func (s *MattermostService) Disconnect() error {
	log.Println("[MattermostService] Disconnecting")
	return s.conn.Close()
}

// SendDirectMessage 发送直接消息给用户
func (s *MattermostService) SendDirectMessage(userID, message string) error {
	log.Printf("[MattermostService] Sending direct message to user %s", userID)

	// 在真实实现中，这里会调用Mattermost API发送消息
	// 现在我们只记录日志
	log.Printf("[MattermostService] Direct message sent to user %s (simulation)", userID)

	return nil
}

// SendChannelMessage 发送消息到频道
func (s *MattermostService) SendChannelMessage(channelID, message string) error {
	log.Printf("[MattermostService] Sending message to channel %s", channelID)

	// 在真实实现中，这里会调用Mattermost API发送消息
	// 现在我们只记录日志
	log.Printf("[MattermostService] Message sent to channel %s (simulation)", channelID)

	return nil
}

// CreateEventListener 创建事件监听器
func (s *MattermostService) CreateEventListener(useMockEvents bool) *mattermost.EventListener {
	return mattermost.NewEventListener(s.conn, useMockEvents)
}

// AddChannelFilter 为事件监听器添加频道过滤器
func (s *MattermostService) AddChannelFilter(listener *mattermost.EventListener, channelIDs []string) {
	filter := &mattermost.ChannelFilter{
		ChannelIDs: channelIDs,
	}
	listener.AddFilter(filter)
}

// AddEventTypeFilter 为事件监听器添加事件类型过滤器
func (s *MattermostService) AddEventTypeFilter(listener *mattermost.EventListener, eventTypes []mattermost.EventType) {
	filter := &mattermost.EventTypeFilter{
		EventTypes: eventTypes,
	}
	listener.AddFilter(filter)
}

// SendTaskReport 发送任务报告到配置的频道
func (s *MattermostService) SendTaskReport(report string) error {
	channelID := s.appConfig.Mattermost.ChannelID
	log.Printf("[MattermostService] Sending task report to channel %s", channelID)

	message := "## 任务执行报告\n\n" + report

	return s.SendChannelMessage(channelID, message)
}
