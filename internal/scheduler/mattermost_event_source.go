package scheduler

import (
	"log"
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
	log.Printf("[MattermostEventSource] Received event: %+v", event)

	// 简单示例: 看到某关键词就添加Task
	if event.Message == "trigger-task" {
		newTask := &models.Task{
			Name:   "MattermostTriggeredTask",
			Status: models.StatusPending,
			// 不带Cron => 立即执行
			CreatedAt: time.Now(),
		}
		_ = s.repo.AddTask(newTask)
		log.Printf("[MattermostEventSource] Created new task from event, ID pending.")
	}
}
