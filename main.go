package main

import (
	"log"
	"net/http"
	"time"

	"my-scheduler-go/internal/api"
	"my-scheduler-go/internal/config"
	"my-scheduler-go/internal/repository"
	"my-scheduler-go/internal/scheduler"
)

func main() {
	// 1. 加载配置
	appConfig, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. (示例) 初始化日志（这里用标准库log做简单示例）
	log.Println("[main] Starting My Scheduler Go...")

	// 3. 初始化存储 (In-Memory)
	repo := repository.NewInMemoryTaskRepository()
	log.Println("[main] In-memory TaskRepository initialized.")

	// 4. 初始化执行器
	executor := scheduler.NewTaskExecutor(repo)

	// 5. 创建调度服务
	pollInterval := time.Duration(appConfig.Scheduler.PollInterval) * time.Second
	schedService := scheduler.NewSchedulerService(repo, executor, pollInterval)
	schedService.Start()
	defer schedService.Stop()
	log.Println("[main] SchedulerService started.")

	// 6. 启动 HTTP 服务
	router := api.SetupRouter(repo)
	port := ":8000"
	log.Printf("[main] Starting HTTP server on %s\n", port)
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
	log.Println("[main] My Scheduler Go shutdown gracefully.")
}
