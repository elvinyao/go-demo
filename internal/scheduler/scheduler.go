package scheduler

import (
	"fmt"
	"log"
	"time"

	"my-scheduler-go/internal/models"
	"my-scheduler-go/internal/repository"

	"github.com/robfig/cron/v3"
)

type SchedulerService struct {
	cron         *cron.Cron
	repo         repository.TaskRepository
	executor     *TaskExecutor
	pollInterval time.Duration
}

func NewSchedulerService(repo repository.TaskRepository, executor *TaskExecutor, pollInterval time.Duration) *SchedulerService {
	return &SchedulerService{
		cron:         cron.New(cron.WithSeconds()),
		repo:         repo,
		executor:     executor,
		pollInterval: pollInterval,
	}
}

func (s *SchedulerService) Start() {
	// 轮询PENDING任务
	s.cron.AddFunc(fmt.Sprintf("@every %ds", int(s.pollInterval.Seconds())), func() {
		s.pollDBForNewTasks()
	})

	// 演示: 每分钟都做些事情
	s.cron.AddFunc("0 * * * * *", func() {
		log.Println("[SchedulerService] Running a demo cron job every minute.")
	})

	s.cron.Start()
}

func (s *SchedulerService) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("[SchedulerService] Cron stopped.")
}

func (s *SchedulerService) pollDBForNewTasks() {
	pending := s.repo.GetTasksByStatus(models.StatusPending)
	for _, task := range pending {
		if task.CronExpr != "" {
			// 有cron表达式
			s.addScheduledJob(task)
		} else {
			// 无cron立即执行
			s.addImmediateJob(task)
		}
	}
}

func (s *SchedulerService) addScheduledJob(task *models.Task) {
	_ = s.repo.UpdateTaskStatus(task.ID, models.StatusScheduled)
	_, err := s.cron.AddFunc(task.CronExpr, func() {
		s.executor.ExecuteTask(task.ID)
	})
	if err != nil {
		log.Printf("[SchedulerService] Failed to add cron job: %v", err)
		_ = s.repo.UpdateTaskStatus(task.ID, models.StatusFailed)
	} else {
		log.Printf("[SchedulerService] Added scheduled job for task %d with cron %s", task.ID, task.CronExpr)
	}
}

func (s *SchedulerService) addImmediateJob(task *models.Task) {
	_ = s.repo.UpdateTaskStatus(task.ID, models.StatusScheduled)
	go s.executor.ExecuteTask(task.ID)
	log.Printf("[SchedulerService] Added immediate job for task %d", task.ID)
}
