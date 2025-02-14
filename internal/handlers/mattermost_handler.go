package handlers

import "log"

// MattermostHandler 示例
type MattermostHandler struct {
	baseURL string
	token   string
}

func NewMattermostHandler(baseURL, token string) *MattermostHandler {
	return &MattermostHandler{baseURL, token}
}

func (h *MattermostHandler) SendMessage(channelID, message string) error {
	log.Printf("[MattermostHandler] Sending message to channel %s: %s", channelID, message)
	return nil
}
