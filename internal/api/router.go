package api

import (
	"net/http"

	"my-scheduler-go/internal/models"
	"my-scheduler-go/internal/repository"

	"github.com/gin-gonic/gin"
)

func SetupRouter(repo repository.TaskRepository) *gin.Engine {
	r := gin.Default()

	// GET /tasks - 返回所有任务
	r.GET("/tasks", func(c *gin.Context) {
		tasks := repo.GetAllTasks()
		c.JSON(http.StatusOK, gin.H{
			"total_count": len(tasks),
			"data":        tasks,
		})
	})

	// GET /tasks/status/:status - 按状态过滤
	r.GET("/tasks/status/:status", func(c *gin.Context) {
		status := c.Param("status")
		tasks := repo.GetTasksByStatus(models.TaskStatus(status))
		c.JSON(http.StatusOK, gin.H{
			"total_count": len(tasks),
			"data":        tasks,
		})
	})

	// GET /task_history - 查看 DONE / FAILED 任务
	r.GET("/task_history", func(c *gin.Context) {
		done := repo.GetTasksByStatus(models.StatusDone)
		failed := repo.GetTasksByStatus(models.StatusFailed)
		tasks := append(done, failed...)
		c.JSON(http.StatusOK, gin.H{
			"total_count": len(tasks),
			"data":        tasks,
		})
	})

	return r
}
