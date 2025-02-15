package models

import "time"

type TaskStatus string

const (
	StatusPending   TaskStatus = "PENDING"
	StatusScheduled TaskStatus = "SCHEDULED"
	StatusRunning   TaskStatus = "RUNNING"
	StatusDone      TaskStatus = "DONE"
	StatusFailed    TaskStatus = "FAILED"
)

// 在文件顶部或合适位置新增:
type TaskType string

const (
	TaskTypeDefault TaskType = "DEFAULT"
	TaskTypeSpecial TaskType = "SPECIAL"
)

type Task struct {
	ID       int64      `json:"id"`
	Name     string     `json:"name"`
	CronExpr string     `json:"cron_expr,omitempty"`
	Status   TaskStatus `json:"status"`
	// 新增:
	TaskType  TaskType  `json:"task_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (t *Task) UpdateStatus(newStatus TaskStatus) {
	t.Status = newStatus
	t.UpdatedAt = time.Now()
}
