package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"my-scheduler-go/internal/api"
	"my-scheduler-go/internal/config"
	"my-scheduler-go/internal/mattermost"
	"my-scheduler-go/internal/models"
	"my-scheduler-go/internal/repository"
	"my-scheduler-go/internal/scheduler"
	"my-scheduler-go/internal/service"
)

func main() {
	// 1. 加载配置
	appConfig, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 设置日志
	setupLogging(appConfig)
	log.Println("[main] Starting APScheduler Task Management System...")

	// 3. 初始化任务仓库
	repo := repository.NewInMemoryTaskRepository()
	log.Println("[main] Task repository initialized")

	// 4. 初始化任务执行器
	executor := scheduler.NewTaskExecutor(repo)
	log.Println("[main] Task executor initialized")

	// 5. 创建调度服务
	pollInterval := time.Duration(appConfig.Scheduler.PollInterval) * time.Second
	schedService := scheduler.NewSchedulerService(repo, executor, pollInterval)

	// 设置最大并发度
	schedService.SetMaxConcurrency(appConfig.Scheduler.Concurrency)

	// 6. 初始化Mattermost服务
	mattermostService := service.NewMattermostService(appConfig)
	log.Println("[main] Mattermost service initialized")

	// 7. 初始化Confluence服务
	confluenceService := service.NewConfluenceService(appConfig)
	log.Println("[main] Confluence service initialized")

	// 8. 创建配置获取器
	useMockData := appConfig.Environment == "development"
	configFetcher := service.NewConfluenceConfigFetcher(confluenceService, appConfig, useMockData)
	log.Println("[main] Configuration fetcher initialized")

	// 9. 创建配置服务 (每180秒更新一次配置)
	configService := scheduler.NewConfigurationService(configFetcher, 180*time.Second)
	log.Println("[main] Configuration service initialized")

	// 10. 创建Mattermost事件监听器
	eventListener := mattermostService.CreateEventListener(useMockData)
	log.Println("[main] Mattermost event listener created")

	// 11. 添加事件过滤器
	mattermostService.AddChannelFilter(eventListener, []string{appConfig.Mattermost.ChannelID})
	mattermostService.AddEventTypeFilter(eventListener, []mattermost.EventType{
		mattermost.EventTypePosted,
		mattermost.EventTypeUserAdded,
	})
	log.Println("[main] Event filters configured")

	// 12. 创建Mattermost事件源
	eventSource := scheduler.NewMattermostEventSource(repo, eventListener, configService)
	log.Println("[main] Mattermost event source created")

	// 13. 注册事件处理器
	eventSource.RegisterProcessor("posted_messages", scheduler.NewPostedMessageProcessor([]string{
		"task", "schedule", "urgent", "important",
	}))
	eventSource.RegisterProcessor("user_added", scheduler.NewUserAddedProcessor())
	log.Println("[main] Event processors registered")

	// 14. 创建Mattermost任务处理器，但目前暂不使用
	// 在完整实现中，这里会注册任务处理器到执行器
	_ = service.NewMattermostTaskHandler(mattermostService, appConfig)
	log.Println("[main] Mattermost task handler created")

	// 15. 任务处理器配置
	log.Println("[main] Task handlers configured")

	// 16. 启动各服务
	configService.Start()
	log.Println("[main] Configuration service started")

	eventSource.Start()
	log.Println("[main] Mattermost event source started")

	// 启动调度服务
	schedService.Start()
	log.Println("[main] Scheduler service started")

	// 17. 初始化并启动结果报告服务
	reportingService := service.NewResultReportingService(repo, appConfig)
	reportingService.Start()
	log.Println("[main] Result reporting service started")

	// 18. 开发模式下创建示例任务
	if appConfig.Environment == "development" {
		createExampleTasks(schedService)
	}

	// 19. 设置HTTP服务器和API路由
	router := api.SetupRouter(repo, schedService, reportingService)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    ":8000",
		Handler: router,
	}

	// 20. 在独立的goroutine中启动HTTP服务器
	go func() {
		log.Printf("[main] HTTP server listening on %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 21. 设置优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[main] Shutdown signal received, stopping services...")

	// 22. 优雅关闭服务
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 停止事件源
	eventSource.Stop()
	log.Println("[main] Mattermost event source stopped")

	// 停止配置服务
	configService.Stop()
	log.Println("[main] Configuration service stopped")

	// 停止报告服务
	reportingService.Stop()
	log.Println("[main] Result reporting service stopped")

	// 停止调度器
	schedService.Stop()
	log.Println("[main] Scheduler service stopped")

	// 关闭HTTP服务器
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
	log.Println("[main] HTTP server stopped")

	log.Println("[main] APScheduler Task Management System shutdown complete")
}

// setupLogging 配置应用日志
func setupLogging(appConfig *config.AppConfig) {
	// 这个例子使用标准log包
	// 在生产环境中，您可能需要使用更健壮的日志解决方案
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// TODO: 根据配置实现基于文件的日志记录
}

// createExampleTasks 创建一些示例任务用于开发目的
func createExampleTasks(sched *scheduler.SchedulerService) {
	// Example 1: 即时Mattermost任务
	mattermostTask := &models.Task{
		Name:     "Mattermost消息处理示例",
		TaskType: models.TypeImmediate,
		Status:   models.StatusPending,
		Priority: models.PriorityHigh,
		Tags:     []string{"MATTERMOST"},
		Parameters: map[string]interface{}{
			"channel_id":   "channel1",
			"message":      "这是一条测试消息",
			"forward_type": "notification",
			"event_type":   "posted",
			"channel_name": "测试频道",
			"username":     "测试用户",
			"notify_admin": "true",
		},
	}

	// Example 2: 定时Mattermost任务
	mattermostScheduledTask := &models.Task{
		Name:     "Mattermost定时通知示例",
		TaskType: models.TypeScheduled,
		CronExpr: "0 0 9 * * *", // 每天早上9点
		Status:   models.StatusPending,
		Priority: models.PriorityMedium,
		Tags:     []string{"MATTERMOST"},
		Parameters: map[string]interface{}{
			"channel_id":        "channel1",
			"message":           "这是一条定时发送的通知",
			"forward_type":      "channel_message",
			"target_channel_id": "channel456",
		},
	}

	// 添加任务
	err := sched.AddTask(mattermostTask)
	if err != nil {
		log.Printf("[main] Failed to add example task 1: %v", err)
	}

	err = sched.AddTask(mattermostScheduledTask)
	if err != nil {
		log.Printf("[main] Failed to add example task 2: %v", err)
	}

	log.Println("[main] Example tasks created")
}
