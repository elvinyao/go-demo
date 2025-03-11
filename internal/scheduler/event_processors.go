package scheduler

import (
	"fmt"
	"log"
	"my-scheduler-go/internal/mattermost"
	"my-scheduler-go/internal/models"
	"strings"
	"time"
)

// PostedMessageProcessor 处理新发布的消息
type PostedMessageProcessor struct {
	// 配置选项
	KeywordTriggers []string
}

// NewPostedMessageProcessor 创建消息处理器
func NewPostedMessageProcessor(triggers []string) *PostedMessageProcessor {
	return &PostedMessageProcessor{
		KeywordTriggers: triggers,
	}
}

// ShouldProcess 判断是否处理事件
func (p *PostedMessageProcessor) ShouldProcess(event *mattermost.Event) bool {
	// 只处理消息发布事件
	if event.Type != mattermost.EventTypePosted {
		return false
	}

	// 必须有消息内容
	if event.Post == nil || event.Post.Message == "" {
		return false
	}

	// 检查是否包含触发关键词
	if len(p.KeywordTriggers) > 0 {
		message := strings.ToLower(event.Post.Message)
		for _, keyword := range p.KeywordTriggers {
			if strings.Contains(message, strings.ToLower(keyword)) {
				return true
			}
		}
		return false
	}

	// 没有关键词过滤时处理所有消息
	return true
}

// ProcessEvent 处理事件
func (p *PostedMessageProcessor) ProcessEvent(event *mattermost.Event) (*models.Task, error) {
	if event.Post == nil {
		return nil, fmt.Errorf("event has no post data")
	}

	log.Printf("[PostedMessageProcessor] Processing message: %s", event.Post.Message)

	// 提取消息中的关键信息
	message := event.Post.Message
	channelID := event.Post.ChannelID
	userID := event.Post.UserID

	// 分析消息内容，查找任务相关信息
	// 这里可以添加特定的业务逻辑，根据消息内容创建不同类型的任务
	taskType := determineTaskType(message)
	priority := determinePriority(message)

	// 创建任务参数
	params := map[string]interface{}{
		"event_type":      string(event.Type),
		"channel_id":      channelID,
		"message":         message,
		"user_id":         userID,
		"original_post":   event.Post,
		"processing_time": time.Now().Format(time.RFC3339),
	}

	// 添加频道信息
	if event.Channel != nil {
		params["channel_name"] = event.Channel.Name
	}

	// 添加用户信息
	if event.User != nil {
		params["username"] = event.User.Username
	}

	// 创建任务
	task := &models.Task{
		Name:       fmt.Sprintf("处理消息: %s", truncateString(message, 30)),
		TaskType:   models.TaskType(taskType),
		Status:     models.StatusPending,
		Priority:   models.TaskPriority(priority),
		Tags:       []string{"MATTERMOST", "MESSAGE"},
		Parameters: params,
		CreatedAt:  time.Now(),
	}

	return task, nil
}

// UserAddedProcessor 处理用户添加事件
type UserAddedProcessor struct{}

// NewUserAddedProcessor 创建用户添加事件处理器
func NewUserAddedProcessor() *UserAddedProcessor {
	return &UserAddedProcessor{}
}

// ShouldProcess 判断是否处理事件
func (p *UserAddedProcessor) ShouldProcess(event *mattermost.Event) bool {
	return event.Type == mattermost.EventTypeUserAdded
}

// ProcessEvent 处理事件
func (p *UserAddedProcessor) ProcessEvent(event *mattermost.Event) (*models.Task, error) {
	log.Printf("[UserAddedProcessor] Processing user added event")

	// 从事件中提取频道和用户信息
	data := event.Data
	channelID, _ := data["channel_id"].(string)
	userID, _ := data["user_id"].(string)

	params := map[string]interface{}{
		"event_type": string(event.Type),
		"channel_id": channelID,
		"user_id":    userID,
	}

	// 创建任务
	task := &models.Task{
		Name:       fmt.Sprintf("处理用户添加事件"),
		TaskType:   models.TypeImmediate,
		Status:     models.StatusPending,
		Priority:   models.PriorityLow,
		Tags:       []string{"MATTERMOST", "USER_ADDED"},
		Parameters: params,
		CreatedAt:  time.Now(),
	}

	return task, nil
}

// 辅助函数

// determineTaskType 根据消息决定任务类型
func determineTaskType(message string) string {
	// 查找调度标记
	if strings.Contains(strings.ToLower(message), "schedule:") ||
		strings.Contains(strings.ToLower(message), "cron:") {
		return string(models.TypeScheduled)
	}

	// 默认为即时任务
	return string(models.TypeImmediate)
}

// determinePriority 根据消息决定优先级
func determinePriority(message string) string {
	lowerMsg := strings.ToLower(message)

	// 紧急关键词
	if strings.Contains(lowerMsg, "urgent") ||
		strings.Contains(lowerMsg, "emergency") ||
		strings.Contains(lowerMsg, "asap") {
		return string(models.PriorityHigh)
	}

	// 低优先级关键词
	if strings.Contains(lowerMsg, "low priority") ||
		strings.Contains(lowerMsg, "when possible") {
		return string(models.PriorityLow)
	}

	// 默认为中优先级
	return string(models.PriorityMedium)
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen] + "..."
}
