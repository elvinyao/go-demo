package api

import (
	"net/http"
	"time"

	"my-scheduler-go/internal/models"
	"my-scheduler-go/internal/repository"
	"my-scheduler-go/internal/scheduler"

	"github.com/gin-gonic/gin"
)

// API represents the API handler
type API struct {
	repo      repository.TaskRepository
	scheduler *scheduler.SchedulerService
}

// NewAPI creates a new API handler
func NewAPI(repo repository.TaskRepository, scheduler *scheduler.SchedulerService) *API {
	return &API{
		repo:      repo,
		scheduler: scheduler,
	}
}

// SetupRouter sets up the API routes
func SetupRouter(repo repository.TaskRepository, scheduler *scheduler.SchedulerService) *gin.Engine {
	r := gin.Default()
	api := NewAPI(repo, scheduler)

	// Task management endpoints
	r.GET("/tasks", api.GetAllTasks)
	r.GET("/tasks/status/:status", api.GetTasksByStatus)
	r.GET("/tasks/tags/:tag", api.GetTasksByTag)
	r.GET("/tasks/:id", api.GetTaskByID)
	r.POST("/tasks", api.CreateTask)
	r.PUT("/tasks/:id", api.UpdateTask)
	r.DELETE("/tasks/:id", api.DeleteTask)

	// Task history endpoint
	r.GET("/task_history", api.GetTaskHistory)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now(),
		})
	})

	return r
}

// GetAllTasks returns all tasks
func (api *API) GetAllTasks(c *gin.Context) {
	tasks := api.repo.GetAllTasks()
	c.JSON(http.StatusOK, gin.H{
		"total_count": len(tasks),
		"data":        tasks,
	})
}

// GetTasksByStatus returns tasks by status
func (api *API) GetTasksByStatus(c *gin.Context) {
	status := c.Param("status")
	tasks := api.repo.GetTasksByStatus(models.TaskStatus(status))
	c.JSON(http.StatusOK, gin.H{
		"total_count": len(tasks),
		"data":        tasks,
	})
}

// GetTasksByTag returns tasks with a specific tag
func (api *API) GetTasksByTag(c *gin.Context) {
	tag := c.Param("tag")
	tasks := api.repo.GetTasksByTags([]string{tag})
	c.JSON(http.StatusOK, gin.H{
		"total_count": len(tasks),
		"data":        tasks,
	})
}

// GetTaskByID returns a task by ID
func (api *API) GetTaskByID(c *gin.Context) {
	id := c.Param("id")
	task, err := api.repo.GetTaskByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
		})
		return
	}
	c.JSON(http.StatusOK, task)
}

// CreateTask creates a new task
func (api *API) CreateTask(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Add the task using the scheduler service
	err := api.scheduler.AddTask(&task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// UpdateTask updates an existing task
func (api *API) UpdateTask(c *gin.Context) {
	id := c.Param("id")

	// Check if task exists
	_, err := api.repo.GetTaskByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
		})
		return
	}

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Ensure ID matches
	task.ID = id

	// Update the task
	err = api.repo.UpdateTask(&task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTask deletes a task
func (api *API) DeleteTask(c *gin.Context) {
	id := c.Param("id")

	// Check if task exists
	_, err := api.repo.GetTaskByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
		})
		return
	}

	// Delete the task
	err = api.repo.DeleteTask(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task deleted successfully",
	})
}

// GetTaskHistory returns completed or failed tasks
func (api *API) GetTaskHistory(c *gin.Context) {
	done := api.repo.GetTasksByStatus(models.StatusDone)
	failed := api.repo.GetTasksByStatus(models.StatusFailed)
	timeout := api.repo.GetTasksByStatus(models.StatusTimeout)

	// Combine all history tasks
	tasks := append(done, failed...)
	tasks = append(tasks, timeout...)

	c.JSON(http.StatusOK, gin.H{
		"total_count": len(tasks),
		"data":        tasks,
	})
}
