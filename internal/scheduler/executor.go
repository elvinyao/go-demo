package scheduler

import (
	"log"

	"my-scheduler-go/internal/models"
	"my-scheduler-go/internal/repository"
)

type TaskExecutor struct {
	repo repository.TaskRepository
}

func NewTaskExecutor(repo repository.TaskRepository) *TaskExecutor {
	return &TaskExecutor{repo: repo}
}

func (te *TaskExecutor) ExecuteTask(taskID int64) {
	err := te.repo.UpdateTaskStatus(taskID, models.StatusRunning)
	if err != nil {
		log.Printf("[TaskExecutor] Update task status error: %v", err)
		return
	}

	log.Printf("[TaskExecutor] Task %d is RUNNING...", taskID)

	// 模拟处理流程
	// (这里仅打印日志表示"执行")
	// 真实逻辑可调用 handlers/confluence_handler 等

	// 随机失败或成功(示例)
	// 这里简单地都成功处理
	_ = te.repo.UpdateTaskStatus(taskID, models.StatusDone)
	log.Printf("[TaskExecutor] Task %d is DONE.", taskID)
}
