package scheduler

import (
	"log"
	"sync"
	"time"
)

// Configuration 定义配置接口
type Configuration interface{}

// ConfigurationService 管理配置的获取和更新
type ConfigurationService struct {
	fetcher        ConfigurationFetcher
	configurations []Configuration
	mu             sync.RWMutex
	updateInterval time.Duration
	isRunning      bool
	stopChan       chan struct{}
	lastUpdateTime time.Time
}

// ConfigurationFetcher 定义配置获取接口
type ConfigurationFetcher interface {
	// FetchConfigurations 从数据源获取配置
	FetchConfigurations() ([]Configuration, error)
}

// NewConfigurationService 创建配置服务
func NewConfigurationService(fetcher ConfigurationFetcher, updateInterval time.Duration) *ConfigurationService {
	return &ConfigurationService{
		fetcher:        fetcher,
		configurations: make([]Configuration, 0),
		updateInterval: updateInterval,
		isRunning:      false,
		stopChan:       make(chan struct{}),
	}
}

// Start 启动配置服务
func (s *ConfigurationService) Start() {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = true
	s.mu.Unlock()

	log.Println("[ConfigurationService] Starting configuration service")

	// 立即获取一次配置
	s.updateConfigurations()

	// 启动定时更新
	go s.runUpdateLoop()
}

// Stop 停止配置服务
func (s *ConfigurationService) Stop() {
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = false
	close(s.stopChan)
	s.mu.Unlock()

	log.Println("[ConfigurationService] Stopping configuration service")
}

// GetCurrentConfigurations 获取当前配置
func (s *ConfigurationService) GetCurrentConfigurations() []Configuration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 返回配置副本，避免竞争条件
	result := make([]Configuration, len(s.configurations))
	copy(result, s.configurations)

	return result
}

// GetLastUpdateTime 获取上次更新时间
func (s *ConfigurationService) GetLastUpdateTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastUpdateTime
}

// 更新配置
func (s *ConfigurationService) updateConfigurations() {
	configs, err := s.fetcher.FetchConfigurations()
	if err != nil {
		log.Printf("[ConfigurationService] Failed to fetch configurations: %v", err)
		return
	}

	log.Printf("[ConfigurationService] Fetched %d configurations", len(configs))

	s.mu.Lock()
	s.configurations = configs
	s.lastUpdateTime = time.Now()
	s.mu.Unlock()
}

// 运行更新循环
func (s *ConfigurationService) runUpdateLoop() {
	ticker := time.NewTicker(s.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.updateConfigurations()
		case <-s.stopChan:
			return
		}
	}
}
