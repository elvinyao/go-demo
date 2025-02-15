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
	// 在更新状态为 RUNNING 之前，先获取任务对象:
	tasks := te.repo.GetAllTasks()
	var currentTask *models.Task
	for _, t := range tasks {
		if t.ID == taskID {
			currentTask = t
			break
		}
	}

	if currentTask == nil {
		log.Printf("[TaskExecutor] Task %d not found.", taskID)
		return
	}

	err := te.repo.UpdateTaskStatus(taskID, models.StatusRunning)
	if err != nil {
		log.Printf("[TaskExecutor] Update task status error: %v", err)
		return
	}
	log.Printf("[TaskExecutor] Task %d is RUNNING... Type=%s", taskID, currentTask.TaskType)

	// 根据不同TaskType做不同处理(仅演示):
	switch currentTask.TaskType {
	case models.TaskTypeSpecial:
		log.Printf("[TaskExecutor] Executing SPECIAL logic for Task %d...", taskID)
		// 这里可以写具体的Special任务实现
	default:
		log.Printf("[TaskExecutor] Executing default logic for Task %d...", taskID)
	}

	_ = te.repo.UpdateTaskStatus(taskID, models.StatusDone)
	log.Printf("[TaskExecutor] Task %d is DONE.", taskID)
}
