package service

import (
	"log"

	"my-scheduler-go/internal/models"
	"my-scheduler-go/internal/repository"
)

// TaskService 示例, 综合调用repo/executor等逻辑
type TaskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) CreateImmediateTask(name string) (*models.Task, error) {
	task := &models.Task{
		Name:   name,
		Status: models.StatusPending,
	}
	if err := s.repo.AddTask(task); err != nil {
		return nil, err
	}
	log.Printf("[TaskService] Created immediate task: %s", name)
	return task, nil
}
