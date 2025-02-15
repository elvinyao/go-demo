package scheduler

import (
	"my-scheduler-go/internal/logger"
	"my-scheduler-go/internal/mattermost"
	"my-scheduler-go/internal/models"
	"my-scheduler-go/internal/repository"
	"time"
)

// MattermostEventSource 示例, 将Mattermost事件与Task系统对接
type MattermostEventSource struct {
	repo     repository.TaskRepository
	listener *mattermost.EventListener
}

func NewMattermostEventSource(repo repository.TaskRepository, listener *mattermost.EventListener) *MattermostEventSource {
	return &MattermostEventSource{
		repo:     repo,
		listener: listener,
	}
}

// OnEvent 供事件监听器回调时调用, 创建新的任务
func (s *MattermostEventSource) OnEvent(event models.Event) {
	// 1. 事件类型过滤，只处理 "posted" (示例)；其它类型丢弃
	allowedType := "posted"
	if event.EventType != allowedType {
		logger.L.Info("[MattermostEventSource] Discard event. event_type=%s (allowed=%s)", event.EventType, allowedType)
		return
	}

	// 2. Channel 过滤: 如果 config 里指定了 channel_id，则只处理匹配的channel
	//   假设我们在此处可以访问 "s.channelID" 或通过某种方式拿到 config
	//   如果你的设计是通过构造函数或字段注入 config 里的 channelID，示例:
	configChannel := "" // 你需提前在 struct 或构造函数中注入
	if configChannel != "" && event.ChannelID != configChannel {
		logger.L.Info("[MattermostEventSource] Discard event. channelID=%s not match configChannel=%s", event.ChannelID, configChannel)
		return
	}

	logger.L.Info("[MattermostEventSource] Received valid event: %+v", event)

	// 简单示例: 看到某关键词就添加Task
	if event.Message == "trigger-task" {
		newTask := &models.Task{
			Name:   "MattermostTriggeredTask",
			Status: models.StatusPending,
			// 不带Cron => 立即执行
			CreatedAt: time.Now(),
		}
		_ = s.repo.AddTask(newTask)
		logger.L.Info("[MattermostEventSource] Created new task from event, ID pending.")
	}
}
