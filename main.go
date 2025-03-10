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
	"my-scheduler-go/internal/models"
	"my-scheduler-go/internal/repository"
	"my-scheduler-go/internal/scheduler"
	"my-scheduler-go/internal/service"
)

func main() {
	// 1. Load configuration
	appConfig, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Setup logging
	setupLogging(appConfig)
	log.Println("[main] Starting APScheduler Task Management System...")

	// 3. Initialize task repository
	repo := repository.NewInMemoryTaskRepository()
	log.Println("[main] Task repository initialized")

	// 4. Initialize task executor
	executor := scheduler.NewTaskExecutor(repo)
	log.Println("[main] Task executor initialized")

	// 5. Create scheduler service
	pollInterval := time.Duration(appConfig.Scheduler.PollInterval) * time.Second
	schedService := scheduler.NewSchedulerService(repo, executor, pollInterval)

	// Set max concurrency from config
	schedService.SetMaxConcurrency(appConfig.Scheduler.Concurrency)

	// Start scheduler service
	schedService.Start()
	log.Println("[main] Scheduler service started")

	// 6. Initialize and start result reporting service
	reportingService := service.NewResultReportingService(repo, appConfig)
	reportingService.Start()
	log.Println("[main] Result reporting service started")

	// 7. Create example tasks if in development mode
	if appConfig.Environment == "development" {
		createExampleTasks(schedService)
	}

	// 8. Setup HTTP server with API routes
	router := api.SetupRouter(repo, schedService, reportingService)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":8000",
		Handler: router,
	}

	// 9. Start HTTP server in a separate goroutine
	go func() {
		log.Printf("[main] HTTP server listening on %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 10. Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[main] Shutdown signal received, stopping services...")

	// 11. Shutdown services gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Stop reporting service
	reportingService.Stop()
	log.Println("[main] Result reporting service stopped")

	// Stop scheduler
	schedService.Stop()
	log.Println("[main] Scheduler service stopped")

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
	log.Println("[main] HTTP server stopped")

	log.Println("[main] APScheduler Task Management System shutdown complete")
}

// setupLogging configures the application logging
func setupLogging(appConfig *config.AppConfig) {
	// For this example, we're using the standard log package
	// In a production environment, you might want to use a more robust logging solution
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// TODO: Implement file-based logging based on config
}

// createExampleTasks creates some example tasks for development purposes
func createExampleTasks(sched *scheduler.SchedulerService) {
	// Example 1: Immediate JIRA task
	jiraImmediateTask := &models.Task{
		Name:     "JIRA Extraction - Root Ticket",
		TaskType: models.TypeImmediate,
		Status:   models.StatusPending,
		Priority: models.PriorityHigh,
		Tags:     []string{"JIRA_TASK_EXP"},
		Parameters: map[string]interface{}{
			"jira_envs": []string{"env1.jira.com", "env2.jira.com"},
			"key_type":  "root_ticket",
			"key_value": "PROJ-123",
			"user":      "johndoe",
		},
	}

	// Example 2: Scheduled JIRA task (daily at midnight)
	jiraScheduledTask := &models.Task{
		Name:     "JIRA Extraction - Project (Daily)",
		TaskType: models.TypeScheduled,
		CronExpr: "0 0 0 * * *", // Seconds Minutes Hours Day Month DayOfWeek
		Status:   models.StatusPending,
		Priority: models.PriorityMedium,
		Tags:     []string{"JIRA_TASK_EXP"},
		Parameters: map[string]interface{}{
			"jira_envs": []string{"env1.jira.com"},
			"key_type":  "project",
			"key_value": "PROJ",
			"user":      "johndoe",
		},
	}

	// Example 3: Task with timeout and retry policy
	taskWithRetry := &models.Task{
		Name:           "Task with Timeout and Retry",
		TaskType:       models.TypeImmediate,
		Status:         models.StatusPending,
		Priority:       models.PriorityLow,
		TimeoutSeconds: 5,
		RetryPolicy: &models.RetryPolicy{
			MaxRetries:    3,
			RetryDelay:    time.Second * 5,
			BackoffFactor: 2.0,
		},
	}

	// Example 4: Task with dependencies (depends on Example 1)
	dependentTask := &models.Task{
		Name:         "Dependent Task",
		TaskType:     models.TypeImmediate,
		Status:       models.StatusPending,
		Priority:     models.PriorityMedium,
		Dependencies: []string{}, // Will be populated after the first task is created
	}

	// Add tasks
	err := sched.AddTask(jiraImmediateTask)
	if err != nil {
		log.Printf("[main] Failed to add example task 1: %v", err)
	} else {
		// Update dependent task to depend on the first task
		dependentTask.Dependencies = []string{jiraImmediateTask.ID}
	}

	err = sched.AddTask(jiraScheduledTask)
	if err != nil {
		log.Printf("[main] Failed to add example task 2: %v", err)
	}

	err = sched.AddTask(taskWithRetry)
	if err != nil {
		log.Printf("[main] Failed to add example task 3: %v", err)
	}

	err = sched.AddTask(dependentTask)
	if err != nil {
		log.Printf("[main] Failed to add example task 4: %v", err)
	}

	log.Println("[main] Example tasks created")
}
