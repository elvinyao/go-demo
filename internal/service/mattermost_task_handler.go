package service

import (
	"fmt"
	"log"
	"my-scheduler-go/internal/config"
	"my-scheduler-go/internal/models"
)

// MattermostTaskHandler 处理Mattermost相关的任务
type MattermostTaskHandler struct {
	mmService *MattermostService
	appConfig *config.AppConfig
}

// NewMattermostTaskHandler 创建Mattermost任务处理器
func NewMattermostTaskHandler(mmService *MattermostService, appConfig *config.AppConfig) *MattermostTaskHandler {
	return &MattermostTaskHandler{
		mmService: mmService,
		appConfig: appConfig,
	}
}

// HandleTask 处理任务
func (h *MattermostTaskHandler) HandleTask(task *models.Task) error {
	log.Printf("[MattermostTaskHandler] Handling task: %s", task.Name)

	// 检查任务标签
	isMattermostTask := false
	for _, tag := range task.Tags {
		if tag == "MATTERMOST" || tag == "MATTERMOST_EVENT" {
			isMattermostTask = true
			break
		}
	}

	if !isMattermostTask {
		return fmt.Errorf("task is not a Mattermost task")
	}

	// 获取任务参数
	params := task.Parameters

	// 根据ForwardType决定如何处理
	forwardType, _ := params["forward_type"].(string)

	switch forwardType {
	case "direct_message":
		return h.handleDirectMessage(task)
	case "channel_message":
		return h.handleChannelMessage(task)
	case "notification":
		return h.handleNotification(task)
	default:
		return fmt.Errorf("unknown forward type: %s", forwardType)
	}
}

// 处理直接消息
func (h *MattermostTaskHandler) handleDirectMessage(task *models.Task) error {
	params := task.Parameters

	// 获取目标用户ID
	targetUserID, ok := params["target_user_id"].(string)
	if !ok {
		// 尝试从Custom配置获取
		if customMap, ok := params["custom"].(map[string]interface{}); ok {
			targetUserID, _ = customMap["target_user_id"].(string)
		}
	}

	if targetUserID == "" {
		return fmt.Errorf("no target user ID found for direct message")
	}

	// 获取消息内容
	var message string
	if originalMessage, ok := params["message"].(string); ok {
		message = fmt.Sprintf("转发消息: %s", originalMessage)
	} else {
		message = "收到了一条新通知"
	}

	// 添加消息元数据
	if channelName, ok := params["channel_name"].(string); ok {
		message += fmt.Sprintf("\n\n来源频道: %s", channelName)
	}

	if username, ok := params["username"].(string); ok {
		message += fmt.Sprintf("\n原始发送者: %s", username)
	}

	// 发送直接消息
	err := h.mmService.SendDirectMessage(targetUserID, message)
	if err != nil {
		return fmt.Errorf("failed to send direct message: %v", err)
	}

	log.Printf("[MattermostTaskHandler] Sent direct message to user %s", targetUserID)
	return nil
}

// 处理频道消息
func (h *MattermostTaskHandler) handleChannelMessage(task *models.Task) error {
	params := task.Parameters

	// 获取目标频道ID
	targetChannelID, ok := params["target_channel_id"].(string)
	if !ok {
		// 尝试从Custom配置获取
		if customMap, ok := params["custom"].(map[string]interface{}); ok {
			targetChannelID, _ = customMap["target_channel_id"].(string)
		}
	}

	if targetChannelID == "" {
		return fmt.Errorf("no target channel ID found for channel message")
	}

	// 获取消息内容
	var message string
	if originalMessage, ok := params["message"].(string); ok {
		message = fmt.Sprintf("转发消息: %s", originalMessage)
	} else {
		message = "收到了一条需要处理的新通知"
	}

	// 添加消息元数据
	if channelName, ok := params["channel_name"].(string); ok {
		message += fmt.Sprintf("\n\n来源频道: %s", channelName)
	}

	if username, ok := params["username"].(string); ok {
		message += fmt.Sprintf("\n原始发送者: %s", username)
	}

	// 发送频道消息
	err := h.mmService.SendChannelMessage(targetChannelID, message)
	if err != nil {
		return fmt.Errorf("failed to send channel message: %v", err)
	}

	log.Printf("[MattermostTaskHandler] Sent message to channel %s", targetChannelID)
	return nil
}

// 处理通知
func (h *MattermostTaskHandler) handleNotification(task *models.Task) error {
	params := task.Parameters

	// 判断是否需要通知管理员
	notifyAdmin := false
	if val, ok := params["notify_admin"].(string); ok && val == "true" {
		notifyAdmin = true
	} else if customMap, ok := params["custom"].(map[string]interface{}); ok {
		if val, ok := customMap["notify_admin"].(string); ok && val == "true" {
			notifyAdmin = true
		}
	}

	// 创建通知消息
	message := "系统通知: "
	if eventType, ok := params["event_type"].(string); ok {
		message += fmt.Sprintf("收到 %s 类型的事件", eventType)
	} else {
		message += "收到了一个系统事件"
	}

	if userID, ok := params["user_id"].(string); ok {
		message += fmt.Sprintf("\n涉及用户: %s", userID)
	}

	if channelID, ok := params["channel_id"].(string); ok {
		message += fmt.Sprintf("\n涉及频道: %s", channelID)
	}

	// 发送通知
	if notifyAdmin {
		// 发送给管理员
		adminChannelID := h.appConfig.Mattermost.ChannelID // 使用配置的管理员频道
		err := h.mmService.SendChannelMessage(adminChannelID, message)
		if err != nil {
			return fmt.Errorf("failed to send admin notification: %v", err)
		}
		log.Printf("[MattermostTaskHandler] Sent admin notification to channel %s", adminChannelID)
	} else {
		// 发送给默认频道
		defaultChannelID := h.appConfig.Mattermost.ChannelID
		err := h.mmService.SendChannelMessage(defaultChannelID, message)
		if err != nil {
			return fmt.Errorf("failed to send notification: %v", err)
		}
		log.Printf("[MattermostTaskHandler] Sent notification to channel %s", defaultChannelID)
	}

	return nil
}
