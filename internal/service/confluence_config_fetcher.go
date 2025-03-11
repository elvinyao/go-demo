package service

import (
	"errors"
	"log"
	"my-scheduler-go/internal/config"
	"my-scheduler-go/internal/scheduler"
	"regexp"
	"strings"
)

// ConfluenceConfigFetcher 从Confluence获取配置
type ConfluenceConfigFetcher struct {
	confluenceService *ConfluenceService
	appConfig         *config.AppConfig
	useMockData       bool
}

// NewConfluenceConfigFetcher 创建Confluence配置获取器
func NewConfluenceConfigFetcher(confluenceService *ConfluenceService, appConfig *config.AppConfig, useMockData bool) *ConfluenceConfigFetcher {
	return &ConfluenceConfigFetcher{
		confluenceService: confluenceService,
		appConfig:         appConfig,
		useMockData:       useMockData,
	}
}

// FetchConfigurations 从Confluence获取配置
func (f *ConfluenceConfigFetcher) FetchConfigurations() ([]scheduler.Configuration, error) {
	log.Println("[ConfluenceConfigFetcher] Fetching configurations from Confluence")

	// 如果使用模拟数据，则返回模拟配置
	if f.useMockData {
		return f.getMockConfigurations(), nil
	}

	// 获取Confluence页面
	pageID := f.appConfig.Confluence.MainPageID
	page, err := f.confluenceService.GetPage(pageID)
	if err != nil {
		return nil, err
	}

	// 获取页面内容
	content, ok := page["content"].(string)
	if !ok {
		return nil, errors.New("unable to get page content")
	}

	// 解析表格内容
	return f.parseTableConfigurations(content)
}

// parseTableConfigurations 解析页面内容中的表格
func (f *ConfluenceConfigFetcher) parseTableConfigurations(content string) ([]scheduler.Configuration, error) {
	// 在实际环境中，您需要实现从HTML或Confluence存储格式中解析表格的逻辑
	// 这里简化为解析Confluence表格标记语法

	// 表格匹配正则表达式
	tableRegex := regexp.MustCompile(`(?s)\|\|(.*?)\|\|(.*?)\|\|`)
	matches := tableRegex.FindAllStringSubmatch(content, -1)

	if len(matches) == 0 {
		return nil, errors.New("no table found in content")
	}

	// 解析表头
	headerMatch := matches[0]
	headerCells := strings.Split(headerMatch[1], "||")

	// 解析行
	rows := matches[1:]
	var configs []scheduler.Configuration

	for _, rowMatch := range rows {
		cells := strings.Split(rowMatch[1], "|")
		if len(cells) < 4 {
			log.Println("[ConfluenceConfigFetcher] Skipping row with insufficient cells")
			continue
		}

		// 解析单元格内容
		config := scheduler.MattermostConfig{
			ID:          strings.TrimSpace(cells[0]),
			ChannelID:   strings.TrimSpace(cells[1]),
			MessageType: strings.TrimSpace(cells[2]),
			ForwardType: strings.TrimSpace(cells[3]),
			Custom:      make(map[string]interface{}),
		}

		// 添加额外自定义字段
		for i := 4; i < len(cells) && i < len(headerCells); i++ {
			config.Custom[headerCells[i]] = strings.TrimSpace(cells[i])
		}

		configs = append(configs, config)
	}

	return configs, nil
}

// getMockConfigurations 返回模拟配置数据
func (f *ConfluenceConfigFetcher) getMockConfigurations() []scheduler.Configuration {
	log.Println("[ConfluenceConfigFetcher] Generating mock configurations")

	return []scheduler.Configuration{
		scheduler.MattermostConfig{
			ID:          "config1",
			ChannelID:   "channel1",
			MessageType: string(EventTypePosted),
			ForwardType: "direct_message",
			Custom: map[string]interface{}{
				"target_user_id": "user123",
				"include_files":  "true",
			},
		},
		scheduler.MattermostConfig{
			ID:          "config2",
			ChannelID:   "channel2",
			MessageType: string(EventTypePosted),
			ForwardType: "channel_message",
			Custom: map[string]interface{}{
				"target_channel_id": "channel456",
				"include_files":     "false",
			},
		},
		scheduler.MattermostConfig{
			ID:          "config3",
			ChannelID:   "channel1",
			MessageType: string(EventTypeUserAdded),
			ForwardType: "notification",
			Custom: map[string]interface{}{
				"notify_admin": "true",
			},
		},
	}
}

// Constants for event types
const (
	EventTypePosted    = "posted"
	EventTypeUserAdded = "user_added"
)
