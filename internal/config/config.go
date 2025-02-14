package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// AppConfig 结构体示例，用于映射 config.yaml
type AppConfig struct {
	Scheduler struct {
		PollInterval int  `mapstructure:"poll_interval"`
		Concurrency  int  `mapstructure:"concurrency"`
		Coalesce     bool `mapstructure:"coalesce"`
		MaxInstances int  `mapstructure:"max_instances"`
	} `mapstructure:"scheduler"`

	Mattermost struct {
		ServerURL         string `mapstructure:"server_url"`
		Token             string `mapstructure:"token"`
		ChannelID         string `mapstructure:"channel_id"`
		ReconnectInterval int    `mapstructure:"reconnect_interval"`
	} `mapstructure:"mattermost"`

	Log struct {
		Level       string `mapstructure:"level"`
		Filename    string `mapstructure:"filename"`
		MaxBytes    int    `mapstructure:"max_bytes"`
		BackupCount int    `mapstructure:"backup_count"`
		Format      string `mapstructure:"format"`
	} `mapstructure:"log"`
}

// LoadConfig 示例方法
func LoadConfig(path string) (*AppConfig, error) {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	fmt.Println("[config] Loaded config from:", path)
	return &cfg, nil
}
