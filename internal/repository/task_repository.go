package repository

import (
	"fmt"
	"my-scheduler-go/internal/models"
	"sync"
	"time"
)

type TaskRepository interface {
	AddTask(task *models.Task) error
	GetAllTasks() []*models.Task
	GetTasksByStatus(status models.TaskStatus) []*models.Task
	UpdateTaskStatus(id int64, newStatus models.TaskStatus) error
}

type InMemoryTaskRepository struct {
	tasks  map[int64]*models.Task
	mu     sync.RWMutex
	nextID int64
}

func NewInMemoryTaskRepository() *InMemoryTaskRepository {
	return &InMemoryTaskRepository{
		tasks:  make(map[int64]*models.Task),
		nextID: 1,
	}
}

func (r *InMemoryTaskRepository) AddTask(task *models.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task.ID = r.nextID
	r.nextID++
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	r.tasks[task.ID] = task
	return nil
}

func (r *InMemoryTaskRepository) GetAllTasks() []*models.Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Task
	for _, t := range r.tasks {
		result = append(result, t)
	}
	return result
}

func (r *InMemoryTaskRepository) GetTasksByStatus(status models.TaskStatus) []*models.Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Task
	for _, t := range r.tasks {
		if t.Status == status {
			result = append(result, t)
		}
	}
	return result
}

func (r *InMemoryTaskRepository) UpdateTaskStatus(id int64, newStatus models.TaskStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok {
		return fmt.Errorf("task %d not found", id)
	}
	task.UpdateStatus(newStatus)
	return nil
}
