package main

import (
	"log"
	"net/http"
	"time"

	"my-scheduler-go/internal/api"
	"my-scheduler-go/internal/config"
	"my-scheduler-go/internal/logger"
	"my-scheduler-go/internal/repository"
	"my-scheduler-go/internal/scheduler"
)

func main() {
	// 1. 加载配置
	appConfig, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化日志
	// 将 config 里 log 字段映射到 logger.LogConfig
	logCfg := logger.LogConfig{
		Level:       appConfig.Log.Level,
		Output:      appConfig.Log.Output,
		Format:      appConfig.Log.Format,
		Filename:    appConfig.Log.Filename,
		MaxBytes:    appConfig.Log.MaxBytes,
		BackupCount: appConfig.Log.BackupCount,
		MaxAgeDays:  appConfig.Log.MaxAgeDays,
		Compress:    appConfig.Log.Compress,
	}
	if err = logger.InitLogger(logCfg); err != nil {
		panic("Failed to init logger: " + err.Error())
	}
	defer logger.Sync()

	logger.L.Info("My Scheduler Go started with new logger config.")

	// 3. 初始化存储 (In-Memory)
	repo := repository.NewInMemoryTaskRepository()
	logger.L.Info("[main] In-memory TaskRepository initialized.")

	// 4. 初始化执行器
	executor := scheduler.NewTaskExecutor(repo)

	// 5. 创建调度服务
	pollInterval := time.Duration(appConfig.Scheduler.PollInterval) * time.Second
	schedService := scheduler.NewSchedulerService(repo, executor, pollInterval)
	schedService.Start()
	defer schedService.Stop()
	logger.L.Info("[main] SchedulerService started.")

	// 6. 启动 HTTP 服务
	router := api.SetupRouter(repo)
	port := ":8000"
	logger.L.Info("[main] Starting HTTP server on %s\n", port)
	if err := http.ListenAndServe(port, router); err != nil {
		logger.L.Errorf("HTTP server failed: %v", err)
	}
	logger.L.Info("[main] My Scheduler Go shutdown gracefully.")
}
