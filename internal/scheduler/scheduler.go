package scheduler

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"my-scheduler-go/internal/models"
	"my-scheduler-go/internal/repository"

	"github.com/robfig/cron/v3"
)

type SchedulerService struct {
	cron           *cron.Cron
	repo           repository.TaskRepository
	executor       *TaskExecutor
	pollInterval   time.Duration
	maxConcurrency int
	taskQueue      []*models.Task
	queueMutex     sync.Mutex
	runningTasks   map[string]bool
	runningMutex   sync.Mutex
	cronJobs       map[string]cron.EntryID
	cronMutex      sync.Mutex
	stopChan       chan struct{}
}

func NewSchedulerService(repo repository.TaskRepository, executor *TaskExecutor, pollInterval time.Duration) *SchedulerService {
	return &SchedulerService{
		cron:           cron.New(cron.WithSeconds()),
		repo:           repo,
		executor:       executor,
		pollInterval:   pollInterval,
		maxConcurrency: 5, // Default value, can be configured
		taskQueue:      make([]*models.Task, 0),
		runningTasks:   make(map[string]bool),
		cronJobs:       make(map[string]cron.EntryID),
		stopChan:       make(chan struct{}),
	}
}

func (s *SchedulerService) SetMaxConcurrency(maxConcurrency int) {
	s.maxConcurrency = maxConcurrency
}

func (s *SchedulerService) Start() {
	// Poll for pending tasks
	s.cron.AddFunc(fmt.Sprintf("@every %ds", int(s.pollInterval.Seconds())), func() {
		s.pollForNewTasks()
	})

	// Process queued tasks based on priority and dependencies
	s.cron.AddFunc(fmt.Sprintf("@every %ds", 5), func() { // Process queue every 5 seconds
		s.processTaskQueue()
	})

	// Monitor running tasks for timeouts
	s.cron.AddFunc(fmt.Sprintf("@every %ds", 10), func() { // Check for timeouts every 10 seconds
		s.checkTaskTimeouts()
	})

	s.cron.Start()

	// Start a goroutine to handle queue processing
	go s.queueProcessor()

	log.Println("[SchedulerService] Scheduler service started")
}

func (s *SchedulerService) Stop() {
	close(s.stopChan)
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("[SchedulerService] Scheduler service stopped")
}

func (s *SchedulerService) pollForNewTasks() {
	// Get pending tasks
	pending := s.repo.GetTasksByStatus(models.StatusPending)
	if len(pending) == 0 {
		return
	}

	log.Printf("[SchedulerService] Found %d pending tasks", len(pending))

	for _, task := range pending {
		if task.TaskType == models.TypeScheduled && task.CronExpr != "" {
			// Add scheduled task to cron
			s.addScheduledJob(task)
		} else {
			// Queue immediate task
			s.queueTask(task)
		}
	}
}

func (s *SchedulerService) addScheduledJob(task *models.Task) {
	s.cronMutex.Lock()
	defer s.cronMutex.Unlock()

	// If this task is already scheduled, don't schedule again
	if _, exists := s.cronJobs[task.ID]; exists {
		return
	}

	// Update task status
	err := s.repo.UpdateTaskStatus(task.ID, models.StatusScheduled)
	if err != nil {
		log.Printf("[SchedulerService] Failed to update task status: %v", err)
		return
	}

	// Add to cron
	entryID, err := s.cron.AddFunc(task.CronExpr, func() {
		// When the cron job triggers, queue the task for execution
		taskCopy, err := s.repo.GetTaskByID(task.ID)
		if err != nil {
			log.Printf("[SchedulerService] Failed to get task %s: %v", task.ID, err)
			return
		}

		// Skip if task is already queued or running
		if taskCopy.Status == models.StatusQueued || taskCopy.Status == models.StatusRunning {
			return
		}

		// Queue the task
		s.queueTask(taskCopy)
	})

	if err != nil {
		log.Printf("[SchedulerService] Failed to add cron job: %v", err)
		_ = s.repo.UpdateTaskStatus(task.ID, models.StatusFailed)
	} else {
		log.Printf("[SchedulerService] Added scheduled job for task %s with cron %s", task.ID, task.CronExpr)
		s.cronJobs[task.ID] = entryID
	}
}

func (s *SchedulerService) queueTask(task *models.Task) {
	// Update task status to QUEUED
	err := s.repo.UpdateTaskStatus(task.ID, models.StatusQueued)
	if err != nil {
		log.Printf("[SchedulerService] Failed to update task status: %v", err)
		return
	}

	// Add to queue
	s.queueMutex.Lock()
	defer s.queueMutex.Unlock()

	// Check if task already in queue
	for _, t := range s.taskQueue {
		if t.ID == task.ID {
			return
		}
	}

	log.Printf("[SchedulerService] Queuing task %s (%s)", task.ID, task.Name)
	s.taskQueue = append(s.taskQueue, task)
}

func (s *SchedulerService) processTaskQueue() {
	s.queueMutex.Lock()
	defer s.queueMutex.Unlock()

	if len(s.taskQueue) == 0 {
		return
	}

	// Sort queue by priority (HIGH > MEDIUM > LOW)
	sort.SliceStable(s.taskQueue, func(i, j int) bool {
		priorityOrder := map[models.TaskPriority]int{
			models.PriorityHigh:   0,
			models.PriorityMedium: 1,
			models.PriorityLow:    2,
		}
		return priorityOrder[s.taskQueue[i].Priority] < priorityOrder[s.taskQueue[j].Priority]
	})

	// Get completed task IDs
	completedTasks := s.repo.GetCompletedTaskIDs()

	// Process queue
	s.runningMutex.Lock()
	running := len(s.runningTasks)
	s.runningMutex.Unlock()

	availableSlots := s.maxConcurrency - running
	if availableSlots <= 0 {
		return
	}

	// Process up to availableSlots tasks
	processed := 0
	remainingTasks := make([]*models.Task, 0)

	for _, task := range s.taskQueue {
		// If task can be executed (dependencies are satisfied)
		if task.CanBeExecuted(completedTasks) {
			if processed < availableSlots {
				// Execute task
				go s.executeTask(task)
				processed++
			} else {
				// Keep in queue for next processing cycle
				remainingTasks = append(remainingTasks, task)
			}
		} else {
			// Keep in queue, dependencies not satisfied
			remainingTasks = append(remainingTasks, task)
		}
	}

	// Update queue
	s.taskQueue = remainingTasks
}

func (s *SchedulerService) executeTask(task *models.Task) {
	// Mark as running
	s.runningMutex.Lock()
	s.runningTasks[task.ID] = true
	s.runningMutex.Unlock()

	// Execute
	s.executor.ExecuteTask(task.ID)

	// Remove from running tasks
	s.runningMutex.Lock()
	delete(s.runningTasks, task.ID)
	s.runningMutex.Unlock()
}

func (s *SchedulerService) checkTaskTimeouts() {
	// Get running tasks
	runningTasks := s.repo.GetTasksByStatus(models.StatusRunning)

	for _, task := range runningTasks {
		if task.TimeoutSeconds > 0 {
			// Check if task has exceeded timeout
			if task.IsTimeoutReached(task.UpdatedAt) {
				log.Printf("[SchedulerService] Task %s timed out", task.ID)

				// Update status to TIMEOUT
				_ = s.repo.UpdateTaskStatus(task.ID, models.StatusTimeout)

				// If task has retry policy, queue for retry
				if task.RetryPolicy != nil && task.RetryPolicy.MaxRetries > 0 {
					// Create a copy of the task for retry
					taskCopy, err := s.repo.GetTaskByID(task.ID)
					if err == nil {
						// Update metadata with retry info
						if taskCopy.Metadata == nil {
							taskCopy.Metadata = make(map[string]interface{})
						}

						retryCount, ok := taskCopy.Metadata["retry_count"].(int)
						if !ok {
							retryCount = 0
						}

						if retryCount < task.RetryPolicy.MaxRetries {
							taskCopy.Metadata["retry_count"] = retryCount + 1
							taskCopy.Metadata["original_task_id"] = task.ID

							// Update the task
							_ = s.repo.UpdateTask(taskCopy)

							// Set status to RETRY
							_ = s.repo.UpdateTaskStatus(task.ID, models.StatusRetry)

							// Queue for retry after delay
							go func(t *models.Task) {
								delay := t.RetryPolicy.RetryDelay
								if retryCount > 0 {
									// Apply backoff
									delay = time.Duration(float64(delay) * t.RetryPolicy.BackoffFactor * float64(retryCount))
								}

								log.Printf("[SchedulerService] Task %s will retry in %v", t.ID, delay)
								time.Sleep(delay)
								s.queueTask(t)
							}(taskCopy)
						} else {
							// Max retries reached, mark as failed
							_ = s.repo.UpdateTaskStatus(task.ID, models.StatusFailed)
						}
					}
				} else {
					// No retry policy, mark as failed
					_ = s.repo.UpdateTaskStatus(task.ID, models.StatusFailed)
				}
			}
		}
	}
}

func (s *SchedulerService) queueProcessor() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processTaskQueue()
		case <-s.stopChan:
			return
		}
	}
}

// AddTask adds a new task to the scheduler
func (s *SchedulerService) AddTask(task *models.Task) error {
	// Save to repository
	err := s.repo.AddTask(task)
	if err != nil {
		return err
	}

	// If immediate task, queue immediately
	if task.TaskType == models.TypeImmediate {
		s.queueTask(task)
	} else if task.TaskType == models.TypeScheduled && task.CronExpr != "" {
		s.addScheduledJob(task)
	}

	return nil
}
