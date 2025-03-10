package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// AppConfig structure for config.yaml
type AppConfig struct {
	// Environment - environment mode (development, production, test)
	Environment string `mapstructure:"environment"`

	// Scheduler configuration
	Scheduler struct {
		PollInterval int  `mapstructure:"poll_interval"`
		Concurrency  int  `mapstructure:"concurrency"`
		Coalesce     bool `mapstructure:"coalesce"`
		MaxInstances int  `mapstructure:"max_instances"`
	} `mapstructure:"scheduler"`

	// Jira configuration
	Jira struct {
		URL      string `mapstructure:"url"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"jira"`

	// Confluence configuration
	Confluence struct {
		URL         string `mapstructure:"url"`
		Username    string `mapstructure:"username"`
		Password    string `mapstructure:"password"`
		MainPageID  string `mapstructure:"main_page_id"`
		ResultsPage string `mapstructure:"task_result_page_id"`
	} `mapstructure:"confluence"`

	// Mattermost configuration
	Mattermost struct {
		ServerURL         string `mapstructure:"server_url"`
		Token             string `mapstructure:"token"`
		ChannelID         string `mapstructure:"channel_id"`
		ReconnectInterval int    `mapstructure:"reconnect_interval"`
	} `mapstructure:"mattermost"`

	// Log configuration
	Log struct {
		Level       string `mapstructure:"level"`
		Filename    string `mapstructure:"filename"`
		MaxBytes    int    `mapstructure:"max_bytes"`
		BackupCount int    `mapstructure:"backup_count"`
		Format      string `mapstructure:"format"`
	} `mapstructure:"log"`

	// Storage configuration
	Storage struct {
		Path string `mapstructure:"path"`
	} `mapstructure:"storage"`
}

// LoadConfig loads configuration from the specified file path
func LoadConfig(path string) (*AppConfig, error) {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Set default values if not specified
	if cfg.Environment == "" {
		cfg.Environment = "development"
	}

	fmt.Println("[config] Loaded config from:", path)
	return &cfg, nil
}
