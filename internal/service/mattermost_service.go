package service

import (
	"fmt"
	"log"
	"my-scheduler-go/internal/config"
	"my-scheduler-go/internal/mattermost"
)

// MattermostService handles integration with Mattermost API
type MattermostService struct {
	config     *config.AppConfig
	connection *mattermost.Connection
}

// NewMattermostService creates a new Mattermost service
func NewMattermostService(cfg *config.AppConfig) *MattermostService {
	return &MattermostService{
		config:     cfg,
		connection: mattermost.NewConnection(cfg.Mattermost.ServerURL, cfg.Mattermost.Token),
	}
}

// Connect establishes a connection to Mattermost
func (s *MattermostService) Connect() error {
	log.Printf("[MattermostService] Connecting to %s", s.config.Mattermost.ServerURL)
	return s.connection.Connect()
}

// Disconnect closes the Mattermost connection
func (s *MattermostService) Disconnect() error {
	log.Println("[MattermostService] Disconnecting")
	return s.connection.Close()
}

// SendMessage sends a message to a Mattermost channel
func (s *MattermostService) SendMessage(channelID, message string) error {
	log.Printf("[MattermostService] Sending message to channel %s", channelID)

	// In a real implementation, this would:
	// 1. Ensure connection is established
	// 2. Format and send the message via the Mattermost API
	// 3. Handle any errors

	// Just simulate success for now
	log.Printf("[MattermostService] Message sent successfully (simulation)")
	return nil
}

// SendTaskReport sends a task report to the configured channel
func (s *MattermostService) SendTaskReport(report string) error {
	channelID := s.config.Mattermost.ChannelID
	if channelID == "" {
		return fmt.Errorf("mattermost channel ID not configured")
	}

	// Add a standard header
	message := "## Task Execution Report\n\n" + report

	return s.SendMessage(channelID, message)
}
