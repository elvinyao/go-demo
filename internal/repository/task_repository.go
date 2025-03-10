package repository

import (
	"errors"
	"fmt"
	"my-scheduler-go/internal/models"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

type TaskRepository interface {
	AddTask(task *models.Task) error
	GetAllTasks() []*models.Task
	GetTasksByStatus(status models.TaskStatus) []*models.Task
	GetTasksByStatusAndTags(status models.TaskStatus, tags []string) []*models.Task
	GetTasksByTags(tags []string) []*models.Task
	GetTaskByID(id string) (*models.Task, error)
	UpdateTaskStatus(id string, newStatus models.TaskStatus) error
	UpdateTask(task *models.Task) error
	DeleteTask(id string) error
	GetDependentTasks(taskID string) []*models.Task
	GetCompletedTaskIDs() map[string]bool
}

type InMemoryTaskRepository struct {
	tasks map[string]*models.Task
	mu    sync.RWMutex
}

func NewInMemoryTaskRepository() *InMemoryTaskRepository {
	return &InMemoryTaskRepository{
		tasks: make(map[string]*models.Task),
	}
}

func (r *InMemoryTaskRepository) AddTask(task *models.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Generate UUID if not provided
	if task.ID == "" {
		task.ID = uuid.New().String()
	}

	// Set default values if not provided
	if task.Priority == "" {
		task.Priority = models.PriorityMedium
	}

	if task.Status == "" {
		task.Status = models.StatusPending
	}

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

func (r *InMemoryTaskRepository) GetTasksByStatusAndTags(status models.TaskStatus, tags []string) []*models.Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Task
	for _, t := range r.tasks {
		if t.Status != status {
			continue
		}

		// Check if the task has any of the specified tags
		hasTag := false
		for _, wantTag := range tags {
			for _, taskTag := range t.Tags {
				if taskTag == wantTag {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}

		if hasTag {
			result = append(result, t)
		}
	}
	return result
}

func (r *InMemoryTaskRepository) GetTasksByTags(tags []string) []*models.Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Task
	for _, t := range r.tasks {
		// Check if the task has any of the specified tags
		for _, wantTag := range tags {
			for _, taskTag := range t.Tags {
				if taskTag == wantTag {
					result = append(result, t)
					break
				}
			}
		}
	}
	return result
}

func (r *InMemoryTaskRepository) UpdateTaskStatus(id string, newStatus models.TaskStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok {
		return fmt.Errorf("task %s not found", id)
	}
	task.UpdateStatus(newStatus)
	return nil
}

func (r *InMemoryTaskRepository) GetTaskByID(id string) (*models.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	if !ok {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

func (r *InMemoryTaskRepository) UpdateTask(task *models.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.tasks[task.ID]
	if !ok {
		return ErrTaskNotFound
	}

	task.UpdatedAt = time.Now()
	r.tasks[task.ID] = task
	return nil
}

func (r *InMemoryTaskRepository) DeleteTask(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.tasks[id]
	if !ok {
		return ErrTaskNotFound
	}

	delete(r.tasks, id)
	return nil
}

func (r *InMemoryTaskRepository) GetDependentTasks(taskID string) []*models.Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Task
	for _, t := range r.tasks {
		for _, depID := range t.Dependencies {
			if depID == taskID {
				result = append(result, t)
				break
			}
		}
	}
	return result
}

func (r *InMemoryTaskRepository) GetCompletedTaskIDs() map[string]bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]bool)
	for _, t := range r.tasks {
		if t.Status == models.StatusDone {
			result[t.ID] = true
		}
	}
	return result
}
