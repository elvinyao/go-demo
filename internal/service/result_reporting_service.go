package service

import (
	"fmt"
	"log"
	"sync"
	"time"

	"my-scheduler-go/internal/config"
	"my-scheduler-go/internal/models"
	"my-scheduler-go/internal/repository"
)

// ReportingStrategy is an interface for different reporting methods
type ReportingStrategy interface {
	GenerateReport(tasks []*models.Task) (string, error)
	PublishReport(reportData string) error
}

// ConfluenceReporter implements reporting to Confluence
type ConfluenceReporter struct {
	config            *config.AppConfig
	confluenceService *ConfluenceService
}

// MattermostReporter implements reporting to Mattermost
type MattermostReporter struct {
	config            *config.AppConfig
	mattermostService *MattermostService
}

// ResultReportingService handles task result aggregation and reporting
type ResultReportingService struct {
	repo             repository.TaskRepository
	config           *config.AppConfig
	interval         time.Duration
	reportStrategies map[string]ReportingStrategy
	stopChan         chan struct{}
	runningMutex     sync.Mutex
	isRunning        bool
}

// NewResultReportingService creates a new result reporting service
func NewResultReportingService(repo repository.TaskRepository, config *config.AppConfig) *ResultReportingService {
	// Default to 30 seconds if not configured
	interval := 30
	if config.Reporting.Interval > 0 {
		interval = config.Reporting.Interval
	}

	service := &ResultReportingService{
		repo:             repo,
		config:           config,
		interval:         time.Duration(interval) * time.Second,
		reportStrategies: make(map[string]ReportingStrategy),
		stopChan:         make(chan struct{}),
	}

	// Initialize services
	confluenceService := NewConfluenceService(config)
	mattermostService := NewMattermostService(config)

	// Register reporting strategies based on config
	for _, reportType := range config.Reporting.ReportTypes {
		switch reportType {
		case "confluence":
			service.reportStrategies["confluence"] = &ConfluenceReporter{
				config:            config,
				confluenceService: confluenceService,
			}
		case "mattermost":
			service.reportStrategies["mattermost"] = &MattermostReporter{
				config:            config,
				mattermostService: mattermostService,
			}
		default:
			log.Printf("[ReportingService] Unknown report type: %s", reportType)
		}
	}

	return service
}

// Start begins the periodic report generation
func (s *ResultReportingService) Start() {
	s.runningMutex.Lock()
	if s.isRunning {
		s.runningMutex.Unlock()
		return
	}
	s.isRunning = true
	s.runningMutex.Unlock()

	log.Printf("[ReportingService] Starting result reporting service, interval: %v", s.interval)

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.generateReports()
			case <-s.stopChan:
				log.Println("[ReportingService] Stopping result reporting service")
				return
			}
		}
	}()
}

// Stop halts the reporting service
func (s *ResultReportingService) Stop() {
	s.runningMutex.Lock()
	defer s.runningMutex.Unlock()

	if !s.isRunning {
		return
	}

	close(s.stopChan)
	s.isRunning = false
	log.Println("[ReportingService] Result reporting service stopped")
}

// generateReports collects completed tasks and generates reports
func (s *ResultReportingService) generateReports() {
	// Get completed tasks
	doneTasks := s.repo.GetTasksByStatus(models.StatusDone)

	// If no completed tasks, nothing to report
	if len(doneTasks) == 0 {
		log.Println("[ReportingService] No completed tasks to report")
		return
	}

	log.Printf("[ReportingService] Generating reports for %d tasks", len(doneTasks))

	// Generate and publish reports using configured strategies
	for name, strategy := range s.reportStrategies {
		log.Printf("[ReportingService] Generating %s report", name)

		reportData, err := strategy.GenerateReport(doneTasks)
		if err != nil {
			log.Printf("[ReportingService] Error generating %s report: %v", name, err)
			continue
		}

		err = strategy.PublishReport(reportData)
		if err != nil {
			log.Printf("[ReportingService] Error publishing %s report: %v", name, err)
		}
	}
}

// GenerateReport immediately generates a report (on-demand)
func (s *ResultReportingService) GenerateReport(reportType string) (string, error) {
	strategy, exists := s.reportStrategies[reportType]
	if !exists {
		return "", fmt.Errorf("unknown report type: %s", reportType)
	}

	doneTasks := s.repo.GetTasksByStatus(models.StatusDone)
	return strategy.GenerateReport(doneTasks)
}

// =============== Confluence Reporter Implementation =============== //

// GenerateReport creates a report for Confluence
func (r *ConfluenceReporter) GenerateReport(tasks []*models.Task) (string, error) {
	// Create a new Confluence service if it doesn't exist
	if r.confluenceService == nil {
		r.confluenceService = NewConfluenceService(r.config)
	}

	// This would format task results into a Confluence-appropriate format
	// In a real implementation, this might use a template and the Confluence storage format

	// Create headers
	headers := []string{"Task ID", "Name", "Type", "Status", "Execution Time", "Results"}

	// Create rows
	rows := make([][]string, 0, len(tasks))
	for _, task := range tasks {
		executionTime := ""
		if task.ExecutionResult != nil {
			if time, ok := task.ExecutionResult["execution_time"].(string); ok {
				executionTime = time
			}
		}

		row := []string{
			task.ID,
			task.Name,
			string(task.TaskType),
			string(task.Status),
			executionTime,
			fmt.Sprintf("%d result(s)", len(task.ExecutionResult)),
		}
		rows = append(rows, row)
	}

	// Create table in Confluence format
	table := r.confluenceService.CreateTable(headers, rows)

	// Create full report
	report := fmt.Sprintf("h1. Task Execution Report\n\n")
	report += fmt.Sprintf("_Generated at: %s_\n\n", time.Now().Format(time.RFC3339))
	report += table

	return report, nil
}

// PublishReport uploads the report to Confluence
func (r *ConfluenceReporter) PublishReport(reportData string) error {
	// Create a new Confluence service if it doesn't exist
	if r.confluenceService == nil {
		r.confluenceService = NewConfluenceService(r.config)
	}

	// In a real implementation, this would:
	// 1. Connect to the Confluence API
	// 2. Update the page specified in the config

	pageID := r.config.Confluence.ResultsPage
	if pageID == "" {
		// Fallback to the page ID in reporting config
		pageID = r.config.Reporting.Confluence.PageID
	}

	log.Printf("[ConfluenceReporter] Publishing to Confluence page ID: %s", pageID)

	// Update the page
	err := r.confluenceService.UpdatePage(
		pageID,
		"Task Execution Report",
		reportData,
	)

	return err
}

// =============== Mattermost Reporter Implementation =============== //

// GenerateReport creates a report for Mattermost
func (r *MattermostReporter) GenerateReport(tasks []*models.Task) (string, error) {
	// Create a new Mattermost service if it doesn't exist
	if r.mattermostService == nil {
		r.mattermostService = NewMattermostService(r.config)
	}

	// Format for Mattermost markdown
	report := "### Task Execution Summary\n\n"
	report += fmt.Sprintf("*Generated at:* %s\n\n", time.Now().Format(time.RFC3339))

	// Group tasks by status
	statusCounts := make(map[models.TaskStatus]int)
	for _, task := range tasks {
		statusCounts[task.Status]++
	}

	report += "**Status Summary:**\n\n"
	for status, count := range statusCounts {
		report += fmt.Sprintf("- %s: %d tasks\n", status, count)
	}

	report += "\n**Recent Completed Tasks:**\n\n"

	// Show most recent 5 tasks
	maxDisplay := 5
	if len(tasks) < maxDisplay {
		maxDisplay = len(tasks)
	}

	// Sort tasks by completion time (we'd need to implement this sorting)
	// For now, just take the first few
	for i := 0; i < maxDisplay; i++ {
		task := tasks[i]
		report += fmt.Sprintf("- **%s** (%s) - %s\n", task.Name, task.ID, task.Status)
	}

	return report, nil
}

// PublishReport sends the report to Mattermost
func (r *MattermostReporter) PublishReport(reportData string) error {
	// Create a new Mattermost service if it doesn't exist
	if r.mattermostService == nil {
		r.mattermostService = NewMattermostService(r.config)
	}

	// Connect to Mattermost
	err := r.mattermostService.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to Mattermost: %v", err)
	}

	// Send the report
	err = r.mattermostService.SendTaskReport(reportData)
	if err != nil {
		return fmt.Errorf("failed to send report to Mattermost: %v", err)
	}

	// Disconnect
	_ = r.mattermostService.Disconnect()

	return nil
}
