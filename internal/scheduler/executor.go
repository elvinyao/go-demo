package scheduler

import (
	"log"
	"time"

	"my-scheduler-go/internal/models"
	"my-scheduler-go/internal/repository"
)

// TaskExecutor executes tasks from the queue
type TaskExecutor struct {
	repo repository.TaskRepository
}

// NewTaskExecutor creates a new task executor
func NewTaskExecutor(repo repository.TaskRepository) *TaskExecutor {
	return &TaskExecutor{repo: repo}
}

// ExecuteTask executes a task by ID
func (te *TaskExecutor) ExecuteTask(taskID string) {
	// Get latest task state from repository
	task, err := te.repo.GetTaskByID(taskID)
	if err != nil {
		log.Printf("[TaskExecutor] Failed to get task %s: %v", taskID, err)
		return
	}

	// Update task status to RUNNING
	err = te.repo.UpdateTaskStatus(taskID, models.StatusRunning)
	if err != nil {
		log.Printf("[TaskExecutor] Update task status error: %v", err)
		return
	}

	log.Printf("[TaskExecutor] Task %s (%s) is RUNNING...", taskID, task.Name)

	startTime := time.Now()

	// Execute task based on tags
	result := te.handleTaskExecution(task)

	// Store execution result
	if task.ExecutionResult == nil {
		task.ExecutionResult = make(map[string]interface{})
	}

	// Add execution info to result
	task.ExecutionResult["execution_time"] = time.Since(startTime).String()
	task.ExecutionResult["executed_at"] = time.Now()

	// Merge custom execution results if any
	if result != nil {
		for k, v := range result {
			task.ExecutionResult[k] = v
		}
	}

	// Update task with results
	err = te.repo.UpdateTask(task)
	if err != nil {
		log.Printf("[TaskExecutor] Failed to update task with results: %v", err)
	}

	// Mark task as DONE
	err = te.repo.UpdateTaskStatus(taskID, models.StatusDone)
	if err != nil {
		log.Printf("[TaskExecutor] Failed to mark task as done: %v", err)
		return
	}

	log.Printf("[TaskExecutor] Task %s is DONE. Execution time: %v", taskID, time.Since(startTime))
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
