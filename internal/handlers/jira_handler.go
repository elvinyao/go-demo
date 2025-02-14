package handlers

import "log"

// JiraHandler 示例
type JiraHandler struct {
	baseURL  string
	username string
	token    string
}

func NewJiraHandler(baseURL, username, token string) *JiraHandler {
	return &JiraHandler{baseURL, username, token}
}

func (h *JiraHandler) CreateIssue(projectKey, summary string) error {
	log.Printf("[JiraHandler] Creating issue in project %s, summary: %s", projectKey, summary)
	return nil
}
