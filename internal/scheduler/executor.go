package scheduler

import (
	"fmt"
	"log"
	"my-scheduler-go/internal/models"
	"my-scheduler-go/internal/repository"
	"sync"
	"time"
)

// TaskHandler 定义任务处理函数类型
type TaskHandler func(task *models.Task) error

// TaskExecutor 负责执行任务的组件
type TaskExecutor struct {
	repo         repository.TaskRepository
	handlerMutex sync.RWMutex
	taskHandlers map[string]TaskHandler // 通过标签映射到处理函数
}

// NewTaskExecutor 创建新的任务执行器
func NewTaskExecutor(repo repository.TaskRepository) *TaskExecutor {
	return &TaskExecutor{
		repo:         repo,
		taskHandlers: make(map[string]TaskHandler),
	}
}

// RegisterHandler 注册特定类型任务的处理函数
func (e *TaskExecutor) RegisterHandler(tag string, handler TaskHandler) {
	e.handlerMutex.Lock()
	defer e.handlerMutex.Unlock()
	e.taskHandlers[tag] = handler
	log.Printf("[TaskExecutor] Registered handler for tag: %s", tag)
}

// ExecuteTask 执行单个任务
func (e *TaskExecutor) ExecuteTask(task *models.Task) error {
	log.Printf("[TaskExecutor] Executing task '%s' (ID: %s)", task.Name, task.ID)

	// 已存在的执行器代码...
	if task.Status != models.StatusPending && task.Status != models.StatusRetry {
		return fmt.Errorf("task not in executable state: %s", task.Status)
	}

	// 更新任务状态
	task.Status = models.StatusRunning
	task.StartTime = time.Now()
	if err := e.repo.UpdateTask(task); err != nil {
		return err
	}

	var err error
	var result string

	// 首先查找匹配的处理器
	handler := e.findHandler(task)

	if handler != nil {
		// 使用注册的处理器处理任务
		err = handler(task)
		if err != nil {
			result = fmt.Sprintf("Error: %v", err)
		} else {
			result = "Success"
		}
	} else {
		// 使用通用处理逻辑
		result, err = e.executeTaskLogic(task)
	}

	// 更新任务结果
	task.EndTime = time.Now()
	task.ExecutionResult = map[string]interface{}{
		"result": result,
	}
	task.Status = models.StatusDone

	if err != nil {
		log.Printf("[TaskExecutor] Task execution failed: %v", err)
		task.Status = models.StatusFailed

		// 重试逻辑
		if task.RetryPolicy != nil && task.RetryCount < task.RetryPolicy.MaxRetries {
			task.RetryCount++
			task.Status = models.StatusRetry
			task.NextRunAt = time.Now().Add(task.RetryPolicy.RetryDelay * time.Duration(task.RetryPolicy.BackoffFactor))
			log.Printf("[TaskExecutor] Scheduling retry %d for task ID %s at %v",
				task.RetryCount, task.ID, task.NextRunAt)
		}
	}

	// 保存任务状态
	return e.repo.UpdateTask(task)
}

// 查找匹配的处理器
func (e *TaskExecutor) findHandler(task *models.Task) TaskHandler {
	e.handlerMutex.RLock()
	defer e.handlerMutex.RUnlock()

	// 按标签查找处理器
	for _, tag := range task.Tags {
		if handler, exists := e.taskHandlers[tag]; exists {
			return handler
		}
	}

	return nil
}

// executeTaskLogic 包含默认的任务执行逻辑
func (e *TaskExecutor) executeTaskLogic(task *models.Task) (string, error) {
	// 简单模拟任务执行过程
	log.Printf("[TaskExecutor] Simulating execution of task: %s", task.Name)

	// 通用处理逻辑...
	time.Sleep(1 * time.Second) // 模拟工作

	// 获取任务参数
	params := task.Parameters
	if params == nil {
		return "Task completed with no parameters", nil
	}

	// 打印参数
	log.Printf("[TaskExecutor] Task parameters: %+v", params)

	return "Task executed successfully", nil
}

// handleTaskExecution handles the execution of a task based on its tags
func (te *TaskExecutor) handleTaskExecution(task *models.Task) map[string]interface{} {
	result := make(map[string]interface{})

	// Check for JIRA related tasks
	if containsTag(task.Tags, "JIRA_TASK_EXP") {
		// Here would be the JIRA API integration
		log.Printf("[TaskExecutor] Executing JIRA task: %s", task.Name)

		// Simulate JIRA task execution
		if task.Parameters != nil {
			if keyType, ok := task.Parameters["key_type"].(string); ok {
				if keyValue, ok := task.Parameters["key_value"].(string); ok {
					log.Printf("[TaskExecutor] Processing JIRA %s: %s", keyType, keyValue)
					result["jira_processed"] = true
					result["jira_key_type"] = keyType
					result["jira_key_value"] = keyValue
				}
			}
		}
	}

	// Check for Confluence related tasks
	if containsTag(task.Tags, "CONFLUENCE_TASK") {
		// Here would be the Confluence API integration
		log.Printf("[TaskExecutor] Executing Confluence task: %s", task.Name)

		// Simulate Confluence task execution
		result["confluence_processed"] = true
	}

	// Generic task processing for other task types
	if len(task.Tags) == 0 || (!containsTag(task.Tags, "JIRA_TASK_EXP") && !containsTag(task.Tags, "CONFLUENCE_TASK")) {
		log.Printf("[TaskExecutor] Executing generic task: %s", task.Name)
		result["generic_processed"] = true
	}

	// Simulate some processing time
	time.Sleep(1 * time.Second)

	return result
}

// containsTag checks if a tag is in the tags slice
func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
