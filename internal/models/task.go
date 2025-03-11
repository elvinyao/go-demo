package models

import (
	"time"
)

type TaskStatus string
type TaskType string
type TaskPriority string

const (
	// Task Status Constants
	StatusPending   TaskStatus = "PENDING"
	StatusQueued    TaskStatus = "QUEUED"
	StatusScheduled TaskStatus = "SCHEDULED"
	StatusRunning   TaskStatus = "RUNNING"
	StatusDone      TaskStatus = "DONE"
	StatusFailed    TaskStatus = "FAILED"
	StatusTimeout   TaskStatus = "TIMEOUT"
	StatusRetry     TaskStatus = "RETRY"

	// Task Type Constants
	TypeImmediate TaskType = "IMMEDIATE"
	TypeScheduled TaskType = "SCHEDULED"

	// Task Priority Constants
	PriorityHigh   TaskPriority = "HIGH"
	PriorityMedium TaskPriority = "MEDIUM"
	PriorityLow    TaskPriority = "LOW"
)

// RetryPolicy defines how a task should be retried if it fails
type RetryPolicy struct {
	MaxRetries    int           `json:"max_retries"`
	RetryDelay    time.Duration `json:"retry_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// Task represents a scheduled job in the system
type Task struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	TaskType        TaskType               `json:"task_type"`
	CronExpr        string                 `json:"cron_expr,omitempty"`
	Status          TaskStatus             `json:"status"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	StartTime       time.Time              `json:"start_time,omitempty"`
	EndTime         time.Time              `json:"end_time,omitempty"`
	Priority        TaskPriority           `json:"priority"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	Owner           string                 `json:"owner,omitempty"`
	Dependencies    []string               `json:"dependencies,omitempty"`
	TimeoutSeconds  int                    `json:"timeout_seconds,omitempty"`
	RetryPolicy     *RetryPolicy           `json:"retry_policy,omitempty"`
	Parameters      map[string]interface{} `json:"parameters,omitempty"`
	ExecutionResult map[string]interface{} `json:"execution_result,omitempty"`
	RetryCount      int                    `json:"retry_count,omitempty"`
	NextRunAt       time.Time              `json:"next_run_at,omitempty"`
}

func (t *Task) UpdateStatus(newStatus TaskStatus) {
	t.Status = newStatus
	t.UpdatedAt = time.Now()
}

// IsTimeoutReached checks if the task execution has exceeded its timeout
func (t *Task) IsTimeoutReached(startTime time.Time) bool {
	if t.TimeoutSeconds <= 0 {
		return false
	}
	return time.Since(startTime) > time.Duration(t.TimeoutSeconds)*time.Second
}

// CanBeExecuted checks if a task can be executed based on its dependencies
func (t *Task) CanBeExecuted(completedTaskIDs map[string]bool) bool {
	if len(t.Dependencies) == 0 {
		return true
	}

	for _, depID := range t.Dependencies {
		if !completedTaskIDs[depID] {
			return false
		}
	}
	return true
}
