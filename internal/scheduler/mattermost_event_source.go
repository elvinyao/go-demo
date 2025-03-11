package scheduler

import (
	"log"
	"my-scheduler-go/internal/mattermost"
	"my-scheduler-go/internal/models"
	"my-scheduler-go/internal/repository"
	"sync"
	"time"
)

// MattermostEventSource 将Mattermost事件转换为任务
type MattermostEventSource struct {
	repo           repository.TaskRepository
	listener       *mattermost.EventListener
	configService  *ConfigurationService // 配置管理服务
	processorMutex sync.Mutex
	processors     map[string]EventProcessor // 根据事件类型或其他条件路由到不同处理器
}

// EventProcessor 定义了不同类型事件的处理逻辑
type EventProcessor interface {
	ProcessEvent(event *mattermost.Event) (*models.Task, error)
	ShouldProcess(event *mattermost.Event) bool
}

// NewMattermostEventSource 创建新的事件源
func NewMattermostEventSource(repo repository.TaskRepository, listener *mattermost.EventListener, configService *ConfigurationService) *MattermostEventSource {
	source := &MattermostEventSource{
		repo:          repo,
		listener:      listener,
		configService: configService,
		processors:    make(map[string]EventProcessor),
	}

	// 注册为事件处理器
	listener.AddHandler(source)

	return source
}

// RegisterProcessor 注册事件处理器
func (s *MattermostEventSource) RegisterProcessor(name string, processor EventProcessor) {
	s.processorMutex.Lock()
	defer s.processorMutex.Unlock()
	s.processors[name] = processor
}

// HandleEvent 实现EventHandler接口，处理所有Mattermost事件
func (s *MattermostEventSource) HandleEvent(event *mattermost.Event) {
	log.Printf("[MattermostEventSource] Received event: %s", event.Type)

	// 获取当前配置
	configs := s.configService.GetCurrentConfigurations()
	if len(configs) == 0 {
		log.Println("[MattermostEventSource] No configurations available, skipping event")
		return
	}

	// 找到匹配的配置
	var matchedConfigs []MattermostConfig
	for _, config := range configs {
		mmConfig, ok := config.(MattermostConfig)
		if !ok {
			continue
		}

		// 频道匹配
		if event.Channel != nil && mmConfig.ChannelID == event.Channel.ID {
			// 消息类型匹配
			if string(event.Type) == mmConfig.MessageType || mmConfig.MessageType == "" {
				matchedConfigs = append(matchedConfigs, mmConfig)
			}
		}
	}

	if len(matchedConfigs) == 0 {
		log.Println("[MattermostEventSource] No matching configuration for event, skipping")
		return
	}

	// 找到可处理此事件的处理器
	s.processorMutex.Lock()
	var validProcessors []EventProcessor
	for _, processor := range s.processors {
		if processor.ShouldProcess(event) {
			validProcessors = append(validProcessors, processor)
		}
	}
	s.processorMutex.Unlock()

	// 如果没有处理器可处理此事件，使用默认处理
	if len(validProcessors) == 0 {
		log.Println("[MattermostEventSource] No processor found for event, using default")
		task := s.createDefaultTask(event, matchedConfigs[0])
		if task != nil {
			if err := s.repo.AddTask(task); err != nil {
				log.Printf("[MattermostEventSource] Failed to add task: %v", err)
			} else {
				log.Printf("[MattermostEventSource] Created new task ID: %s", task.ID)
			}
		}
		return
	}

	// 使用第一个匹配的处理器处理事件
	processor := validProcessors[0]
	task, err := processor.ProcessEvent(event)
	if err != nil {
		log.Printf("[MattermostEventSource] Failed to process event: %v", err)
		return
	}

	if task != nil {
		if err := s.repo.AddTask(task); err != nil {
			log.Printf("[MattermostEventSource] Failed to add task: %v", err)
		} else {
			log.Printf("[MattermostEventSource] Created new task ID: %s", task.ID)
		}
	}
}

// createDefaultTask 创建默认任务
func (s *MattermostEventSource) createDefaultTask(event *mattermost.Event, config MattermostConfig) *models.Task {
	if event.Post == nil {
		return nil
	}

	taskName := "Mattermost事件处理"
	if event.Channel != nil {
		taskName = "处理来自 " + event.Channel.Name + " 的消息"
	}

	// 创建任务参数
	params := map[string]interface{}{
		"event_type":    string(event.Type),
		"channel_id":    event.Post.ChannelID,
		"message":       event.Post.Message,
		"user_id":       event.Post.UserID,
		"forward_type":  config.ForwardType,
		"config_id":     config.ID,
		"original_post": event.Post,
	}

	// 创建即时任务
	task := &models.Task{
		Name:       taskName,
		TaskType:   models.TypeImmediate,
		Status:     models.StatusPending,
		Priority:   models.PriorityMedium,
		Tags:       []string{"MATTERMOST_EVENT"},
		Parameters: params,
		CreatedAt:  time.Now(),
	}

	return task
}

// MattermostConfig 表示Mattermost相关的配置
type MattermostConfig struct {
	ID          string                 `json:"id"`
	ChannelID   string                 `json:"channel_id"`
	MessageType string                 `json:"message_type"`
	ForwardType string                 `json:"forward_type"`
	Custom      map[string]interface{} `json:"custom"`
}

// Start 启动事件源监听
func (s *MattermostEventSource) Start() {
	log.Println("[MattermostEventSource] Starting event source")
	s.listener.StartListening()
}

// Stop 停止事件源监听
func (s *MattermostEventSource) Stop() {
	log.Println("[MattermostEventSource] Stopping event source")
	s.listener.StopListening()
}
